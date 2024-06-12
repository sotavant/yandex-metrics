package test

import (
	"context"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/stretchr/testify/assert"
)

func InitConnection(ctx context.Context, t assert.TestingT) (*pgxpool.Pool, string, string, error) {
	internal.InitLogger()
	dns := os.Getenv("TEST_DATABASE_DSN")
	tableName := os.Getenv("TEST_TABLE_NAME")

	if dns == "" || tableName == "" {
		return nil, "", "", nil
	}

	dbConn, err := pgxpool.New(ctx, dns)

	if err != nil {
		return nil, "", "", err
	}

	err = dbConn.Ping(ctx)
	if err != nil {
		return nil, "", "", err
	}

	err = createTable(ctx, dbConn, tableName)
	assert.NoError(t, err)

	return dbConn, tableName, dns, nil
}

func DropTable(ctx context.Context, conn *pgxpool.Pool, tableName string) error {
	_, err := conn.Exec(ctx, strings.ReplaceAll("drop table if exists $1 cascade", "$1", tableName))
	return err
}

func TruncateTable(ctx context.Context, conn pgx.Conn, tableName string) error {
	_, err := conn.Exec(ctx, strings.ReplaceAll("truncate table $1", "$1", tableName))
	return err
}

func createTable(ctx context.Context, conn *pgxpool.Pool, tableName string) error {
	query := strings.ReplaceAll(`create table if not exists #T
		(
			id    varchar not null,
			type  varchar not null,
			delta int8,
			value double precision,
			constraint #T_pk
				unique (id, type)
		);`, "#T", tableName)

	_, err := conn.Exec(ctx, query)

	return err
}
