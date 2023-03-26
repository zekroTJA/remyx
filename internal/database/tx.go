package database

import (
	"database/sql"
	"sync"

	_ "github.com/lib/pq"
)

type TX struct {
	sync.Mutex

	*SQLDriver
	tx *sql.Tx

	committed bool
}

var _ Transaction = (*TX)(nil)

func newTx(tx *sql.Tx, errWrapper func(err error) error) Transaction {
	var t TX

	t.SQLDriver = newSqlDriver(tx, errWrapper)
	t.tx = tx

	return &t
}

func (t *TX) Close() error {
	return t.Commit()
}

func (t *TX) Commit() error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	err := t.tx.Commit()
	if err != nil {
		return err
	}
	t.committed = true
	return nil
}

func (t *TX) Rollback() error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	if t.committed {
		return nil
	}

	return t.tx.Rollback()
}
