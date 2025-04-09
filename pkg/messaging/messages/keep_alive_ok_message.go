package messages

type KeepAliveOkMessage struct{}

func NewKeepAliveOkMessage() *KeepAliveOkMessage {
	return &KeepAliveOkMessage{}
}

func (l *KeepAliveOkMessage) PacketId() uint16 {
	return 20108
}

func (l *KeepAliveOkMessage) PacketVersion() uint16 {
	return 1
}

func (l *KeepAliveOkMessage) Marshal() []byte {
	return make([]byte, 0)
}
