package messages

import (
	"time"

	"github.com/szcvak/sps/pkg/core"
)

type AllianceChatServerMessage struct {
	allianceId int64
	msg        core.AllianceMessage
}

func NewAllianceChatServerMessage(msg core.AllianceMessage, allianceId int64) *AllianceChatServerMessage {
	return &AllianceChatServerMessage{
		allianceId: allianceId,
		msg:        msg,
	}
}

func (a *AllianceChatServerMessage) PacketId() uint16 {
	return 24312
}

func (a *AllianceChatServerMessage) PacketVersion() uint16 {
	return 1
}

func (a *AllianceChatServerMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(64)

	if a.msg.Type > 40 && a.msg.Type < 50 {
		stream.Write(core.VInt(4))
	} else {
		stream.Write(core.VInt(a.msg.Type))
	}

	dispatchEntry(stream, a.msg)

	return stream.Buffer()
}

// --- Stream functions --- //

func dispatchEntry(stream *core.ByteStream, msg core.AllianceMessage) {
	switch msg.Type {
	case 2:
		chatStreamEntry(stream, msg)
	case 41:
		allianceEventStreamEntry(stream, msg, 1)
	case 43:
		allianceEventStreamEntry(stream, msg, 3)
	case 44:
		allianceEventStreamEntry(stream, msg, 4)
	case 45:
		allianceEventStreamEntry(stream, msg, 5)
	case 46:
		allianceEventStreamEntry(stream, msg, 6)
	}
}

func embedStreamEntry(stream *core.ByteStream, msg core.AllianceMessage) {
	stream.Write(core.LogicLong{0, int32(msg.Id)})

	senderHighId := int32(0)
	senderLowId := int32(0)

	if msg.PlayerId != nil {
		senderHighId = msg.PlayerHighId
		senderLowId = msg.PlayerLowId
	}

	stream.Write(core.LogicLong{senderHighId, senderLowId})

	stream.Write(msg.PlayerName)
	stream.Write(core.VInt(msg.PlayerRole))

	ageSeconds := int32(time.Since(msg.Timestamp).Seconds())

	if ageSeconds < 0 {
		ageSeconds = 0
	}

	stream.Write(core.VInt(ageSeconds))
	stream.Write(false)
}

func chatStreamEntry(stream *core.ByteStream, msg core.AllianceMessage) {
	embedStreamEntry(stream, msg)
	stream.Write(msg.Content)
}

func allianceEventStreamEntry(stream *core.ByteStream, msg core.AllianceMessage, event int32) {
	embedStreamEntry(stream, msg)

	stream.Write(core.VInt(event))
	stream.Write(true)
	stream.Write(core.LogicLong{0, int32(*msg.TargetId)})
	stream.Write(msg.TargetName)
}