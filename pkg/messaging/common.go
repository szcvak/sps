package messaging

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type ClientMessage interface {
	Unmarshal(payload []byte)
	Process(player *core.ClientWrapper, dbConn *database.Manager)
}

type ServerMessage interface {
	Marshal() []byte
	PacketId() uint16
	PacketVersion() uint16
}
