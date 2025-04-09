package messages

import (
	"context"
	"errors"
	"fmt"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"github.com/szcvak/sps/pkg/messaging"
	"os"
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

	if err != nil {
		if errors.Is(err, database.ErrPlayerNotFound) {
			err = dbm.CreatePlayer(context.Background(), l.HighId, l.LowId, "Undefined", l.Token, l.Region)

			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error creating player: %v\n", err)

				failMsg := NewLoginFailedMessage(l, "The server is currently experiencing some issues. Sorry!", messaging.LoginFailed)
				wrapper.Send(failMsg.PacketId(), failMsg.PacketVersion(), failMsg.Marshal())

				return
			}

			failMsg := NewLoginFailedMessage(l, "Your account has been created. Please reload the game.", messaging.LoginFailed)
			wrapper.Send(failMsg.PacketId(), failMsg.PacketVersion(), failMsg.Marshal())

			return
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "error querying player by token %s: %v\n", l.Token, err)

			failMsg := NewLoginFailedMessage(l, "The server is currently experiencing some issues. Sorry!", messaging.LoginFailed)
			wrapper.Send(failMsg.PacketId(), failMsg.PacketVersion(), failMsg.Marshal())

			return
		}
	}

	stmt := `update players set last_login = current_timestamp where id = $1`

	_, err = dbm.Pool().Exec(context.Background(), stmt, player.DbId)

	if err != nil {
		fmt.Printf("(non-halting) failed to update last_login for player %d: %v\n", player.DbId, err)
	} else {
		fmt.Printf("updated last_login for %s (%d)\n", player.Name, player.DbId)
	}

	player.SetState(core.StateLogin)
	player.LoggedIn = true

	wrapper.Player = player

	msg := NewLoginOkMessage(l)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
