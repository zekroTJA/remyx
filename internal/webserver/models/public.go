package models

import (
	"time"

	"github.com/zekrotja/remyx/internal/database"
	"github.com/zmb3/spotify/v2"
)

type Playlist struct {
	Uid         spotify.ID `json:"uid"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	URL         string     `json:"url"`
	ImageUrl    string     `json:"image_url"`
	NTracks     uint       `json:"n_tracks"`
}

type RemyxCreateRequest struct {
	PlaylistId spotify.ID `json:"playlist_id"`
	Head       int        `json:"head"`
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
