package database

import (
	"database/sql"
	"os"

	pq "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/zekrotja/rogu/log"
)

type PostgresDriver struct {
	*SQLDriver
	db *sql.DB
}

var _ Database = (*PostgresDriver)(nil)

func NewPostgresDriver(dsn string) (*PostgresDriver, error) {
	var (
		t   PostgresDriver
		err error
	)

	t.db, err = sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	t.SQLDriver = newSqlDriver(t.db, pgErrWrapper)

	goose.SetBaseFS(os.DirFS("./migrations"))
	goose.SetDialect("postgres")
	err = goose.Up(t.db, "postgres")
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *PostgresDriver) BeginTx() (Transaction, error) {
	tx, err := t.db.Begin()
	if err != nil {
		return nil, err
	}
	return newTx(tx, pgErrWrapper), nil
}

func (t *PostgresDriver) Close() error {
	if t.db == nil {
		return nil
	}
	return t.db.Close()
}

func pgErrWrapper(err error) error {
	if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
		log.Debug().Err(pgErr).Msg("PG Conflict Error")
		return ErrConflict
	}
	return err
}
