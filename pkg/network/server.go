package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/szcvak/sps/pkg/core"
)

type Server struct {
	address string
	ln      net.Listener
	quitch  chan struct{}
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
		quitch:  make(chan struct{}),
	}
}

func (s *Server) Serve() error {
	ln, err := net.Listen("tcp", s.address)

	if err != nil {
		return fmt.Errorf("failed to start the server: %w", err)
	}

	fmt.Println("started serving on", s.address)

	defer ln.Close()
	s.ln = ln

	go s.accept()

	<-s.quitch

	return nil
}

func (s *Server) accept() {
	for {
		conn, err := s.ln.Accept()

		if err != nil {
			fmt.Println("server accept error:", err)
			continue
		}

		fmt.Println("received connection from", conn.RemoteAddr())

		wrapper := core.NewClientWrapper(conn)

		go s.handleClient(wrapper)
	}
}

func (s *Server) handleClient(wrapper *core.ClientWrapper) {
	defer wrapper.Close()

	conn := wrapper.Conn()
	header := make([]byte, 7)

	for {
		_, err := conn.Read(header)

		if err != nil {
			if err == io.EOF {
				fmt.Printf("client disconnected (%s)\n", conn.RemoteAddr())
			} else {
				fmt.Println("failed to read header: ", err)
			}

			break
		}

		packetId := binary.BigEndian.Uint16(header[:2])
		payloadSize := uint32(header[2])<<16 | uint32(header[3])<<8 | uint32(header[4])
		_ = binary.BigEndian.Uint16(header[5:])

		fmt.Printf("received packet %d (%d bytes) (%s)\n", packetId, payloadSize, conn.RemoteAddr())

		payload := make([]byte, payloadSize)

		_, err2 := conn.Read(payload)

		if err2 != nil {
			fmt.Println("failed to read header:", err2)
			break
		}

		wrapper.Decrypt(payload)

		factory, exists := ClientRegistry[packetId]

		if !exists {
			fmt.Println("no processor for id", packetId)
			continue
		}

		msg := factory()
		msg.Unmarshal(payload)
		msg.Process(wrapper)
	}
}
