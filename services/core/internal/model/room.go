package model

import "github.com/google/uuid"

type RoomID string

const EmptyRoomID RoomID = ""

func (id RoomID) BuildUUID() uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(id))
}
