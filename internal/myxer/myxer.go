package myxer

import (
	"context"
	"errors"

	"github.com/zekrotja/remyx/internal/database"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Myxer struct {
	db   database.Database
	auth *spotifyauth.Authenticator
}

func New(db database.Database, auth *spotifyauth.Authenticator) *Myxer {
	var t Myxer

	t.db = db
	t.auth = auth

	return &t
}

func (t *Myxer) ScheduleSyncs(remyxUids ...string) error {
	// TODO: add logic here to better schedule syncs

	var mErr error

	if len(remyxUids) == 0 {
		myxes, err := t.db.ListRemyxes("")
		if err != nil {
			return err
		}
		for _, mx := range myxes {
			remyxUids = append(remyxUids, mx.Uid)
		}
	}

	for _, uid := range remyxUids {
		err := t.Sync(uid)
		if err != nil {
			err = errors.Join(mErr, err)
		}
	}

	return mErr
}

func (t *Myxer) Sync(remyxUid string) error {
	tx, err := t.db.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get remyx details

	rmx, err := tx.GetRemyx(remyxUid)
	if err != nil {
		return err
	}

	// Get all source playlists

	sources, err := tx.GetSourcePlaylists(remyxUid)
	if err != nil {
		return err
	}

	// Get songs of source playlists

	sourceTracks := make([][]spotify.PlaylistItem, 0, len(sources))
	for _, pl := range sources {
		tracks, err := t.getSongs(tx, pl.UserUid, pl.PlaylistUid, rmx.Head)
		if err != nil {
			return err
		}
		sourceTracks = append(sourceTracks, tracks)
	}

	// Mix tracks

	remyxedTracks := make([]spotify.ID, 0, len(sources)*rmx.Head)
	for i := 0; i < rmx.Head; i++ {
		for _, tracks := range sourceTracks {
			if i < len(tracks) && tracks[i].Track.Track != nil {
				remyxedTracks = append(remyxedTracks, tracks[i].Track.Track.ID)
			}
		}
	}

	// Create or update traget playlists

	targets, err := tx.GetTargetPlaylists(remyxUid)
	if err != nil {
		return err
	}

	for _, source := range sources {
		contained := false
		for _, target := range targets {
			if source.UserUid == target.UserUid {
				contained = true
				break
			}
		}
		if !contained {
			id, err := t.createTargetPlaylist(tx, source.UserUid)
			if err != nil {
				return err
			}
			newTarget := database.RemyxPlaylist{
				RemyxUid:    rmx.Uid,
				PlaylistUid: id,
				UserUid:     source.UserUid,
			}
			err = tx.AddTargetPlaylist(newTarget)
			if err != nil {
				return err
			}
			targets = append(targets, newTarget)
		}
	}

	for _, target := range targets {
		err = t.updatePlaylist(tx, target.UserUid, target.PlaylistUid, remyxedTracks)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Myxer) getClient(ctx context.Context, tx database.Transaction, userUid string) (*spotify.Client, error) {
	session, err := tx.GetSessionByUserId(userUid)
	if err != nil {
		return nil, err
	}

	httpClient := t.auth.Client(ctx, &oauth2.Token{
		RefreshToken: session.RefreshToken,
	})
	client := spotify.New(httpClient)

	return client, nil
}

func (t *Myxer) getSongs(
	tx database.Transaction,
	userUid string,
	playlistUid spotify.ID,
	limit int,
) ([]spotify.PlaylistItem, error) {
	ctx := context.Background()

	client, err := t.getClient(ctx, tx, userUid)
	if err != nil {
		return nil, err
	}

	tracks, err := client.GetPlaylistItems(ctx, playlistUid, spotify.Limit(limit))
	if err != nil {
		return nil, err
	}

	return tracks.Items, nil
}

func (t *Myxer) createTargetPlaylist(tx database.Transaction, userUid string) (spotify.ID, error) {
	ctx := context.Background()

	client, err := t.getClient(ctx, tx, userUid)
	if err != nil {
		return "", err
	}

	pl, err := client.CreatePlaylistForUser(ctx, userUid, "My Remyx", "", false, false)
	if err != nil {
		return "", err
	}

	return pl.ID, err
}

func (t *Myxer) updatePlaylist(tx database.Transaction, userUid string, playlistUid spotify.ID, remyxTracks []spotify.ID) error {
	ctx := context.Background()

	client, err := t.getClient(ctx, tx, userUid)
	if err != nil {
		return err
	}

	// TODO: make this paged
	items, err := client.GetPlaylistItems(ctx, playlistUid)
	// TODO: capsule that in a util function
	if spErr, ok := err.(spotify.Error); ok && spErr.Status == 404 {
		playlistUid, err = t.createTargetPlaylist(tx, userUid)
	}
	if err != nil {
		return err
	}

	tracks := make([]spotify.ID, 0, len(items.Items))
	for _, item := range items.Items {
		if item.Track.Track != nil {
			tracks = append(tracks, item.Track.Track.ID)
		}
	}

	if len(tracks) > 0 {
		_, err = client.RemoveTracksFromPlaylist(ctx, playlistUid, tracks...)
		if err != nil {
			return err
		}
	}

	_, err = client.AddTracksToPlaylist(ctx, playlistUid, remyxTracks...)
	return err
}
