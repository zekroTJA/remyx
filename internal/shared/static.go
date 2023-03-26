package shared

import "time"

const (
	AuthTokenCookie = "remyx_uid"
	SessionLifetime = 24 * time.Hour

	RemyxExpiry = 24 * time.Hour * 365
)
