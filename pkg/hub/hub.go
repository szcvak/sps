package hub

import (
	"log/slog"
	"sync"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/messaging"
)

type Hub struct {
	mu            sync.RWMutex
	ClientsByAID  map[int64]map[*core.ClientWrapper]bool
	clientsGlobal map[*core.ClientWrapper]bool
}

var globalHub *Hub
var hubOnce sync.Once

func InitHub() {
	hubOnce.Do(func() {
		globalHub = &Hub{
			ClientsByAID:  make(map[int64]map[*core.ClientWrapper]bool),
			clientsGlobal: make(map[*core.ClientWrapper]bool),
		}
	})
}

func GetHub() *Hub {
	if globalHub == nil {
		panic("hub not initialized")
	}

	return globalHub
}

func (h *Hub) AddClient(client *core.ClientWrapper) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clientsGlobal[client] = true

	if client.Player != nil && client.Player.AllianceId != nil {
		allianceId := *client.Player.AllianceId

		if _, ok := h.ClientsByAID[allianceId]; !ok {
			h.ClientsByAID[allianceId] = make(map[*core.ClientWrapper]bool)
		}

		h.ClientsByAID[allianceId][client] = true
	} else {
		slog.Warn("will not add player to hub", "allianceId", client.Player.AllianceId)
	}
}

func (h *Hub) RemoveClient(client *core.ClientWrapper) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clientsGlobal, client)

	if client.Player != nil && client.Player.AllianceId != nil {
		allianceId := *client.Player.AllianceId

		if allianceClients, ok := h.ClientsByAID[allianceId]; ok {
			delete(allianceClients, client)

			if len(allianceClients) == 0 {
				delete(h.ClientsByAID, allianceId)
			}
		}
	}
}

func (h *Hub) UpdateAllianceMembership(client *core.ClientWrapper, oldAllianceId *int64, newAllianceId *int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if oldAllianceId != nil {
		if allianceClients, ok := h.ClientsByAID[*oldAllianceId]; ok {
			delete(allianceClients, client)

			if len(allianceClients) == 0 {
				delete(h.ClientsByAID, *oldAllianceId)
			}
		}
	}

	if newAllianceId != nil {
		if _, ok := h.ClientsByAID[*newAllianceId]; !ok {
			h.ClientsByAID[*newAllianceId] = make(map[*core.ClientWrapper]bool)
		}

		h.ClientsByAID[*newAllianceId][client] = true
	}
}

func (h *Hub) BroadcastToAlliance(allianceId int64, message messaging.ServerMessage) {
	h.mu.RLock()

	allianceClients, ok := h.ClientsByAID[allianceId]

	if !ok {
		h.mu.RUnlock()
		return
	}

	clientsToSend := make([]*core.ClientWrapper, 0, len(allianceClients))

	for client := range allianceClients {
		slog.Info("broadcasted message", "recipient", client.Conn().RemoteAddr())
		clientsToSend = append(clientsToSend, client)
	}

	h.mu.RUnlock()

	packetId := message.PacketId()
	version := message.PacketVersion()
	payload := message.Marshal()

	for _, client := range clientsToSend {
		client.Send(packetId, version, payload)
	}
}
