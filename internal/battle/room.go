package battle

import (
	"sync"

	"l-battle/internal/config"
	"l-battle/internal/messaging/jetstream"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

type SnapshotSender func(protocol.Snapshot)

type Room struct {
	id        string
	world     *world.World
	publisher *jetstream.Publisher
	heroes    *config.HeroStore
	skills    *config.SkillStore
	levels    *config.LevelConfig
	rewards   *config.RewardConfig
	loop      *Loop

	mu       sync.RWMutex
	sessions map[string]SnapshotSender
	closed   bool
}

func NewRoom(id string, publisher *jetstream.Publisher, heroes *config.HeroStore, skills *config.SkillStore, levels *config.LevelConfig, rewards *config.RewardConfig) *Room {
	room := &Room{
		id:        id,
		world:     world.NewWorld(heroes, levels, rewards),
		publisher: publisher,
		heroes:    heroes,
		skills:    skills,
		levels:    levels,
		rewards:   rewards,
		sessions:  make(map[string]SnapshotSender),
	}
	room.loop = NewLoop(room)
	return room
}

func (r *Room) Start() {
	r.loop.Start()
}

func (r *Room) Join(playerID string, hero config.HeroConfig, team world.Team) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.world.SpawnHero(playerID, hero, team)
}

func (r *Room) AttachSession(playerID string, sender SnapshotSender) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[playerID] = sender
}

func (r *Room) Leave(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, playerID)
	r.world.RemovePlayer(playerID)
}

func (r *Room) SubmitInput(playerID string, input protocol.PlayerInput) {
	r.loop.Submit(Input{PlayerID: playerID, Value: input})
}

func (r *Room) apply(inputs []Input, tick uint64) protocol.Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, input := range inputs {
		r.world.ApplyInput(input.PlayerID, input.Value, tick, r.skills, tickRate)
		_ = r.publisher.PublishInput(r.id, input.PlayerID, input.Value)
	}
	r.world.Tick(tick, tickRate)

	snapshot := BuildSnapshot(r.id, tick, r.world)
	_ = r.publisher.PublishSnapshot(snapshot)
	for _, sender := range r.sessions {
		sender(snapshot)
	}
	return snapshot
}

func (r *Room) Close() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	r.sessions = make(map[string]SnapshotSender)
	r.mu.Unlock()

	r.loop.Stop()
	_ = r.publisher.PublishRoomClosed(r.id)
}
