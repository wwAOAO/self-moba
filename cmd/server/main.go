package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	natsgo "github.com/nats-io/nats.go"

	"l-battle/internal/battle"
	"l-battle/internal/config"
	"l-battle/internal/messaging/jetstream"
	messagingnats "l-battle/internal/messaging/nats"
	"l-battle/internal/transport/ws"
	_ "l-battle/internal/world/heroes/archer"
	_ "l-battle/internal/world/heroes/blade"
	_ "l-battle/internal/world/heroes/mage"
	_ "l-battle/internal/world/heroes/sword"
	_ "l-battle/internal/world/heroes/tank"
	_ "l-battle/internal/world/heroes/warrior"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	addr := env("BATTLE_ADDR", ":6969")
	natsURL := env("NATS_URL", natsgo.DefaultURL)
	heroConfigPath := env("HERO_CONFIG", "configs/heroes")
	skillConfigPath := env("SKILL_CONFIG", "configs/skills")
	levelConfigPath := env("LEVEL_CONFIG", "configs/levels.json")
	rewardConfigPath := env("REWARD_CONFIG", "configs/rewards.json")
	equipmentConfigPath := env("EQUIPMENT_CONFIG", "configs/equipment")

	heroes, err := config.LoadHeroes(heroConfigPath)
	if err != nil {
		logger.Error("load hero config", "path", heroConfigPath, "error", err)
		os.Exit(1)
	}
	logger.Info("hero config loaded", "path", heroConfigPath, "count", heroes.Count())
	skills, err := config.LoadSkills(skillConfigPath)
	if err != nil {
		logger.Error("load skill config", "path", skillConfigPath, "error", err)
		os.Exit(1)
	}
	if err := config.ValidateHeroSkills(heroes, skills); err != nil {
		logger.Error("validate hero skills", "error", err)
		os.Exit(1)
	}
	logger.Info("skill config loaded", "path", skillConfigPath, "count", skills.Count())
	levels, err := config.LoadLevels(levelConfigPath)
	if err != nil {
		logger.Error("load level config", "path", levelConfigPath, "error", err)
		os.Exit(1)
	}
	logger.Info("level config loaded", "path", levelConfigPath, "maxLevel", levels.MaxLevel, "totalExp", levels.TotalExp)
	rewards, err := config.LoadRewards(rewardConfigPath)
	if err != nil {
		logger.Error("load reward config", "path", rewardConfigPath, "error", err)
		os.Exit(1)
	}
	logger.Info("reward config loaded", "path", rewardConfigPath, "minionKinds", len(rewards.Minion.KillExp), "jungleKinds", len(rewards.Jungle.KillExp), "epicKinds", len(rewards.Epic), "structureKinds", len(rewards.Structure.TeamExp), "jungleScalingMax", rewards.JungleScaling.MaxMultiplier)
	equipment, err := config.LoadEquipment(equipmentConfigPath)
	if err != nil {
		logger.Error("load equipment config", "path", equipmentConfigPath, "error", err)
		os.Exit(1)
	}
	if err := config.ValidateGameConfig(config.GameConfig{
		Heroes:    heroes,
		Skills:    skills,
		Levels:    levels,
		Rewards:   rewards,
		Equipment: equipment,
	}); err != nil {
		logger.Error("validate game config", "error", err)
		os.Exit(1)
	}
	logger.Info("equipment config loaded", "path", equipmentConfigPath, "count", equipment.Count())

	natsClient, err := messagingnats.Connect(natsURL, "battle-server")
	if err != nil {
		logger.Error("connect nats", "error", err)
		os.Exit(1)
	}
	defer natsClient.Close()

	js, err := jetstream.NewPublisher(natsClient.Conn())
	if err != nil {
		logger.Error("create jetstream publisher", "error", err)
		os.Exit(1)
	}
	if err := jetstream.EnsureStreams(js.Context()); err != nil {
		logger.Error("ensure jetstream streams", "error", err)
		os.Exit(1)
	}

	manager := battle.NewManager(js, heroes, skills, levels, rewards, equipment)
	if err := messagingnats.RegisterHandlers(natsClient.Conn(), manager, logger); err != nil {
		logger.Error("register nats handlers", "error", err)
		os.Exit(1)
	}

	wsServer := ws.NewServer(manager, logger)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           wsServer.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("battle server listening", "addr", addr, "nats", natsURL)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	manager.Close()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown http server", "error", err)
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
