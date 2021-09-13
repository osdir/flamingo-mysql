package migration

import (
	"errors"

	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/cmd"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/spf13/cobra"
	// we need the file source only when the migrator module is included
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/osdir/flamingo-mysql/db"
	"github.com/osdir/flamingo-mysql/migration/application"
)

type (
	// Module basic struct
	Module struct {
		autoMigrate bool
	}
)

// Inject dependencies
func (m *Module) Inject(
	cfg *struct {
		AutoMigrate bool `inject:"config:mysql.migration.automigrate"`
	},
) {
	m.autoMigrate = cfg.AutoMigrate
}

// Configure this for Migration module
func (m *Module) Configure(injector *dingo.Injector) {
	injector.BindMulti(new(cobra.Command)).ToProvider(migrateProvider)
	injector.BindMulti(new(cobra.Command)).ToProvider(seedProvider)
	if m.autoMigrate {
		flamingo.BindEventSubscriber(injector).To(&application.StartUpMigrations{})
	}
}

// FlamingoLegacyConfigAlias maps legacy config entries to new ones
func (m *Module) FlamingoLegacyConfigAlias() map[string]string {
	return map[string]string{
		"migrations": "mysql.migration",
		"seeds":      "mysql.seed",
	}
}

// CueConfig for the module
func (m *Module) CueConfig() string {
	return `
mysql: {
	db: connectionOptions: multiStatements: "true" //required for migration and seed scripts  
	migration: {
		automigrate: bool | *false
		directory: string | *"sql/migrations/"
	}
	seed: {
		directory: string | *"sql/seeds/"
	}
}`
}

// Depends on other modules
func (m *Module) Depends() []dingo.Module {
	return []dingo.Module{
		new(db.Module),
		new(cmd.Module),
	}
}

func exactValidArgs(cmd *cobra.Command, args []string) error {
	err := cobra.ExactArgs(1)(cmd, args)
	if err != nil {
		return err
	}
	err = cobra.OnlyValidArgs(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

func migrateProvider(migrator *application.Migrator) *cobra.Command {
	var migrateCMD = &cobra.Command{
		Use:   "migrate [up|down] (-s[number of steps])",
		Short: "Run migrations from sql/migrations on the DB",
		Long:  "Use migrate up to run all new up migrations, down to run all down migrations. You can limit the number of migrations to run with the steps flag.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var steps *int
			if cmd.Flag("steps").Changed {
				s, _ := cmd.Flags().GetInt("steps")
				steps = &s
			}

			switch mode := args[0]; mode {
			case "up":
				return migrator.Up(steps)
			case "down":
				return migrator.Down(steps)
			default:
				return errors.New("argument up or down missing")
			}
		},
		Args:      exactValidArgs,
		ValidArgs: []string{"up", "down"},
	}
	migrateCMD.Flags().IntP("steps", "s", 0, "Steps to migrate")

	return migrateCMD
}

func seedProvider(seeder *application.Seeder) *cobra.Command {
	var seedCMD = &cobra.Command{
		Use:   "seed",
		Short: "Run all sql files from sql/seeds on the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return seeder.Seed()
		},
		Args: cobra.NoArgs,
	}
	return seedCMD
}
