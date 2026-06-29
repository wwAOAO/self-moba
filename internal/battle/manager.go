package battle

import (
	"errors"
	"fmt"
	"sync"

	"l-battle/internal/config"
	"l-battle/internal/messaging/jetstream"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

var (
	ErrMissingPayload    = errors.New("missing payload")
	ErrUnsupportedPacket = errors.New("unsupported packet")
)

type Manager struct {
	mu        sync.RWMutex
	rooms     map[string]*Room
	publisher *jetstream.Publisher
	heroes    *config.HeroStore
	skills    *config.SkillStore
}

func NewManager(publisher *jetstream.Publisher, heroes *config.HeroStore, skills *config.SkillStore) *Manager {
	return &Manager{
		rooms:     make(map[string]*Room),
		publisher: publisher,
		heroes:    heroes,
		skills:    skills,
	}
}

func (m *Manager) CreateRoom(roomID string) (*Room, error) {
	if roomID == "" {
		return nil, fmt.Errorf("room id is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if room, ok := m.rooms[roomID]; ok {
		return room, nil
	}

	room := NewRoom(roomID, m.publisher, m.skills)
	m.rooms[roomID] = room
	room.Start()
	return room, nil
}

func (m *Manager) JoinRoom(roomID string, playerID string, heroID string, team world.Team) (*Room, error) {
	if playerID == "" {
		return nil, fmt.Errorf("player id is required")
	}
	hero, ok := m.heroes.Get(heroID)
	if !ok {
		return nil, fmt.Errorf("hero %s not found", heroID)
	}
	room, err := m.CreateRoom(roomID)
	if err != nil {
		return nil, err
	}
	room.Join(playerID, hero, team)
	return room, nil
}

func (m *Manager) SubmitInput(roomID string, playerID string, input protocol.PlayerInput) error {
	m.mu.RLock()
	room := m.rooms[roomID]
	m.mu.RUnlock()
	if room == nil {
		return fmt.Errorf("room %s not found", roomID)
	}
	room.SubmitInput(playerID, input)
	return nil
}

func (m *Manager) LeaveRoom(roomID string, playerID string) {
	m.mu.RLock()
	room := m.rooms[roomID]
	m.mu.RUnlock()
	if room != nil {
		room.Leave(playerID)
	}
}

func (m *Manager) CloseRoom(roomID string) {
	m.mu.Lock()
	room := m.rooms[roomID]
	delete(m.rooms, roomID)
	m.mu.Unlock()

	if room != nil {
		room.Close()
	}
}

func (m *Manager) Close() {
	m.mu.Lock()
	rooms := m.rooms
	m.rooms = make(map[string]*Room)
	m.mu.Unlock()

	for _, room := range rooms {
		room.Close()
	}
}
