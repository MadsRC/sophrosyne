// Sophrosyne
//   Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU Affero General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU Affero General Public License for more details.
//
//   You should have received a copy of the GNU Affero General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.

package pgx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/madsrc/sophrosyne"
)

type checkDbEntry struct {
	ID               string     `db:"id"`
	Name             string     `db:"name"`
	UpstreamServices []string   `db:"upstream_services"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	DeletedAt        *time.Time `db:"deleted_at"`
	Profiles         []string   `db:"profiles"`
}

type CheckService struct {
	config *sophrosyne.Config
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewCheckService(ctx context.Context, config *sophrosyne.Config, logger *slog.Logger) (*CheckService, error) {
	pool, err := newPool(ctx, config, logger)
	if err != nil {
		return nil, err
	}
	ps := &CheckService{
		config: config,
		pool:   pool,
		logger: logger,
	}

	return ps, nil
}

func (p *CheckService) nameToID(ctx context.Context, name string) (string, error) {
	row := p.pool.QueryRow(ctx, `SELECT id FROM checks WHERE name = $1 LIMIT 1`, name)
	var id string
	err := row.Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (p *CheckService) GetCheck(ctx context.Context, id string) (sophrosyne.Check, error) {
	p.logger.DebugContext(ctx, "GetCheck", "id", id)
	var rows pgx.Rows
	rows, _ = p.pool.Query(ctx, `SELECT p.*,
       CASE WHEN array_agg(c.name) IS NOT NULL
            THEN array_remove(array_agg(c.name), NULL)
            ELSE '{}'::text[]
       END AS profiles
FROM checks p
LEFT JOIN profiles_checks pc ON p.id = pc.check_id
LEFT JOIN profiles c ON pc.profile_id = c.id AND c.deleted_at IS NULL
WHERE p.id = $1 AND p.deleted_at IS NULL
GROUP BY p.id, p.name
LIMIT 1;`, id)
	check, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[checkDbEntry])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sophrosyne.Check{}, sophrosyne.ErrNotFound
		}
		return sophrosyne.Check{}, err
	}

	var uss []url.URL
	for _, entry := range check.UpstreamServices {
		us, err := url.Parse(entry)
		if err != nil {
			p.logger.ErrorContext(ctx, "unable to parse upstream service", "entry", entry, "error", err)
			return sophrosyne.Check{}, err
		}
		uss = append(uss, *us)
	}

	ret := sophrosyne.Check{
		ID:               check.ID,
		Name:             check.Name,
		UpstreamServices: uss,
		CreatedAt:        check.CreatedAt,
		UpdatedAt:        check.UpdatedAt,
		DeletedAt:        check.DeletedAt,
		Profiles:         make([]sophrosyne.Profile, 0, len(check.Profiles)),
	}
	for _, check := range check.Profiles {
		ret.Profiles = append(ret.Profiles, sophrosyne.Profile{
			Name: check,
		})
	}
	return ret, nil
}

func (p *CheckService) GetCheckByName(ctx context.Context, name string) (sophrosyne.Check, error) {
	id, err := p.nameToID(ctx, name)
	if err != nil {
		return sophrosyne.Check{}, err
	}
	return p.GetCheck(ctx, id)
}

func (p *CheckService) GetChecks(ctx context.Context, cursor *sophrosyne.DatabaseCursor) ([]sophrosyne.Check, error) {
	if cursor == nil {
		cursor = &sophrosyne.DatabaseCursor{}
	}
	p.logger.DebugContext(ctx, "getting checks", "cursor", cursor)
	rows, err := p.pool.Query(ctx, `SELECT * FROM checks WHERE id > $1 AND deleted_at IS NULL ORDER BY id ASC LIMIT $2`, cursor.Position, p.config.Services.Checks.PageSize+1)
	checks, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[sophrosyne.Check])
	if err != nil {
		return []sophrosyne.Check{}, err
	}
	if len(checks) == 0 {
		cursor.Reset()
	} else if len(checks) <= p.config.Services.Profiles.PageSize && len(checks) > 0 {
		cursor.Reset()
	} else if len(checks) > p.config.Services.Profiles.PageSize {
		cursor.Advance(checks[len(checks)-2].ID)
		checks = checks[:len(checks)-1]
	}

	return checks, nil
}

func (p *CheckService) CreateCheck(ctx context.Context, check sophrosyne.CreateCheckRequest) (sophrosyne.Check, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return sophrosyne.Check{}, err
	}
	defer tx.Rollback(ctx)

	rows, _ := tx.Query(ctx, `INSERT INTO checks (name, upstream_services) VALUES ($1, $2) RETURNING *`, check.Name, check.UpstreamServices)
	retP, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[checkDbEntry])
	if err != nil {
		return sophrosyne.Check{}, err
	}

	var uss []url.URL
	for _, entry := range check.UpstreamServices {
		us, err := url.Parse(entry)
		if err != nil {
			p.logger.ErrorContext(ctx, "unable to parse upstream service", "entry", entry, "error", err)
			return sophrosyne.Check{}, err
		}
		uss = append(uss, *us)
	}

	p.logger.DebugContext(ctx, "checking profiles", "profiles", check.Profiles, "count", len(check.Profiles))
	if len(check.Profiles) > 0 {
		// translate the list of profile names into check ID's.
		rows, _ := tx.Query(ctx, `SELECT id from profiles WHERE name = ANY($1) AND deleted_at IS NULL`, check.Profiles)
		profileIDs, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[sophrosyne.Profile])
		if err != nil {
			return sophrosyne.Check{}, err
		}
		p.logger.DebugContext(ctx, "profiles", "profiles", profileIDs)
		if len(profileIDs) != len(check.Profiles) {
			return sophrosyne.Check{}, fmt.Errorf("profiles mismatch")
		}

		var ids []string
		for _, profileID := range profileIDs {
			ids = append(ids, profileID.ID)
		}

		// Insert into the profiles_checks table
		_, err = tx.Exec(ctx, `INSERT INTO profiles_checks (check_id, profile_id)
SELECT $1, unnest($2::TEXT[])`, retP.ID, ids)
		if err != nil {
			return sophrosyne.Check{}, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return sophrosyne.Check{}, err
	}

	return sophrosyne.Check{
		ID:               retP.ID,
		Name:             retP.Name,
		Profiles:         make([]sophrosyne.Profile, 0, len(check.Profiles)),
		UpstreamServices: uss,
		CreatedAt:        retP.CreatedAt,
		UpdatedAt:        retP.UpdatedAt,
		DeletedAt:        retP.DeletedAt,
	}, nil

}

func (p *CheckService) UpdateCheck(ctx context.Context, check sophrosyne.UpdateCheckRequest) (sophrosyne.Check, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return sophrosyne.Check{}, err
	}
	defer tx.Rollback(ctx)

	rows, _ := tx.Query(ctx, `SELECT id FROM checks WHERE name = $1 AND deleted_at IS NULL`, check.Name)
	pp, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[sophrosyne.Check])
	if err != nil {
		return sophrosyne.Check{}, err
	}

	_, err = tx.Exec(ctx, `DELETE FROM profiles_checks
WHERE check_id = $1 AND profile_id NOT IN (SELECT unnest($2));`, pp.ID, check.Profiles)
	if err != nil {
		return sophrosyne.Check{}, err
	}

	_, err = tx.Exec(ctx, `INSERT INTO profiles_checks (profile_id, check_id)
SELECT $1, c.profile_id
FROM unnest($2) AS c(profile_id)
ON CONFLICT (profile_id, check_id) DO NOTHING;`, pp.ID, check.Profiles)
	if err != nil {
		return sophrosyne.Check{}, err
	}

	rows, _ = tx.Query(ctx, `SELECT c.*
FROM profiles c
JOIN profiles_checks pc ON c.id = pc.profile_id
JOIN checks p ON pc.check_id = p.id
WHERE p.id = $1
AND c.name = ANY($2);`, pp.ID, check.Profiles)
	profiles, err := pgx.CollectRows(rows, pgx.RowToStructByName[sophrosyne.Profile])
	if err != nil {
		return sophrosyne.Check{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return sophrosyne.Check{}, err
	}

	return sophrosyne.Check{
		ID:       pp.ID,
		Name:     check.Name,
		Profiles: profiles,
	}, nil
}

func (p *CheckService) DeleteCheck(ctx context.Context, name string) error {
	cmdTag, err := p.pool.Exec(ctx, `UPDATE checks SET deleted_at = NOW() WHERE name = $1 AND deleted_at IS NULL`, name)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return sophrosyne.ErrNotFound
	}
	return nil
}
