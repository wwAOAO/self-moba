package model

type EntityKind string

const (
	EntityKindPlayer       EntityKind = "player"
	EntityKindEnemyHero    EntityKind = "enemy_hero"
	EntityKindSiegeMinion  EntityKind = "siege_minion"
	EntityKindSuperMinion  EntityKind = "super_minion"
	EntityKindMeleeMinion  EntityKind = "melee_minion"
	EntityKindRangedMinion EntityKind = "ranged_minion"
	EntityKindBlueBuff     EntityKind = "blue_buff"
	EntityKindRedBuff      EntityKind = "red_buff"
	EntityKindGromp        EntityKind = "gromp"
	EntityKindRaptor       EntityKind = "raptor"
	EntityKindMurkWolf     EntityKind = "murk_wolf"
	EntityKindKrugCamp     EntityKind = "krug_camp"
	EntityKindBaronNashor  EntityKind = "baron_nashor"
	EntityKindTower        EntityKind = "tower"
	EntityKindCrystal      EntityKind = "crystal"
	EntityKindBarracks     EntityKind = "barracks"
	EntityKindFountain     EntityKind = "fountain"
	EntityKindDummy        EntityKind = "dummy"
	EntityKindFruit        EntityKind = "fruit"
	EntityKindWard         EntityKind = "ward"
)

type Team string

const (
	TeamBlue    Team = "blue"
	TeamRed     Team = "red"
	TeamNeutral Team = "neutral"
)

func IsHeroKind(kind EntityKind) bool {
	return kind == EntityKindPlayer || kind == EntityKindEnemyHero
}

func IsMinionKind(kind EntityKind) bool {
	switch kind {
	case EntityKindMeleeMinion, EntityKindRangedMinion, EntityKindSiegeMinion, EntityKindSuperMinion:
		return true
	default:
		return false
	}
}

func IsMonsterKind(kind EntityKind, team Team) bool {
	if team != TeamNeutral {
		return false
	}
	switch kind {
	case EntityKindPlayer, EntityKindEnemyHero, EntityKindTower, EntityKindBarracks, EntityKindCrystal, EntityKindDummy, EntityKindFruit, EntityKindWard:
		return false
	default:
		return true
	}
}
