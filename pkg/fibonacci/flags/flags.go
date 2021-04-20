package flags

import "github.com/urfave/cli"

var (
	DBHost = cli.StringFlag{
		Name:        "db-host",
		Usage:       "specify the db host name",
		EnvVar:      "DB_HOST",
		Value:       "localhost",
	}
)


