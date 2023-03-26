package models

type Error struct {
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}
