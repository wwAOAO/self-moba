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
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	addr := env("BATTLE_ADDR", ":6969")
	natsURL := env("NATS_URL", natsgo.DefaultURL)
	heroConfigPath := env("HERO_CONFIG", "configs/heroes.json")
	skillConfigPath := env("SKILL_CONFIG", "configs/skills.json")

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

	manager := battle.NewManager(js, heroes, skills)
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
