package test

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func InitConnection(ctx context.Context, t *testing.T) (*pgx.Conn, string, error) {
	internal.InitLogger()
	dns := os.Getenv("DATABASE_DSN")
	tableName := os.Getenv("TABLE_NAME")

	if dns == "" || tableName == "" {
		return nil, "", nil
	}

	dbConn, err := pgx.Connect(ctx, dns)

	if err != nil {
		return nil, "", err
	}

	err = dbConn.Ping(ctx)
	if err != nil {
		return nil, "", err
	}

	err = createTable(ctx, *dbConn, tableName)
	assert.NoError(t, err)

	return dbConn, tableName, nil
}

func DropTable(ctx context.Context, conn pgx.Conn, tableName string) error {
	_, err := conn.Exec(ctx, strings.ReplaceAll("drop table if exists $1", "$1", tableName))
	return err
}

func TruncateTable(ctx context.Context, conn pgx.Conn, tableName string) error {
	_, err := conn.Exec(ctx, strings.ReplaceAll("truncate table $1", "$1", tableName))
	return err
}

func createTable(ctx context.Context, conn pgx.Conn, tableName string) error {
	query := strings.ReplaceAll(`create table if not exists #T
		(
			id    varchar not null,
			type  varchar not null,
			delta integer,
			value double precision,
			constraint #T_pk
				unique (id, type)
		);`, "#T", tableName)

	_, err := conn.Exec(ctx, query)

	return err
}
