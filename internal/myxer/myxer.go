package myxer

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zekrotja/remyx/internal/database"
	"github.com/zekrotja/remyx/internal/shared"
	"github.com/zekrotja/remyx/internal/webserver/util"
	"github.com/zekrotja/rogu/log"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const pageSize = 5

type Myxer struct {
	db         database.Database
	auth       *spotifyauth.Authenticator
	publicAddr string
}

func New(db database.Database, auth *spotifyauth.Authenticator, publicAddr string) *Myxer {
	var t Myxer

	t.db = db
	t.auth = auth
	t.publicAddr = publicAddr

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

	log.Info().Field("n", len(remyxUids)).Msg("Schedule syncs ...")

	for _, uid := range remyxUids {
		err := t.Sync(uid)
		if err != nil {
			mErr = errors.Join(mErr, err)
		}
	}

	log.Info().Field("n", len(remyxUids)).Msg("Sync scheduling finished ...")
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
			err = tx.DeleteSourcePlaylist(remyxUid, pl.UserUid, string(pl.PlaylistUid))
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

	type key struct {
		playlist spotify.ID
		user     string
	}

	mappedPlaylists := make(map[key]*spotify.SimplePlaylist)
	for userID, playlists := range byCreator {
		client, err := t.getClient(ctx, t.db, userID)
		if err != nil {
			return nil, err
		}

		user, err := client.CurrentUser(ctx)
		if err != nil {
			return nil, err
		}

		for _, plID := range playlists {
			if plID == shared.LibraryPlaylistId {
				mappedPlaylists[key{plID, userID}] = &spotify.SimplePlaylist{
					ID:    plID,
					Name:  shared.LibraryPlaylistName,
					Owner: user.User,
				}
				continue
			}

			pl, err := client.GetPlaylist(ctx, plID)
			if err != nil {
				mappedPlaylists[key{plID, userID}] = &spotify.SimplePlaylist{
					ID:    plID,
					Owner: user.User,
				}
				continue
			}

			mappedPlaylists[key{plID, userID}] = &pl.SimplePlaylist
		}
	}

	hydrated := make([]Playlist, 0, len(playlists))
	for _, pl := range playlists {
		hpl := mappedPlaylists[key{pl.PlaylistUid, pl.UserUid}]
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

	description := fmt.Sprintf("Created with Remyx. Here you can find your Remyx: %s/%s", t.publicAddr, rmx.Uid)
	pl, err := client.CreatePlaylistForUser(
		ctx, userUid, plName, description, false, false)
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

func (t *Myxer) getAllPlaylistItems(ctx context.Context, client *spotify.Client, playlistUid spotify.ID) ([]spotify.PlaylistItem, error) {
	var total int
	var res []spotify.PlaylistItem
	var page int

	for {
		items, err := client.GetPlaylistItems(ctx, playlistUid,
			spotify.Limit(pageSize),
			spotify.Offset(page*pageSize))
		if err != nil {
			return nil, err
		}
		if total == 0 {
			total = items.Total
			res = make([]spotify.PlaylistItem, 0, total)
		}
		page++
		res = append(res, items.Items...)
		if len(res) >= total {
			break
		}
	}

	return res, nil
}

func (t *Myxer) updatePlaylist(tx database.Transaction, rmx database.Remyx, userUid string, playlistUid spotify.ID, remyxTracks []spotify.ID) error {
	ctx := context.Background()

	client, err := t.getClient(ctx, tx, userUid)
	if err != nil {
		return err
	}

	items, err := t.getAllPlaylistItems(ctx, client, playlistUid)
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

	tracks := make([]spotify.ID, 0, len(items))
	for _, item := range items {
		if item.Track.Track != nil {
			tracks = append(tracks, item.Track.Track.ID)
		}
	}

	if len(tracks) > 0 {
		fmt.Println(len(tracks), tracks)
		err = util.DoPaged(tracks, pageSize, func(t []spotify.ID) error {
			log.Debug().Fields("pl", playlistUid, "n", len(t)).Msg("Removing songs from playlist ...")
			_, err = client.RemoveTracksFromPlaylist(ctx, playlistUid, t...)
			return err
		})
		if err != nil {
			return err
		}
	}

	err = util.DoPaged(remyxTracks, pageSize, func(t []spotify.ID) error {
		log.Debug().Fields("pl", playlistUid, "n", len(t)).Msg("Adding songs to playlist ...")
		_, err = client.AddTracksToPlaylist(ctx, playlistUid, t...)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}
