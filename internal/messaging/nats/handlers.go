package nats

import (
	"encoding/json"
	"log/slog"

	natsgo "github.com/nats-io/nats.go"

	"l-battle/internal/battle"
)

const (
	SubjectRoomCreate = "battle.room.create"
	SubjectRoomClose  = "battle.room.close"
)

func RegisterHandlers(conn *natsgo.Conn, manager *battle.Manager, logger *slog.Logger) error {
	if _, err := conn.Subscribe(SubjectRoomCreate, func(msg *natsgo.Msg) {
		var command struct {
			RoomID string `json:"roomId"`
		}
		if err := json.Unmarshal(msg.Data, &command); err != nil {
			logger.Warn("decode room create command", "error", err)
			respond(msg, []byte(err.Error()))
			return
		}
		if _, err := manager.CreateRoom(command.RoomID); err != nil {
			respond(msg, []byte(err.Error()))
			return
		}
		respond(msg, []byte("ok"))
	}); err != nil {
		return err
	}

	if _, err := conn.Subscribe(SubjectRoomClose, func(msg *natsgo.Msg) {
		var command struct {
			RoomID string `json:"roomId"`
		}
		if err := json.Unmarshal(msg.Data, &command); err != nil {
			logger.Warn("decode room close command", "error", err)
			respond(msg, []byte(err.Error()))
			return
		}
		manager.CloseRoom(command.RoomID)
		respond(msg, []byte("ok"))
	}); err != nil {
		return err
	}

	return conn.Flush()
}

func respond(msg *natsgo.Msg, data []byte) {
	if msg.Reply == "" {
		return
	}
	_ = msg.Respond(data)
}
