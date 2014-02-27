package kocha

type DatabaseMap map[string]DatabaseConfig

// DatabaseConfig represents a configuration of the database.
type DatabaseConfig struct {
	// name of database driver such as "mysql".
	Driver string

	// Data Source Name.
	// e.g. such as "travis@/db_name".
	DSN string
}
