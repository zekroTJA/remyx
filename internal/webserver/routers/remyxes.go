package routers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zekrotja/remyx/internal/database"
	"github.com/zekrotja/remyx/internal/myxer"
	"github.com/zekrotja/remyx/internal/shared"
	"github.com/zekrotja/remyx/internal/webserver/middleware"
	"github.com/zekrotja/remyx/internal/webserver/models"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type routerRemyxes struct {
	db  database.Database
	mxr *myxer.Myxer
}

func Remyxes(rg *gin.RouterGroup, db database.Database, mxr *myxer.Myxer) {
	r := routerRemyxes{
		db:  db,
		mxr: mxr,
	}

	rg.GET("", r.listMine)
	rg.POST("/create", middleware.Ratelimit(3, 1*time.Minute), r.create)
	rg.POST("/connect/:id", middleware.Ratelimit(3, 1*time.Hour), r.connect)
	rg.GET("/:id", r.get)
	rg.POST("/:id", r.update)
	rg.DELETE("/:id", r.delete)
	rg.DELETE("/:id/:playlistId", r.deletePlaylist)
}

func (t *routerRemyxes) listMine(ctx *gin.Context) {
	client := ctx.MustGet("client").(*http.Client)
	spClient := spotify.New(client)

	me, err := spClient.CurrentUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed getting current user details", Details: err.Error()})
		return
	}

	rmxs, err := t.db.ListRemyxes(me.ID)
	if err != nil {
		if err == database.ErrNotFound {
			ctx.JSON(http.StatusNotFound,
				models.Error{Message: "no remyxes found for your account"})
		} else {
			ctx.JSON(http.StatusInternalServerError,
				models.Error{Message: "failed getting remyxes", Details: err.Error()})
		}
		return
	}

	res := models.MyRemyxesResponse{
		Created:   make([]models.RemyxWithCount, 0, len(rmxs)),
		Connected: make([]models.RemyxWithCount, 0, len(rmxs)),
	}

	for _, rmx := range rmxs {
		if rmx.CreatorUid == me.ID {
			res.Created = append(res.Created, models.RemyxWithCount{
				RemyxWithCount: rmx,
				Expires:        rmx.CreatedAt.Add(shared.RemyxExpiry),
			})
		} else {
			res.Connected = append(res.Connected, models.RemyxWithCount{
				RemyxWithCount: rmx,
				Expires:        rmx.CreatedAt.Add(shared.RemyxExpiry),
			})
		}
	}

	ctx.JSON(http.StatusOK, res)
}

func (t routerRemyxes) get(ctx *gin.Context) {
	id := ctx.Param("id")

	client := ctx.MustGet("client").(*http.Client)
	spClient := spotify.New(client)

	me, err := spClient.CurrentUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed getting current user details", Details: err.Error()})
		return
	}

	rmx, err := t.db.GetRemyx(id)
	if err != nil {
		if err == database.ErrNotFound {
			ctx.JSON(http.StatusNotFound,
				models.Error{Message: "remyx with this id could not be found"})
		} else {
			ctx.JSON(http.StatusInternalServerError,
				models.Error{Message: "failed getting remyx entry", Details: err.Error()})
		}
		return
	}

	sources, err := t.db.GetSourcePlaylists(rmx.Uid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed getting remyx sources", Details: err.Error()})
		return
	}

	pls, err := t.mxr.GetPlaylistInfo(ctx.Request.Context(), sources)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed getting remyx sources info", Details: err.Error()})
		return
	}

	res := models.RemyxWithPlaylists{
		Remyx: models.Remyx{
			Remyx:   rmx,
			Mine:    rmx.CreatorUid == me.ID,
			Expires: rmx.CreatedAt.Add(shared.RemyxExpiry),
		},
		Playlists: pls,
	}

	ctx.JSON(http.StatusOK, res)
}

func (t routerRemyxes) create(ctx *gin.Context) {
	var req models.RemyxCreateRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest,
			models.Error{Message: "invalid json body", Details: err})
		return
	}

	if req.Head < 1 || req.Head > 50 {
		ctx.JSON(http.StatusBadRequest,
			models.Error{Message: "head count must be in range (0, 50]"})
		return
	}

	client := ctx.MustGet("client").(*http.Client)
	spClient := spotify.New(client)

	me, err := spClient.CurrentUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed getting current user details", Details: err.Error()})
		return
	}

	if req.PlaylistId != shared.LibraryPlaylistId {
		_, err := spClient.GetPlaylist(ctx.Request.Context(), req.PlaylistId)
		if err != nil {
			if spErr, ok := err.(spotify.Error); ok && spErr.Status == 404 {
				ctx.JSON(http.StatusBadRequest,
					models.Error{Message: "the specified playlist could not be found"})
			} else {
				ctx.JSON(http.StatusInternalServerError,
					models.Error{Message: "failed getting playlist details", Details: err.Error()})
			}
			return
		}
	}

	token := ctx.MustGet("token").(*oauth2.Token)

	tx, err := t.db.BeginTx()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed creating database transaction", Details: err.Error()})
		return
	}
	defer tx.Rollback()

	err = tx.AddSession(database.Session{
		Entity:       database.NewEntity(),
		UserId:       me.ID,
		RefreshToken: token.RefreshToken,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed creating session entry", Details: err.Error()})
		return
	}

	rmx := database.Remyx{
		Entity:     database.NewEntity(),
		CreatorUid: me.ID,
		Head:       req.Head,
		Name:       req.Name,
	}
	err = tx.AddRemyx(rmx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed creating remyx entry", Details: err.Error()})
		return
	}

	err = tx.AddSourcePlaylist(database.RemyxPlaylist{
		RemyxUid:    rmx.Uid,
		PlaylistUid: req.PlaylistId,
		UserUid:     me.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed creating playlist entry", Details: err.Error()})
		return
	}

	err = tx.Commit()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed committing changes", Details: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, models.RemyxCreateResponse{
		Uid:     rmx.Uid,
		Expires: rmx.CreatedAt.Add(shared.RemyxExpiry),
	})
}

func (t routerRemyxes) update(ctx *gin.Context) {
	var req models.RemyxUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest,
			models.Error{Message: "invalid json body", Details: err})
		return
	}

	id := ctx.Param("id")

	tx, err := t.db.BeginTx()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed creating transaction", Details: err.Error()})
		return
	}
	defer tx.Rollback()

	rmx, err := tx.GetRemyx(id)
	if err != nil {
		if err == database.ErrNotFound {
			ctx.JSON(http.StatusNotFound,
				models.Error{Message: "remyx with this id could not be found"})
		} else {
			ctx.JSON(http.StatusInternalServerError,
				models.Error{Message: "failed getting remyx entry", Details: err.Error()})
		}
		return
	}

	if req.Head != nil {
		head := *req.Head
		if head < 1 || head > 50 {
			ctx.JSON(http.StatusBadRequest,
				models.Error{Message: "head count must be in range (0, 50]"})
			return
		}

		rmx.Head = head
	}

	if req.Name != nil {
		rmx.Name = req.Name
	}

	err = tx.UpdateRemyx(rmx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed updating remyx", Details: err.Error()})
		return
	}

	err = tx.Commit()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed applying transaction", Details: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, rmx)
}

func (t routerRemyxes) delete(ctx *gin.Context) {
	id := ctx.Param("id")

	client := ctx.MustGet("client").(*http.Client)
	spClient := spotify.New(client)

	me, err := spClient.CurrentUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed getting current user details", Details: err.Error()})
		return
	}

	tx, err := t.db.BeginTx()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed creating database transaction", Details: err.Error()})
		return
	}
	defer tx.Rollback()

	rmx, err := tx.GetRemyx(id)
	if err == database.ErrNotFound {
		ctx.JSON(http.StatusNotFound, models.Error{Message: "not found"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed getting remyx", Details: err})
		return
	}

	if rmx.CreatorUid == me.ID {
		err = tx.DeleteRemyx(id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError,
				models.Error{Message: "failed deleting remyx", Details: err})
			return
		}
	} else {
		sources, err := tx.GetSourcePlaylists(id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError,
				models.Error{Message: "failed getting remyx sources", Details: err})
			return
		}

		var mySources []database.RemyxPlaylist
		for _, source := range sources {
			if source.UserUid == me.ID {
				mySources = append(mySources, source)
			}
		}

		if len(mySources) == 0 {
			ctx.JSON(http.StatusNotFound, models.Error{Message: "not found"})
			return
		}

		for _, source := range mySources {
			err = tx.DeleteSourcePlaylist(id, "", string(source.PlaylistUid))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError,
					models.Error{Message: "failed deleting remyx sources", Details: err})
				return
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed committing changes", Details: err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (t routerRemyxes) connect(ctx *gin.Context) {
	id := ctx.Param("id")

	var req models.RemyxConnectRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest,
			models.Error{Message: "invalid json body", Details: err})
		return
	}

	client := ctx.MustGet("client").(*http.Client)
	spClient := spotify.New(client)

	me, err := spClient.CurrentUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed getting current user details", Details: err.Error()})
		return
	}

	tx, err := t.db.BeginTx()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed creating database transaction", Details: err.Error()})
		return
	}
	defer tx.Rollback()

	token := ctx.MustGet("token").(*oauth2.Token)
	err = tx.AddSession(database.Session{
		Entity:       database.NewEntity(),
		UserId:       me.ID,
		RefreshToken: token.RefreshToken,
	})
	if err != nil && err != database.ErrConflict {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed creating session entry", Details: err.Error()})
		return
	}

	rmx, err := tx.GetRemyx(id)
	if err != nil {
		if err == database.ErrNotFound {
			ctx.JSON(http.StatusNotFound,
				models.Error{Message: "remyx with this id could not be found"})
		} else {
			ctx.JSON(http.StatusInternalServerError,
				models.Error{Message: "failed getting remyx entry", Details: err.Error()})
		}
		return
	}

	if req.PlaylistId != shared.LibraryPlaylistId {
		_, err := spClient.GetPlaylist(ctx.Request.Context(), req.PlaylistId)
		if err != nil {
			if spErr, ok := err.(spotify.Error); ok && spErr.Status == 404 {
				ctx.JSON(http.StatusBadRequest,
					models.Error{Message: "the specified playlist could not be found"})
			} else {
				ctx.JSON(http.StatusInternalServerError,
					models.Error{Message: "failed getting playlist details", Details: err.Error()})
			}
			return
		}
	}

	err = tx.AddSourcePlaylist(database.RemyxPlaylist{
		RemyxUid:    rmx.Uid,
		PlaylistUid: req.PlaylistId,
		UserUid:     me.ID,
	})
	if err != nil {
		if err == database.ErrConflict {
			ctx.JSON(http.StatusConflict,
				models.Error{Message: "the playlist has already been connected to this remyx"})
		} else {
			ctx.JSON(http.StatusInternalServerError,
				models.Error{Message: "failed creating playlist entry", Details: err.Error()})
		}
		return
	}

	err = tx.Commit()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed committing changes", Details: err.Error()})
		return
	}

	go t.mxr.ScheduleSyncs(rmx.Uid)

	ctx.JSON(http.StatusOK, models.RemyxCreateResponse{
		Uid:     rmx.Uid,
		Expires: rmx.CreatedAt.Add(shared.RemyxExpiry),
	})
}

func (t routerRemyxes) deletePlaylist(ctx *gin.Context) {
	id := ctx.Param("id")
	playlistId := ctx.Param("playlistId")

	client := ctx.MustGet("client").(*http.Client)
	spClient := spotify.New(client)

	me, err := spClient.CurrentUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed getting current user details", Details: err.Error()})
		return
	}

	tx, err := t.db.BeginTx()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed creating database transaction", Details: err.Error()})
		return
	}
	defer tx.Rollback()

	rmx, err := tx.GetRemyx(id)
	if err != nil {
		if err == database.ErrNotFound {
			ctx.JSON(http.StatusNotFound,
				models.Error{Message: "remyx with this id could not be found"})
		} else {
			ctx.JSON(http.StatusInternalServerError,
				models.Error{Message: "failed getting remyx entry", Details: err.Error()})
		}
		return
	}

	sources, err := tx.GetSourcePlaylists(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed getting remyx sources", Details: err.Error()})
		return
	}

	var source *database.RemyxPlaylist
	for i := range sources {
		if sources[i].PlaylistUid == spotify.ID(playlistId) {
			source = &sources[i]
		}
	}
	if source == nil {
		ctx.JSON(http.StatusNotFound,
			models.Error{Message: "source playlist not found", Details: err.Error()})
		return
	}

	remyxDeleted := false
	if rmx.CreatorUid == me.ID {
		if len(sources)-1 <= 0 {
			remyxDeleted = true
			err = tx.DeleteRemyx(id)
		} else {
			err = tx.DeleteSourcePlaylist(id, me.ID, playlistId)
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError,
				models.Error{Message: "failed deleting source playlist", Details: err.Error()})
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			models.Error{Message: "failed committing changes", Details: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.RemyxPlaylistDeleetResponse{
		RemyxDeleted: remyxDeleted,
	})
}
