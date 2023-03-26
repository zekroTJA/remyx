package database

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
)

type SQLDriver struct {
	b          sq.StatementBuilderType
	runner     sq.BaseRunner
	errWrapper func(err error) error
}

var _ Database = (*SQLDriver)(nil)

func newSqlDriver(runner sq.BaseRunner, errWrapper func(err error) error) *SQLDriver {
	var t SQLDriver

	t.runner = runner
	t.errWrapper = errWrapper

	t.b = sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		RunWith(t.runner)

	return &t
}

func (t *SQLDriver) BeginTx() (Transaction, error) {
	return nil, errors.New("never call BeginTx() directly on the SQLDriver instance")
}

func (t *SQLDriver) Close() error {
	return errors.New("never call Close() directly on the SQLDriver instance")
}

func (t *SQLDriver) AddRemyx(link Remyx) error {
	_, err := t.b.
		Insert("remyx").
		Columns("uid", "creator_uid", "head", "created_at").
		Values(&link.Uid, &link.CreatorUid, &link.Head, &link.CreatedAt).
		Exec()
	return err
}

func (t *SQLDriver) DeleteRemyx(uid string) error {
	_, err := t.b.
		Delete("remyx").
		Where(sq.Eq{"uid": uid}).
		Exec()
	return err
}

func (t *SQLDriver) GetRemyx(uid string) (Remyx, error) {
	var l Remyx
	err := t.b.
		Select("uid", "head", "creator_uid", "created_at").
		From("remyx").
		Where(sq.Eq{"uid": uid}).
		Scan(&l.Uid, &l.Head, &l.CreatorUid, &l.CreatedAt)
	return l, t.wrapErr(err)
}

func (t *SQLDriver) ListRemyxes(userId string) ([]RemyxWithCount, error) {
	var cond any
	if userId != "" {
		cond = sq.Eq{"creator_uid": userId}
	}

	rows, err := t.b.
		Select("remyx.uid", "created_at", "creator_uid", "head", "COUNT(playlist_uid)").
		From("remyx").
		Join("source_playlist ON remyx.uid = source_playlist.remyx_uid").
		Where(cond).
		GroupBy("remyx.uid").
		Query()
	if err != nil {
		return nil, t.wrapErr(err)
	}

	return scanRows(rows, func(v *RemyxWithCount) []any {
		return []any{&v.Uid, &v.CreatedAt, &v.CreatorUid, &v.Head, &v.PlaylistCount}
	})
}

func (t *SQLDriver) AddSession(session Session) error {
	_, err := t.b.
		Insert("session").
		Columns("uid", "user_id", "refresh_token", "created_at").
		Values(session.Uid, session.UserId, session.RefreshToken, session.CreatedAt).
		Exec()
	return err
}

func (t *SQLDriver) DeleteSession(uid string) error {
	_, err := t.b.
		Delete("session").
		Where(sq.Eq{"uid": uid}).
		Exec()
	return t.wrapErr(err)
}

func (t *SQLDriver) GetSessionByUserId(userId string) (Session, error) {
	var s Session
	err := t.b.
		Select("uid", "user_id", "refresh_token", "created_at").
		From("session").
		Where(sq.Eq{"user_id": userId}).
		Scan(&s.Uid, &s.UserId, &s.RefreshToken, &s.CreatedAt)
	return s, t.wrapErr(err)
}

func (t *SQLDriver) AddSourcePlaylist(pl RemyxPlaylist) error {
	_, err := t.b.
		Insert("source_playlist").
		Columns("remyx_uid", "playlist_uid", "user_uid").
		Values(pl.RemyxUid, pl.PlaylistUid, pl.UserUid).
		Exec()
	return t.wrapErr(err)
}

func (t *SQLDriver) DeleteSourcePlaylist(remyxUid, playlistUid string) error {
	_, err := t.b.
		Delete("source_playlist").
		Where(sq.And{
			sq.Eq{"remyx_uid": remyxUid},
			sq.Eq{"playlist_uid": playlistUid},
		}).
		Exec()
	return t.wrapErr(err)
}

func (t *SQLDriver) GetSourcePlaylists(remyxUid string) ([]RemyxPlaylist, error) {
	rows, err := t.b.
		Select("remyx_uid", "playlist_uid", "user_uid").
		From("source_playlist").
		Where(sq.Eq{"remyx_uid": remyxUid}).
		Query()
	if err != nil {
		return nil, t.wrapErr(err)
	}

	return scanRows(rows, func(v *RemyxPlaylist) []any {
		return []any{&v.RemyxUid, &v.PlaylistUid, &v.UserUid}
	})
}

func (t *SQLDriver) AddTargetPlaylist(pl RemyxPlaylist) error {
	_, err := t.b.
		Insert("target_playlist").
		Columns("remyx_uid", "playlist_uid", "user_uid").
		Values(pl.RemyxUid, pl.PlaylistUid, pl.UserUid).
		Exec()
	return t.wrapErr(err)
}

func (t *SQLDriver) DeleteTargetPlaylist(remyxUid, userUid string) error {
	_, err := t.b.
		Delete("target_playlist").
		Where(sq.And{
			sq.Eq{"remyx_uid": remyxUid},
			sq.Eq{"user_uid": userUid},
		}).
		Exec()
	return t.wrapErr(err)
}

func (t *SQLDriver) GetTargetPlaylists(remyxUid string) ([]RemyxPlaylist, error) {
	rows, err := t.b.
		Select("remyx_uid", "playlist_uid", "user_uid").
		From("target_playlist").
		Where(sq.Eq{"remyx_uid": remyxUid}).
		Query()
	if err != nil {
		return nil, t.wrapErr(err)
	}

	return scanRows(rows, func(v *RemyxPlaylist) []any {
		return []any{&v.RemyxUid, &v.PlaylistUid, &v.UserUid}
	})
}

func (t *SQLDriver) wrapErr(err error) error {
	if err == sql.ErrNoRows {
		err = ErrNotFound
	}
	if t.errWrapper != nil {
		err = t.errWrapper(err)
	}
	return err
}

func scanRows[T any](rows *sql.Rows, scanner func(v *T) []any) ([]T, error) {
	res := make([]T, 0, 5)
	for rows.Next() {
		var pl T
		err := rows.Scan(scanner(&pl)...)
		if err != nil {
			return nil, err
		}
		res = append(res, pl)
	}
	return res, nil
}
