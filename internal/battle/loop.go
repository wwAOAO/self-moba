package battle

import (
	"sync"
	"time"
)

const tickRate = 20

type Loop struct {
	room   *Room
	inputs chan Input
	done   chan struct{}
	once   sync.Once
}

func NewLoop(room *Room) *Loop {
	return &Loop{
		room:   room,
		inputs: make(chan Input, 256),
		done:   make(chan struct{}),
	}
}

func (l *Loop) Start() {
	go l.run()
}

func (l *Loop) Submit(input Input) {
	select {
	case l.inputs <- input:
	default:
	}
}

func (l *Loop) Stop() {
	l.once.Do(func() {
		close(l.done)
	})
}

func (l *Loop) run() {
	ticker := time.NewTicker(time.Second / tickRate)
	defer ticker.Stop()

	var tick uint64
	pending := make([]Input, 0, 64)

	for {
		select {
		case input := <-l.inputs:
			pending = append(pending, input)
		case <-ticker.C:
			tick++
			batch := pending
			pending = make([]Input, 0, 64)
			l.room.apply(batch, tick)
		case <-l.done:
			return
		}
	}
}
