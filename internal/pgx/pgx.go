package pgx

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/madsrc/sophrosyne"
)

type UserService struct {
	config       *sophrosyne.Config
	pool         *pgxpool.Pool
	logger       *slog.Logger
	randomSource io.Reader
}

func NewUserService(ctx context.Context, config *sophrosyne.Config, logger *slog.Logger, randomSource io.Reader) (*UserService, error) {
	pgxconfig, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%d/%s", config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name))
	if err != nil {
		return nil, err
	}
	pgxconfig.ConnConfig.Tracer = otelpgx.NewTracer()
	pgxconfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		logger.DebugContext(ctx, "database connection established")
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxconfig)
	if err != nil {
		return nil, err
	}

	ue := &UserService{
		config:       config,
		pool:         pool,
		logger:       logger,
		randomSource: randomSource,
	}

	err = ue.createRootUser(ctx)
	if err != nil {
		return nil, err
	}

	return ue, nil
}

func (s *UserService) getUser(ctx context.Context, column, input any) (sophrosyne.User, error) {
	var rows pgx.Rows
	if column == "email" {
		rows, _ = s.pool.Query(ctx, "SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL LIMIT 1", input)
	} else if column == "name" {
		rows, _ = s.pool.Query(ctx, "SELECT * FROM users WHERE name = $1 AND deleted_at IS NULL LIMIT 1", input)
	} else if column == "id" {
		rows, _ = s.pool.Query(ctx, "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1", input)
	} else if column == "token" {
		rows, _ = s.pool.Query(ctx, "SELECT * FROM users WHERE token = $1 AND deleted_at IS NULL LIMIT 1", input)
	} else {
		return sophrosyne.User{}, sophrosyne.NewUnreachableCodeError()
	}
	user, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[sophrosyne.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sophrosyne.User{}, sophrosyne.ErrNotFound
		}
		return sophrosyne.User{}, err
	}
	return *user, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (sophrosyne.User, error) {
	return s.getUser(ctx, "id", id)
}
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (sophrosyne.User, error) {
	return s.getUser(ctx, "email", email)
}
func (s *UserService) GetUserByName(ctx context.Context, name string) (sophrosyne.User, error) {
	return s.getUser(ctx, "name", name)
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

func (s *UserService) createRootUser(ctx context.Context) error {
	// Begin transaction
	tx, err := s.pool.Begin(ctx)
	defer func() {
		s.logger.DebugContext(ctx, "rolling back transaction")
		tx.Rollback(ctx)
	}()
	if err != nil {
		return err
	}
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
