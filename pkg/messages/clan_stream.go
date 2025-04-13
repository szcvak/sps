package messages

import (
	"context"
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type ClanStreamMessage struct {
	wrapper *core.ClientWrapper
	dbm     *database.Manager
}

func NewClanStreamMessage(wrapper *core.ClientWrapper, dbm *database.Manager) *ClanStreamMessage {
	return &ClanStreamMessage{
		wrapper: wrapper,
		dbm:     dbm,
	}
}

func (c *ClanStreamMessage) PacketId() uint16 {
	return 24311
}

func (c *ClanStreamMessage) PacketVersion() uint16 {
	return 1
}

func (c *ClanStreamMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(8)

	allianceId := c.wrapper.Player.AllianceId

	if allianceId == nil {
		stream.Write(core.VInt(0))
		stream.Write(core.VInt(0))

		return stream.Buffer()
	}

	msgs, err := c.dbm.LoadAllianceMessages(context.Background(), *allianceId, 50)

	if err != nil {
		slog.Error("failed to load alliance messages!", "err", err)
		return stream.Buffer()
	}

	stream.Write(core.VInt(len(msgs)))

	if len(msgs) == 0 {
		stream.Write(core.VInt(0))
	} else {
		for _, msg := range msgs {
			if msg.Type == 43 || msg.Type == 44 {
				stream.Write(core.VInt(4))
			} else {
				stream.Write(core.VInt(msg.Type))
			}

			dispatchEntry(stream, msg)
		}
	}

	return stream.Buffer()
}
