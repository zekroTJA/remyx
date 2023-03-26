package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zekrotja/remyx/internal/webserver/models"
	"github.com/zekrotja/remyx/internal/webserver/util"
	"github.com/zmb3/spotify/v2"
)

type routerPlaylists struct {
}

func Playlists(rg *gin.RouterGroup) {
	r := routerPlaylists{}

	rg.GET("", r.list)
}

func (t routerPlaylists) list(ctx *gin.Context) {
	client := ctx.MustGet("client").(*http.Client)
	offset := util.QueryInt(ctx, "offset", 0, 0)
	limit := util.QueryInt(ctx, "limit", 20, 1, 50)

	page, err := spotify.New(client).
		CurrentUsersPlaylists(ctx.Request.Context(),
			spotify.Offset(offset),
			spotify.Limit(limit))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Error{
			Message: "failed getting playlists", Details: err.Error(),
		})
		return
	}

	resp := make([]models.Playlist, 0, len(page.Playlists))
	for _, pl := range page.Playlists {
		rpl := models.Playlist{
			Uid:         pl.ID,
			Name:        pl.Name,
			Description: pl.Description,
			URL:         string(pl.URI),
			NTracks:     pl.Tracks.Total,
		}
		if len(pl.Images) > 0 {
			rpl.ImageUrl = pl.Images[0].URL
		}
		resp = append(resp, rpl)
	}

	ctx.JSON(http.StatusOK, resp)
}
