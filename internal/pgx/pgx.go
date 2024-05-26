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
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/madsrc/sophrosyne"
)

const (
	// Name of the default profile in the database
	DefaultProfileName = "default"
)

type conn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// afterConnect is a pgx.ConnConfig.AfterConnect function that logs 'database connection established', at the debug level.
//
// If logger is nil, this function panics.
func afterConnect(logger *slog.Logger) func(ctx context.Context, conn *pgx.Conn) error {
	return func(ctx context.Context, conn *pgx.Conn) error {
		logger.DebugContext(ctx, "database connection established")
		return nil
	}
}

func newPool(ctx context.Context, config *sophrosyne.Config, logger *slog.Logger) (*pgxpool.Pool, error) {
	pgxconfig, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	))
	if err != nil {
		return nil, err
	}
	pgxconfig.ConnConfig.Tracer = otelpgx.NewTracer()
	pgxconfig.AfterConnect = afterConnect(logger)
	return pgxpool.NewWithConfig(ctx, pgxconfig)
}

type UserService struct {
	config         *sophrosyne.Config
	pool           conn
	logger         *slog.Logger
	randomSource   io.Reader
	profileService sophrosyne.ProfileService
}

func NewUserService(ctx context.Context, config *sophrosyne.Config, logger *slog.Logger, randomSource io.Reader, profileService sophrosyne.ProfileService, pool conn) (*UserService, error) {
	var err error
	if pool == nil {
		pool, err = newPool(ctx, config, logger)
		if err != nil {
			return nil, err
		}
	}

	ue := &UserService{
		config:         config,
		pool:           pool,
		logger:         logger,
		randomSource:   randomSource,
		profileService: profileService,
	}

	err = ue.createRootUser(ctx)
	if err != nil {
		return nil, err
	}

	return ue, nil
}

var getUserQueryMap = map[string]string{
	"email": "SELECT id, name, email, token, is_admin, default_profile, created_at, updated_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS NULL LIMIT 1",
	"name":  "SELECT id, name, email, token, is_admin, default_profile, created_at, updated_at, deleted_at FROM users WHERE name = $1 AND deleted_at IS NULL LIMIT 1",
	"id":    "SELECT id, name, email, token, is_admin, default_profile, created_at, updated_at, deleted_at FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1",
	"token": "SELECT id, name, email, token, is_admin, default_profile, created_at, updated_at, deleted_at FROM users WHERE token = $1 AND deleted_at IS NULL LIMIT 1",
}

type getUserDbReturn struct {
	ID             string      `db:"id"`
	Name           string      `db:"name"`
	Email          string      `db:"email"`
	Token          []byte      `db:"token"`
	IsAdmin        bool        `db:"is_admin"`
	DefaultProfile pgtype.Text `db:"default_profile"`
	CreatedAt      time.Time   `db:"created_at"`
	UpdatedAt      time.Time   `db:"updated_at"`
	DeletedAt      *time.Time  `db:"deleted_at"`
}

func (g getUserDbReturn) ToUser(ctx context.Context, profileService sophrosyne.ProfileService) (sophrosyne.User, error) {
	user := sophrosyne.User{
		ID:        g.ID,
		Name:      g.Name,
		Email:     g.Email,
		Token:     g.Token,
		IsAdmin:   g.IsAdmin,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
		DeletedAt: g.DeletedAt,
	}

	if g.DefaultProfile.String == "" {
		prof, err := profileService.GetProfileByName(ctx, DefaultProfileName)
		if err != nil {
			return sophrosyne.User{}, err
		}
		user.DefaultProfile = prof
	} else {
		prof, err := profileService.GetProfile(ctx, g.DefaultProfile.String)
		if err != nil {
			return sophrosyne.User{}, err
		}
		user.DefaultProfile = prof
	}

	return user, nil
}

func (s *UserService) getUser(ctx context.Context, column string, input []byte) (sophrosyne.User, error) {
	query, ok := getUserQueryMap[column]
	if !ok {
		return sophrosyne.User{}, sophrosyne.NewUnreachableCodeError()
	}
	rows, _ := s.pool.Query(ctx, query, input)
	user, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[getUserDbReturn])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sophrosyne.User{}, sophrosyne.ErrNotFound
		}
		return sophrosyne.User{}, err
	}

	ret, err := user.ToUser(ctx, s.profileService)
	if err != nil {
		return sophrosyne.User{}, err
	}

	return ret, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (sophrosyne.User, error) {
	return s.getUser(ctx, "id", []byte(id))
}
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (sophrosyne.User, error) {
	return s.getUser(ctx, "email", []byte(email))
}
func (s *UserService) GetUserByName(ctx context.Context, name string) (sophrosyne.User, error) {
	return s.getUser(ctx, "name", []byte(name))
}
func (s *UserService) GetUserByToken(ctx context.Context, token []byte) (sophrosyne.User, error) {
	return s.getUser(ctx, "token", token)
}
func (s *UserService) GetUsers(ctx context.Context, cursor *sophrosyne.DatabaseCursor) ([]sophrosyne.User, error) {
	if cursor == nil {
		cursor = &sophrosyne.DatabaseCursor{}
	}
	s.logger.DebugContext(ctx, "getting users", "cursor", cursor)
	rows, _ := s.pool.Query(ctx, "SELECT * FROM users WHERE id > $1 AND deleted_at IS NULL ORDER BY id ASC LIMIT $2", cursor.Position, s.config.Services.Users.PageSize+1)
	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[sophrosyne.User])
	if err != nil {
		return []sophrosyne.User{}, err
	}
	// Advance the cursor
	if len(users) == 0 {
		cursor.Reset() // No users were read, so reset the cursor
	} else if len(users) <= s.config.Services.Users.PageSize && len(users) > 0 {
		cursor.Reset() // We read all the users, so reset the cursor
	} else if len(users) > s.config.Services.Users.PageSize {
		cursor.Advance(users[len(users)-2].ID) // We read one extra user, so set the cursor to the second-to-last user
		users = users[:len(users)-1]           // Remove the last user
	}
	return users, nil
}
func (s *UserService) CreateUser(ctx context.Context, user sophrosyne.CreateUserRequest) (sophrosyne.User, error) {
	token, err := sophrosyne.NewToken(s.randomSource)
	if err != nil {
		return sophrosyne.User{}, err
	}
	tokenHash := sophrosyne.ProtectToken(token, s.config)

	rows, _ := s.pool.Query(ctx, "INSERT INTO users (name, email, token, is_admin) VALUES ($1, $2, $3, $4) RETURNING *", user.Name, user.Email, tokenHash, user.IsAdmin)
	newUser, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[sophrosyne.User])
	if err != nil {
		s.logger.DebugContext(ctx, "database returned error", "error", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return sophrosyne.User{}, sophrosyne.NewConstraintViolationError(pgErr, pgErr.Code, pgErr.Detail, pgErr.TableName, pgErr.ConstraintName)
			}
		}
		return sophrosyne.User{}, err
	}
	newUser.Token = token // ensure returned token is the raw token, not the hashed token
	return *newUser, nil
}
func (s *UserService) UpdateUser(ctx context.Context, user sophrosyne.UpdateUserRequest) (sophrosyne.User, error) {
	rows, _ := s.pool.Query(ctx, "UPDATE users SET email = $1, is_admin = $2 WHERE name = $3 AND deleted_at IS NULL RETURNING *", user.Email, user.IsAdmin, user.Name)
	updatedUser, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[sophrosyne.User])
	if err != nil {
		s.logger.DebugContext(ctx, "database returned error", "error", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return sophrosyne.User{}, sophrosyne.NewConstraintViolationError(pgErr, pgErr.Code, pgErr.Detail, pgErr.TableName, pgErr.ConstraintName)
			}
		}
		return sophrosyne.User{}, err
	}
	return *updatedUser, nil
}
func (s *UserService) DeleteUser(ctx context.Context, name string) error {
	cmdTag, err := s.pool.Exec(ctx, "UPDATE users SET deleted_at = NOW() WHERE name = $1 AND deleted_at IS NULL", name)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return sophrosyne.ErrNotFound
	}
	return nil
}

func (s *UserService) RotateToken(ctx context.Context, name string) ([]byte, error) {
	token, err := sophrosyne.NewToken(s.randomSource)
	if err != nil {
		return nil, err
	}
	tokenHash := sophrosyne.ProtectToken(token, s.config)

	cmdTag, err := s.pool.Exec(ctx, "UPDATE users SET token = $1 WHERE name = $2 AND deleted_at IS NULL", tokenHash, name)
	if err != nil {
		return nil, err
	}
	if cmdTag.RowsAffected() == 0 {
		return nil, sophrosyne.ErrNotFound
	}
	return token, nil
}

func (s *UserService) Health(ctx context.Context) (bool, []byte) {
	_, err := s.pool.Exec(ctx, "SELECT 1")
	if err != nil {
		s.logger.DebugContext(ctx, "healthcheck database error", "error", err)
		return false, []byte(`{"users":{"healthy":false}}`)
	}
	return true, []byte(`{"users":{"healthy":true}}`)
}

func (s *UserService) createRootUser(ctx context.Context) error {
	// Begin transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		s.logger.DebugContext(ctx, "rolling back transaction")
		_ = tx.Rollback(ctx)
	}()
	// Check if root user exists and exit early if it does
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE name = $1 AND email = $2 AND is_admin = true)", s.config.Principals.Root.Name, s.config.Principals.Root.Email).Scan(&exists)
	if err != nil {
		return err
	}
	s.logger.DebugContext(ctx, "root user existence", "exists", exists)
	if exists {
		if !s.config.Principals.Root.Recreate {
			s.logger.DebugContext(ctx, "root user exists and recreate is false")
			return nil
		}
	}
	var token []byte
	if s.config.Development.StaticRootToken == "" {
		token, err = sophrosyne.NewToken(s.randomSource)
	} else {
		token = []byte(s.config.Development.StaticRootToken)
	}

	if err != nil {
		return err
	}
	s.logger.InfoContext(ctx, "root token", "token", base64.StdEncoding.EncodeToString(token))
	tokenHash := sophrosyne.ProtectToken(token, s.config)
	_, err = tx.Exec(ctx, "INSERT INTO users (name, email, token, is_admin) VALUES ($1, $2, $3, true) ON CONFLICT (name) DO UPDATE SET email = $2, token = $3, is_admin = true", s.config.Principals.Root.Name, s.config.Principals.Root.Email, tokenHash)
	if err != nil {
		return err
	}
	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}
