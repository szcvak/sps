package messages

type KeepAliveOkMessage struct{}

func NewKeepAliveOkMessage() *KeepAliveOkMessage {
	return &KeepAliveOkMessage{}
}

func (k *KeepAliveOkMessage) PacketId() uint16 {
	return 20108
}

func (k *KeepAliveOkMessage) PacketVersion() uint16 {
	return 1
}

func (k *KeepAliveOkMessage) Marshal() []byte {
	return make([]byte, 0)
}
