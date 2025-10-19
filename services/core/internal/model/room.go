package model

import "github.com/google/uuid"

type RoomStatus = string

const (
	StatusLobby  RoomStatus = "LOBBY"
	StatusVoting RoomStatus = "VOTING"
)

type Room struct {
	ID         uuid.UUID
	PublicCode string
	Status     string
}
