package core

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/crypt"
)

type ClientWrapper struct {
	conn   net.Conn
	player *Player

	encryptor *crypt.Rc4
	decryptor *crypt.Rc4
}

func NewClientWrapper(conn net.Conn) *ClientWrapper {
	fullKey := append([]byte(config.Rc4Key), []byte(config.Rc4KeyNonce)...)

	encryptor := crypt.NewRc4(fullKey)
	decryptor := crypt.NewRc4(fullKey)

	encryptor.Process(fullKey)
	decryptor.Process(fullKey)

	return &ClientWrapper{
		conn:   conn,
		player: NewPlayer("Brawler"),

		encryptor: encryptor,
		decryptor: decryptor,
	}
}

func (w *ClientWrapper) Close() {
	w.conn.Close()
}

func (w *ClientWrapper) Decrypt(payload []byte) {
	w.decryptor.Process(payload)
}

func (w *ClientWrapper) Encrypt(payload []byte) {
	w.encryptor.Process(payload)
}

func (w *ClientWrapper) Send(id uint16, version uint16, payload []byte) {
	w.Encrypt(payload)

	packetSize := 2 + 3 + 2 + len(payload)
	packet := make([]byte, packetSize)

	l := len(payload)

	binary.BigEndian.PutUint16(packet[0:2], id)

	packet[2] = byte((l >> 16) & 0xFF)
	packet[3] = byte((l >> 8) & 0xFF)
	packet[4] = byte(l & 0xFF)

	binary.BigEndian.PutUint16(packet[5:7], version)

	copy(packet[7:], payload)

	_, err := w.conn.Write(payload)

	if err != nil {
		fmt.Println("failed to write payload:", err)
	}

	fmt.Printf("sent %d (%d bytes)\n", id, l)
}

func (w *ClientWrapper) Conn() net.Conn {
	return w.conn
}
