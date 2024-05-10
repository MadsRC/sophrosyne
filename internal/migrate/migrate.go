package migrate

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/madsrc/sophrosyne"
)

var ErrNoChange = migrate.ErrNoChange

//go:embed migrations
var fs embed.FS

type MigrationService struct {
	migrate *migrate.Migrate
}

func NewMigrationService(config *sophrosyne.Config) (*MigrationService, error) {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, fmt.Sprintf("pgx5://%s:%s@%s:%d/%s", config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name))
	if err != nil {
		return nil, err
	}
	return &MigrationService{
		migrate: m,
	}, nil
}

func (m *MigrationService) Up() error {
	return m.migrate.Up()
}

func (m *MigrationService) Down() error {
	return m.migrate.Down()
}

func (m *MigrationService) Close() (source error, database error) {
	return m.migrate.Close()
}

func (m *MigrationService) Versions() (version uint, dirty bool, err error) {
	return m.migrate.Version()
}
