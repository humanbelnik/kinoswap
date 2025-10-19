package model

type Preference struct {
	Text string
}
type Reaction = int

const (
	PassReaction  Reaction = 1
	SmashReaction Reaction = 0
)

type Reactions struct {
	Reactions map[*MovieMeta]Reaction
}

type Result struct {
	MM    MovieMeta
	Likes int
}
