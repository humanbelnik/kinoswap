package model

type Preference struct {
	Text string
}
type Reaction = bool

const (
	PassReaction  Reaction = true
	SmashReaction Reaction = false
)

type VoteResult struct {
	Results map[*MovieMeta]Reaction
}
