// Sophrosyne
//
//	Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//
//	it under the terms of the GNU Affero General Public License as published by
//	the Free Software Foundation, either version 3 of the License, or
//	(at your option) any later version.
//
//	This program is distributed in the hope that it will be useful,
//	but WITHOUT ANY WARRANTY; without even the implied warranty of
//	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//	GNU Affero General Public License for more details.
//
//	You should have received a copy of the GNU Affero General Public License
//	along with this program.  If not, see <http://www.gnu.org/licenses/>.
package pgx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/madsrc/sophrosyne"
)

type ProfileService struct {
	config       *sophrosyne.Config
	pool         *pgxpool.Pool
	logger       *slog.Logger
	checkService sophrosyne.CheckService
}

func NewProfileService(ctx context.Context, config *sophrosyne.Config, logger *slog.Logger, checkService sophrosyne.CheckService) (*ProfileService, error) {
	pool, err := newPool(ctx, config, logger)
	if err != nil {
		return nil, err
	}
	ps := &ProfileService{
		config:       config,
		pool:         pool,
		logger:       logger,
		checkService: checkService,
	}

	err = ps.createDefaultProfile(ctx)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (p *ProfileService) nameToID(ctx context.Context, name string) (string, error) {
	row := p.pool.QueryRow(ctx, `SELECT id FROM profiles WHERE name = $1 LIMIT 1`, name)
	var id string
	err := row.Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (p *ProfileService) GetProfile(ctx context.Context, id string) (sophrosyne.Profile, error) {
	type dbret struct {
		ID        string     `db:"id"`
		Name      string     `db:"name"`
		CreatedAt time.Time  `db:"created_at"`
		UpdatedAt time.Time  `db:"updated_at"`
		DeletedAt *time.Time `db:"deleted_at"`
		Checks    []string   `db:"checks"`
	}
	p.logger.DebugContext(ctx, "GetProfile", "id", id)
	var rows pgx.Rows
	rows, _ = p.pool.Query(ctx, `SELECT p.*,
       CASE WHEN array_agg(c.name) IS NOT NULL
            THEN array_remove(array_agg(c.name), NULL)
            ELSE '{}'::text[]
       END AS checks
FROM profiles p
LEFT JOIN profiles_checks pc ON p.id = pc.profile_id
LEFT JOIN checks c ON pc.check_id = c.id AND c.deleted_at IS NULL
WHERE p.id = $1 AND p.deleted_at IS NULL
GROUP BY p.id, p.name
LIMIT 1;`, id)
	profile, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbret])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sophrosyne.Profile{}, sophrosyne.ErrNotFound
		}
		return sophrosyne.Profile{}, err
	}

	ret := sophrosyne.Profile{
		ID:        profile.ID,
		Name:      profile.Name,
		CreatedAt: profile.CreatedAt,
		UpdatedAt: profile.UpdatedAt,
		DeletedAt: profile.DeletedAt,
		Checks:    make([]sophrosyne.Check, 0, len(profile.Checks)),
	}
	for _, check := range profile.Checks {
		c, err := p.checkService.GetCheckByName(ctx, check)
		if err != nil {
			return sophrosyne.Profile{}, err
		}
		ret.Checks = append(ret.Checks, c)
	}
	return ret, nil
}

func (p *ProfileService) GetProfileByName(ctx context.Context, name string) (sophrosyne.Profile, error) {
	id, err := p.nameToID(ctx, name)
	if err != nil {
		return sophrosyne.Profile{}, err
	}
	return p.GetProfile(ctx, id)
}

func (p *ProfileService) GetProfiles(ctx context.Context, cursor *sophrosyne.DatabaseCursor) ([]sophrosyne.Profile, error) {
	if cursor == nil {
		cursor = &sophrosyne.DatabaseCursor{}
	}
	p.logger.DebugContext(ctx, "getting profiles", "cursor", cursor)
	rows, err := p.pool.Query(ctx, `SELECT * FROM profiles WHERE id > $1 AND deleted_at IS NULL ORDER BY id ASC LIMIT $2`, cursor.Position, p.config.Services.Profiles.PageSize+1)
	profiles, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[sophrosyne.Profile])
	if err != nil {
		return []sophrosyne.Profile{}, err
	}
	if len(profiles) == 0 {
		cursor.Reset()
	} else if len(profiles) <= p.config.Services.Profiles.PageSize && len(profiles) > 0 {
		cursor.Reset()
	} else if len(profiles) > p.config.Services.Profiles.PageSize {
		cursor.Advance(profiles[len(profiles)-2].ID)
		profiles = profiles[:len(profiles)-1]
	}

	return profiles, nil
}

func (p *ProfileService) CreateProfile(ctx context.Context, profile sophrosyne.CreateProfileRequest) (sophrosyne.Profile, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return sophrosyne.Profile{}, err
	}
	defer tx.Rollback(ctx)

	rows, _ := tx.Query(ctx, `INSERT INTO profiles (name) VALUES ($1) RETURNING *`, profile.Name)
	retP, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[sophrosyne.Profile])
	if err != nil {
		return sophrosyne.Profile{}, err
	}

	if len(profile.Checks) > 0 {
		// translate the list of check names into check ID's.
		rows, _ := tx.Query(ctx, `SELECT id from checks WHERE name IN $1 AND deleted_at IS NULL`, profile.Checks)
		checkIDs, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[sophrosyne.Check])
		if err != nil {
			return sophrosyne.Profile{}, err
		}
		if len(checkIDs) != len(profile.Checks) {
			return sophrosyne.Profile{}, fmt.Errorf("checks mismatch")
		}

		// Insert into the profiles_checks table
		_, err = tx.Exec(ctx, `INSERT INTO profiles_checks (profile_id, check_id)
SELECT $1, unnest($2);`, retP.ID, checkIDs)
		if err != nil {
			return sophrosyne.Profile{}, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return sophrosyne.Profile{}, err
	}

	return *retP, nil

}

func (p *ProfileService) UpdateProfile(ctx context.Context, profile sophrosyne.UpdateProfileRequest) (sophrosyne.Profile, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return sophrosyne.Profile{}, err
	}
	defer tx.Rollback(ctx)

	rows, _ := tx.Query(ctx, `SELECT id FROM profiles WHERE name = $1 AND deleted_at IS NULL`, profile.Name)
	pp, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[sophrosyne.Profile])
	if err != nil {
		return sophrosyne.Profile{}, err
	}

	_, err = tx.Exec(ctx, `DELETE FROM profiles_checks
WHERE profile_id = $1 AND check_id NOT IN (SELECT unnest($2));`, pp.ID, profile.Checks)
	if err != nil {
		return sophrosyne.Profile{}, err
	}

	_, err = tx.Exec(ctx, `INSERT INTO profiles_checks (profile_id, check_id)
SELECT $1, c.check_id
FROM unnest($2) AS c(check_id)
ON CONFLICT (profile_id, check_id) DO NOTHING;`, pp.ID, profile.Checks)
	if err != nil {
		return sophrosyne.Profile{}, err
	}

	rows, _ = tx.Query(ctx, `SELECT c.*
FROM checks c
JOIN profiles_checks pc ON c.id = pc.check_id
JOIN profiles p ON pc.profile_id = p.id
WHERE p.id = $1
AND c.name = ANY($2);`, pp.ID, profile.Checks)
	checks, err := pgx.CollectRows(rows, pgx.RowToStructByName[sophrosyne.Check])
	if err != nil {
		return sophrosyne.Profile{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return sophrosyne.Profile{}, err
	}

	return sophrosyne.Profile{
		ID:     pp.ID,
		Name:   profile.Name,
		Checks: checks,
	}, nil
}

func (p *ProfileService) DeleteProfile(ctx context.Context, name string) error {
	cmdTag, err := p.pool.Exec(ctx, `UPDATE profiles SET deleted_at = NOW() WHERE name = $1 AND deleted_at IS NULL`, name)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return sophrosyne.ErrNotFound
	}
	return nil
}

func (p *ProfileService) createDefaultProfile(ctx context.Context) error {
	p.logger.DebugContext(ctx, "creating default profile")
	defaultProfile := sophrosyne.CreateProfileRequest{
		Name: "default",
	}
	// Check if root user exists and exit early if it does
	var exists bool
	err := p.pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM profiles WHERE name = $1)", "default").Scan(&exists)
	if err != nil {
		return err
	}
	p.logger.DebugContext(ctx, "default profile existence", "exists", exists)
	if exists {
		return nil
	}

	_, err = p.CreateProfile(ctx, defaultProfile)
	return err
}
