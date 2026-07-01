package jetstream

import natsgo "github.com/nats-io/nats.go"

const BattleEventsStream = "BATTLE_EVENTS"
const BattleEventsMaxBytes = 5 << 30

func EnsureStreams(js natsgo.JetStreamContext) error {
	_, err := js.AddStream(&natsgo.StreamConfig{
		Name:      BattleEventsStream,
		Subjects:  []string{SubjectRoomEventPrefix + ".>"},
		Retention: natsgo.LimitsPolicy,
		Storage:   natsgo.FileStorage,
		MaxBytes:  BattleEventsMaxBytes,
		Discard:   natsgo.DiscardOld,
	})
	if err == nil {
		return nil
	}
	_, err = js.UpdateStream(&natsgo.StreamConfig{
		Name:      BattleEventsStream,
		Subjects:  []string{SubjectRoomEventPrefix + ".>"},
		Retention: natsgo.LimitsPolicy,
		Storage:   natsgo.FileStorage,
		MaxBytes:  BattleEventsMaxBytes,
		Discard:   natsgo.DiscardOld,
	})
	return err
}
