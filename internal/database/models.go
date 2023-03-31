package database

import (
	"time"

	"github.com/rs/xid"
	"github.com/zmb3/spotify/v2"
)

type Entity struct {
	Uid       string    `json:"uid"`
	CreatedAt time.Time `json:"created_at"`
}

type Remyx struct {
	Entity

	CreatorUid string  `json:"creator_uid"`
	Head       int     `json:"head"`
	Name       *string `json:"name"`
}

type RemyxWithCount struct {
	Remyx

	PlaylistCount int `json:"playlist_count"`
}

type Session struct {
	Entity

	UserId       string `json:"user_uid"`
	RefreshToken string `json:"refresh_token"`
}

type RemyxPlaylist struct {
	RemyxUid    string     `json:"remyx_uid"`
	PlaylistUid spotify.ID `json:"playlist_uid"`
	UserUid     string     `json:"user_uid"`
}

func NewEntity() Entity {
	return Entity{
		Uid:       xid.New().String(),
		CreatedAt: time.Now(),
	}
}
