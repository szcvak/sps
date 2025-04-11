package messages

import (
	"context"
	"errors"
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"github.com/szcvak/sps/pkg/messaging"
)

type LoginMessage struct {
	HighId int32
	LowId  int32

	Token string

	Major, Minor, Build int32

	FingerprintSha   string
	DeviceUuid       string
	DeviceIdentifier string
	Region           string

	SystemLanguage int32

	unmarshalled bool
}

func NewLoginMessage() *LoginMessage {
	return &LoginMessage{}
}

func (l *LoginMessage) Unmarshalled() bool {
	return l.unmarshalled
}

func (l *LoginMessage) Unmarshal(payload []byte) {
	stream := core.NewByteStream(payload)
	defer stream.Close()

	l.HighId, _ = stream.ReadInt()
	l.LowId, _ = stream.ReadInt()

	l.Token, _ = stream.ReadString()

	l.Major, _ = stream.ReadInt()
	l.Minor, _ = stream.ReadInt()
	l.Build, _ = stream.ReadInt()

	l.FingerprintSha, _ = stream.ReadString()
	_, _ = stream.ReadString()

	l.DeviceUuid, _ = stream.ReadString()
	_, _ = stream.ReadString()

	l.DeviceIdentifier, _ = stream.ReadString()

	lang, _ := stream.ReadVInt()
	l.SystemLanguage = int32(lang)

	l.Region, _ = stream.ReadString()

	l.unmarshalled = true
}

func (l *LoginMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if !l.Unmarshalled() {
		return
	}

	player, err := dbm.LoadPlayerByToken(context.Background(), l.Token)
	isNew := false

	if err != nil {
		if errors.Is(err, database.ErrPlayerNotFound) {
			temp, err := dbm.CreatePlayer(context.Background(), l.HighId, l.LowId, "Undefined", l.Token, l.Region)

			if err != nil {
				slog.Error("failed to create player!", "err", err)

				failMsg := NewLoginFailedMessage(l, "The server is currently experiencing some issues. Sorry!", messaging.LoginFailed)
				wrapper.Send(failMsg.PacketId(), failMsg.PacketVersion(), failMsg.Marshal())

				return
			}

			isNew = true
			player = temp
		} else {
			slog.Error("failed to find player!", "token", l.Token, "err", err)

			failMsg := NewLoginFailedMessage(l, "The server is currently experiencing some issues. Sorry!", messaging.LoginFailed)
			wrapper.Send(failMsg.PacketId(), failMsg.PacketVersion(), failMsg.Marshal())

			return
		}
	}

	if !isNew {
		stmt := `update players set last_login = current_timestamp where id = $1`

		_, err = dbm.Pool().Exec(context.Background(), stmt, player.DbId)

		if err != nil {
			slog.Warn("(failed to update last_login!", "playerId", player.DbId, "err", err)
		}
	}

	player.SetState(core.StateLogin)
	wrapper.Player = player

	msg := NewLoginOkMessage(l)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())

	msg2 := NewOwnHomeDataMessage(wrapper, dbm)
	wrapper.Send(msg2.PacketId(), msg2.PacketVersion(), msg2.Marshal())

	msg3 := NewClanStreamMessage()
	wrapper.Send(msg3.PacketId(), msg3.PacketVersion(), msg3.Marshal())

	msg4 := NewMyAllianceMessage()
	wrapper.Send(msg4.PacketId(), msg4.PacketVersion(), msg4.Marshal())
}
