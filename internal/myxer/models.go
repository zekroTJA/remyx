package myxer

import "github.com/zmb3/spotify/v2"

type Playlist struct {
	Uid         spotify.ID `json:"uid"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	URL         string     `json:"url"`
	ImageUrl    string     `json:"image_url"`
	NTracks     uint       `json:"n_tracks"`
	OwnerName   string     `json:"owner_name,omitempty"`
}

func PlaylistFromSimplePlaylist(pl *spotify.SimplePlaylist) Playlist {
	rpl := Playlist{
		Uid:         pl.ID,
		Name:        pl.Name,
		Description: pl.Description,
		URL:         string(pl.URI),
		NTracks:     pl.Tracks.Total,
		OwnerName:   pl.Owner.DisplayName,
	}

	if len(pl.Images) > 0 {
		rpl.ImageUrl = pl.Images[0].URL
	}

	return rpl
}
