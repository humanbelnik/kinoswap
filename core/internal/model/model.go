package model

import "github.com/google/uuid"

type RoomID = string

const EmptyRoomID RoomID = ""

type Preference struct {
	Text string
}

type MovieMeta struct {
	ID         uuid.UUID
	PosterLink string
	Title      string
	Genres     []string
	Year       int
	Rating     float64

	Overview string
}

type Embedding struct {
	ID uuid.UUID
	E  []byte
}

type Reaction = bool

const (
	PassReaction  Reaction = true
	SmashReaction Reaction = false
)

type VoteResult struct {
	Results map[*MovieMeta]Reaction
}
