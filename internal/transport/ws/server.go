package ws

import (
	"log/slog"
	"net/http"

	"nhooyr.io/websocket"

	"l-battle/internal/battle"
)

type Server struct {
	manager *battle.Manager
	logger  *slog.Logger
	codec   Codec
}

func NewServer(manager *battle.Manager, logger *slog.Logger) *Server {
	return &Server{
		manager: manager,
		logger:  logger,
		codec:   Codec{},
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/configs/", noCache(http.StripPrefix("/configs/", http.FileServer(http.Dir("configs")))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		setNoCache(w)
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "web/index.html")
			return
		}
		http.FileServer(http.Dir("web")).ServeHTTP(w, r)
	})
	return mux
}

func noCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setNoCache(w)
		next.ServeHTTP(w, r)
	})
}

func setNoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		s.logger.Warn("accept websocket", "error", err)
		return
	}

	s.logger.Info("websocket accepted", "remote", r.RemoteAddr)
	session := NewSession(conn, s.manager, s.codec, s.logger)
	session.Run(r.Context())
}
