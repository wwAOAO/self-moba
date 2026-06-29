package ws

import (
	"encoding/json"

	"l-battle/internal/protocol"
)

type Codec struct{}

func (Codec) Decode(data []byte) (protocol.Packet, error) {
	var packet protocol.Packet
	err := json.Unmarshal(data, &packet)
	return packet, err
}

func (Codec) Encode(packet protocol.Packet) ([]byte, error) {
	return json.Marshal(packet)
}

func rawPayload(value any) (*json.RawMessage, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	raw := json.RawMessage(data)
	return &raw, nil
}
