package model

import "github.com/google/uuid"

type RoomID string

const EmptyRoomID RoomID = ""

func (id RoomID) BuildUUID() uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(id))
}

type Preference struct {
	Text string
}

const EmptyTitle string = ""

type MovieMeta struct {
	ID         uuid.UUID
	PosterLink string
	Title      string
	Genres     []string
	Year       int
	Rating     float64

	Overview string
}

type Embedding []float32

type Reaction = bool

const (
	PassReaction  Reaction = true
	SmashReaction Reaction = false
)

type VoteResult struct {
	Results map[*MovieMeta]Reaction
}
