package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"nhooyr.io/websocket"

	"l-battle/internal/battle"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

type Session struct {
	conn     *websocket.Conn
	manager  *battle.Manager
	codec    Codec
	logger   *slog.Logger
	roomID   string
	playerID string
}

func NewSession(conn *websocket.Conn, manager *battle.Manager, codec Codec, logger *slog.Logger) *Session {
	return &Session{
		conn:    conn,
		manager: manager,
		codec:   codec,
		logger:  logger,
	}
}

func (s *Session) Run(ctx context.Context) {
	defer func() {
		if s.roomID != "" && s.playerID != "" {
			s.manager.LeaveRoom(s.roomID, s.playerID)
		}
		_ = s.conn.Close(websocket.StatusNormalClosure, "session closed")
		s.logger.Info("websocket session closed", "roomId", s.roomID, "playerId", s.playerID)
	}()

	for {
		_, data, err := s.conn.Read(ctx)
		if err != nil {
			s.logger.Info("websocket read stopped", "error", err)
			return
		}

		packet, err := s.codec.Decode(data)
		if err != nil {
			s.writeError(ctx, "invalid packet")
			continue
		}

		if err := s.handlePacket(ctx, packet); err != nil {
			s.logger.Warn("handle websocket packet", "type", packet.Type, "error", err)
			s.writeError(ctx, err.Error())
		}
	}
}

func (s *Session) handlePacket(ctx context.Context, packet protocol.Packet) error {
	switch packet.Type {
	case protocol.PacketJoinRoom:
		var join protocol.JoinRoom
		if err := decodePayload(packet, &join); err != nil {
			return err
		}
		room, err := s.manager.JoinRoom(join.RoomID, join.PlayerID, join.HeroID, parseTeam(join.Team))
		if err != nil {
			return err
		}
		s.roomID = join.RoomID
		s.playerID = join.PlayerID
		room.AttachSession(join.PlayerID, s.sendSnapshot)
		return nil
	case protocol.PacketInput:
		var input protocol.PlayerInput
		if err := decodePayload(packet, &input); err != nil {
			return err
		}
		return s.manager.SubmitInput(packet.RoomID, packet.PlayerID, input)
	case protocol.PacketLeave:
		if packet.RoomID != "" && packet.PlayerID != "" {
			s.manager.LeaveRoom(packet.RoomID, packet.PlayerID)
		}
		_ = s.conn.Close(websocket.StatusNormalClosure, "left room")
		return nil
	default:
		return battle.ErrUnsupportedPacket
	}
}

func parseTeam(value string) world.Team {
	if world.Team(value) == world.TeamRed {
		return world.TeamRed
	}
	return world.TeamBlue
}

func (s *Session) sendSnapshot(snapshot protocol.Snapshot) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	payload, err := rawPayload(snapshot)
	if err != nil {
		s.logger.Warn("encode snapshot payload", "error", err)
		return
	}
	data, err := s.codec.Encode(protocol.Packet{
		Type:    protocol.PacketSnapshot,
		RoomID:  snapshot.RoomID,
		Seq:     snapshot.Tick,
		Payload: payload,
	})
	if err != nil {
		s.logger.Warn("encode snapshot packet", "error", err)
		return
	}
	if err := s.conn.Write(ctx, websocket.MessageText, data); err != nil {
		s.logger.Debug("write snapshot", "error", err)
	}
}

func (s *Session) writeError(ctx context.Context, message string) {
	payload, err := rawPayload(protocol.Error{Message: message})
	if err != nil {
		return
	}
	data, err := s.codec.Encode(protocol.Packet{
		Type:    protocol.PacketError,
		Payload: payload,
	})
	if err != nil {
		return
	}
	_ = s.conn.Write(ctx, websocket.MessageText, data)
}

func decodePayload(packet protocol.Packet, dst any) error {
	if packet.Payload == nil {
		return battle.ErrMissingPayload
	}
	return json.Unmarshal(*packet.Payload, dst)
}
