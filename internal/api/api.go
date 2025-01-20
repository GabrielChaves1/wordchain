package api

import (
	"GabrielChaves1/game/internal/models"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
)

type apiHandler struct {
	r           *chi.Mux
	upgrader    websocket.Upgrader
	subscribers map[string]map[models.Player]context.CancelFunc
	mu          *sync.Mutex
}

func (h apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

func NewHandler() http.Handler {
	a := apiHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		subscribers: make(map[string]map[models.Player]context.CancelFunc),
		mu:          &sync.Mutex{},
	}

	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.Recoverer,
		middleware.Logger,
	)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/subscribe/{room_id}", a.handleSubscribeToRoom)
	r.Post("/create_room", a.handleCreateRoom)

	a.r = r
	return a
}

func (h apiHandler) handleSubscribeToRoom(w http.ResponseWriter, r *http.Request) {
	rawRoomID := chi.URLParam(r, "room_id")

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("Failed to upgrade connection", "error", err)
		http.Error(w, "Failed to upgrade to ws connection", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())

	defer func() {
		cancel()
		conn.Close()
	}()

	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				cancel()
				break
			}
		}
	}()

	player := &models.Player{
		Name:       "Droyen",
		Connection: conn,
		IsReady:    false,
	}

	h.mu.Lock()
	if _, ok := h.subscribers[rawRoomID]; !ok {
		h.subscribers[rawRoomID] = make(map[models.Player]context.CancelFunc)
	}

	h.subscribers[rawRoomID][*player] = cancel
	h.mu.Unlock()

	<-ctx.Done()

	fmt.Printf("User [%s] disconnected\n", player.Name)

	h.mu.Lock()
	delete(h.subscribers[rawRoomID], *player)

	if len(h.subscribers[rawRoomID]) <= 1 {
		delete(h.subscribers, rawRoomID)
	}
	h.mu.Unlock()
}

func (a apiHandler) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	
}