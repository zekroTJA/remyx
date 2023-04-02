package database

type Database interface {
	BeginTx() (Transaction, error)
	Close() error

	AddSession(session Session) error
	DeleteSession(uid string) error
	GetSessionByUserId(userId string) (Session, error)

	AddRemyx(rmx Remyx) error
	UpdateRemyx(rmx Remyx) error
	DeleteRemyx(uid string) error
	GetRemyx(uid string) (Remyx, error)
	ListRemyxes(userId string) ([]RemyxWithCount, error)

	AddSourcePlaylist(pl RemyxPlaylist) error
	DeleteSourcePlaylist(remyxUid, userUid, playlistUid string) error
	GetSourcePlaylists(remyxUid string) ([]RemyxPlaylist, error)

	AddTargetPlaylist(pl RemyxPlaylist) error
	DeleteTargetPlaylist(remyxUid, userUid string) error
	GetTargetPlaylists(remyxUid string) ([]RemyxPlaylist, error)
}

type Transaction interface {
	Database
	Commit() error
	Rollback() error
}
