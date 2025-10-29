package model

import "github.com/google/uuid"

type Preference struct {
	Text string
}
type Reaction = int

const (
	PassReaction  Reaction = 1
	SmashReaction Reaction = 0
)

type Reactions struct {
	Reactions map[uuid.UUID]Reaction
}

type Result struct {
	MM    MovieMeta
	Likes int
}
