package myxer

import (
	"context"
	"errors"
	"net/http"

	"github.com/zekrotja/remyx/internal/database"
	"github.com/zekrotja/remyx/internal/shared"
	"github.com/zekrotja/remyx/internal/webserver/util"
	"github.com/zekrotja/rogu/log"
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

	log.Debug().Field("n", len(remyxUids)).Msg("Schedule syncs ...")

	for _, uid := range remyxUids {
		err := t.Sync(uid)
		if err != nil {
			mErr = errors.Join(mErr, err)
		}
	}

	log.Debug().Field("n", len(remyxUids)).Msg("Sync scheduling finished ...")
	if mErr != nil {
		log.Error().Err(mErr).Msg("Scheduled sync failed")
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
		if util.IsSpotifyError(err, http.StatusNotFound) {
			log.Debug().Fields("rmx", remyxUid, "pl", pl.PlaylistUid).Msg("Removing source playlist")
			err = tx.DeleteSourcePlaylist(remyxUid, string(pl.PlaylistUid))
			if err != nil {
				return err
			}
			if len(sources)-1 <= 1 {
				log.Debug().Fields("rmx", remyxUid).Msg("Cancel sync because only one source playlist is available")
				return tx.Commit()
			}
		}
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
			newTarget, err := t.createTargetPlaylist(tx, rmx, source.UserUid)
			if err != nil {
				return err
			}
			targets = append(targets, newTarget)
		}
	}

	for _, target := range targets {
		err = t.updatePlaylist(tx, rmx, target.UserUid, target.PlaylistUid, remyxedTracks)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (t *Myxer) GetPlaylistInfo(ctx context.Context, playlists []database.RemyxPlaylist) ([]Playlist, error) {
	byCreator := make(map[string][]spotify.ID)
	for _, pl := range playlists {
		byCreator[pl.UserUid] = append(byCreator[pl.UserUid], pl.PlaylistUid)
	}

	mappedPlaylists := make(map[spotify.ID]*spotify.SimplePlaylist)
	for userID, playlists := range byCreator {
		client, err := t.getClient(ctx, t.db, userID)
		if err != nil {
			return nil, err
		}
		for _, plID := range playlists {
			pl, err := client.GetPlaylist(ctx, plID)
			if err != nil {
				mappedPlaylists[plID] = &spotify.SimplePlaylist{
					ID: plID,
				}
				continue
			}
			mappedPlaylists[plID] = &pl.SimplePlaylist
		}
	}

	hydrated := make([]Playlist, 0, len(playlists))
	for _, pl := range playlists {
		hpl := mappedPlaylists[pl.PlaylistUid]
		hydrated = append(hydrated, PlaylistFromSimplePlaylist(hpl))
	}

	return hydrated, nil
}

func (t *Myxer) getClient(ctx context.Context, tx database.Database, userUid string) (*spotify.Client, error) {
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

	var items []spotify.PlaylistItem
	if playlistUid == shared.LibraryPlaylistId {
		tracks, err := client.CurrentUsersTracks(ctx, spotify.Limit(limit))
		if err != nil {
			return nil, err
		}
		items = make([]spotify.PlaylistItem, 0, len(tracks.Tracks))
		for i := range tracks.Tracks {
			items = append(items, spotify.PlaylistItem{
				Track: spotify.PlaylistItemTrack{
					Track: &tracks.Tracks[i].FullTrack,
				},
			})
		}
	} else {
		tracks, err := client.GetPlaylistItems(ctx, playlistUid, spotify.Limit(limit))
		if err != nil {
			return nil, err
		}
		items = tracks.Items
	}

	return items, nil
}

func (t *Myxer) createTargetPlaylist(tx database.Transaction, rmx database.Remyx, userUid string) (database.RemyxPlaylist, error) {
	ctx := context.Background()

	client, err := t.getClient(ctx, tx, userUid)
	if err != nil {
		return database.RemyxPlaylist{}, err
	}

	plName := "My Remyx"
	if rmx.Name != nil {
		plName = *rmx.Name
	}
	pl, err := client.CreatePlaylistForUser(ctx, userUid, plName, "", false, false)
	if err != nil {
		return database.RemyxPlaylist{}, err
	}

	target := database.RemyxPlaylist{
		RemyxUid:    rmx.Uid,
		PlaylistUid: pl.ID,
		UserUid:     userUid,
	}
	err = tx.AddTargetPlaylist(target)

	return target, err
}

func (t *Myxer) updatePlaylist(tx database.Transaction, rmx database.Remyx, userUid string, playlistUid spotify.ID, remyxTracks []spotify.ID) error {
	ctx := context.Background()

	client, err := t.getClient(ctx, tx, userUid)
	if err != nil {
		return err
	}

	// TODO: make this paged
	items, err := client.GetPlaylistItems(ctx, playlistUid)
	if util.IsSpotifyError(err, http.StatusNotFound) {
		err = tx.DeleteTargetPlaylist(rmx.Uid, userUid)
		if err != nil {
			return err
		}
		pl, err := t.createTargetPlaylist(tx, rmx, userUid)
		if err != nil {
			return err
		}
		return t.updatePlaylist(tx, rmx, userUid, pl.PlaylistUid, remyxTracks)
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
