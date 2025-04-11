package network

import (
	"sync/atomic"
	"encoding/binary"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"io"
	"net"
	"log/slog"
)

type Server struct {
	address string
	ln      net.Listener
	quitch  chan struct{}

	dbm *database.Manager

	totalClients atomic.Int64
}

func NewServer(address string, dbm *database.Manager) *Server {
	return &Server{
		address: address,
		quitch:  make(chan struct{}),
		dbm:     dbm,
	}
}

func (s *Server) Serve() error {
	ln, err := net.Listen("tcp", s.address)

	if err != nil {
		return err
	}

	slog.Info("serving", "address", s.address)

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
			slog.Error("failed to accept client!", "err", err)
			continue
		}

		s.totalClients.Add(1)
		slog.Info("client connected", "total", s.totalClients.Load())

		wrapper := core.NewClientWrapper(conn)

		go s.handleClient(wrapper)
	}
}

func (s *Server) handleClient(wrapper *core.ClientWrapper) {
	defer func() {
		s.totalClients.Add(-1)
		slog.Info("client disconnected", "total", s.totalClients.Load())
		wrapper.Close()
	}()

	conn := wrapper.Conn()
	header := make([]byte, 7)

	for {
		_, err := conn.Read(header)

		if err != nil {
			if err != io.EOF {
				slog.Error("failed to read packet header!", "err", err)
			}

			break
		}

		packetId := binary.BigEndian.Uint16(header[:2])
		payloadSize := uint32(header[2])<<16 | uint32(header[3])<<8 | uint32(header[4])
		_ = binary.BigEndian.Uint16(header[5:])

		slog.Info("got packet", "id", packetId, "size", payloadSize)

		payload := make([]byte, payloadSize)

		_, err = conn.Read(payload)

		if err != nil {
			slog.Error("failed to read payload!", "err", err)
			break
		}

		wrapper.Decrypt(payload)

		factory, exists := ClientRegistry[packetId]

		if !exists {
			slog.Warn("got unknown packet", "id", packetId)
			continue
		}

		msg := factory()
		msg.Unmarshal(payload)
		msg.Process(wrapper, s.dbm)
	}
}
