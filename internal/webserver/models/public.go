package models

import (
	"time"

	"github.com/zekrotja/remyx/internal/database"
	"github.com/zekrotja/remyx/internal/myxer"
	"github.com/zmb3/spotify/v2"
)

type RemyxCreateRequest struct {
	PlaylistId spotify.ID `json:"playlist_id"`
	Head       int        `json:"head"`
	Name       *string    `json:"name"`
}

type RemyxCreateResponse struct {
	Uid     string    `json:"uid"`
	Expires time.Time `json:"expires"`
}

type RemyxConnectRequest struct {
	PlaylistId spotify.ID `json:"playlist_id"`
}

type RemyxWithCount struct {
	database.RemyxWithCount

	Expires time.Time `json:"expires"`
}

type RemyxWithPlaylists struct {
	database.Remyx

	Playlists []myxer.Playlist `json:"playlists"`
}

type MyRemyxesResponse struct {
	Created   []RemyxWithCount `json:"created"`
	Connected []RemyxWithCount `json:"connected"`
}
