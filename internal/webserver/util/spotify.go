package util

import "github.com/zmb3/spotify/v2"

func IsSpotifyError(err error, code int) bool {
	spErr, ok := err.(spotify.Error)
	return ok && spErr.Status == code
}
