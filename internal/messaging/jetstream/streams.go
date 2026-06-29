package jetstream

import natsgo "github.com/nats-io/nats.go"

const BattleEventsStream = "BATTLE_EVENTS"

func EnsureStreams(js natsgo.JetStreamContext) error {
	_, err := js.AddStream(&natsgo.StreamConfig{
		Name:      BattleEventsStream,
		Subjects:  []string{SubjectRoomEventPrefix + ".>"},
		Retention: natsgo.LimitsPolicy,
		Storage:   natsgo.FileStorage,
	})
	if err == nil {
		return nil
	}
	_, err = js.UpdateStream(&natsgo.StreamConfig{
		Name:      BattleEventsStream,
		Subjects:  []string{SubjectRoomEventPrefix + ".>"},
		Retention: natsgo.LimitsPolicy,
		Storage:   natsgo.FileStorage,
	})
	return err
}
