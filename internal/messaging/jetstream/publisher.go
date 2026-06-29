package jetstream

import (
	"encoding/json"

	natsgo "github.com/nats-io/nats.go"

	"l-battle/internal/protocol"
)

const (
	SubjectRoomEventPrefix = "battle.room"
)

type Publisher struct {
	js natsgo.JetStreamContext
}

func NewPublisher(conn *natsgo.Conn) (*Publisher, error) {
	js, err := conn.JetStream()
	if err != nil {
		return nil, err
	}
	return &Publisher{js: js}, nil
}

func (p *Publisher) Context() natsgo.JetStreamContext {
	return p.js
}

func (p *Publisher) PublishInput(roomID string, playerID string, input protocol.PlayerInput) error {
	return p.publish(subject(roomID, "input"), map[string]any{
		"roomId":   roomID,
		"playerId": playerID,
		"input":    input,
	})
}

func (p *Publisher) PublishSnapshot(snapshot protocol.Snapshot) error {
	return p.publish(subject(snapshot.RoomID, "snapshot"), snapshot)
}

func (p *Publisher) PublishRoomClosed(roomID string) error {
	return p.publish(subject(roomID, "closed"), map[string]string{"roomId": roomID})
}

func (p *Publisher) publish(subject string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = p.js.Publish(subject, data)
	return err
}

func subject(roomID string, event string) string {
	return SubjectRoomEventPrefix + "." + roomID + "." + event
}
