package game

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type ActionType string

const (
	ActionStart   ActionType = "start"
	ActionEndTurn ActionType = "end_turn"

	ActionIncome     ActionType = "income"
	ActionForeignAid ActionType = "foreign_aid"
	ActionCoup       ActionType = "coup"

	ActionTax         ActionType = "tax"
	ActionAssassinate ActionType = "assassinate"
	ActionSteal       ActionType = "steal"
	ActionExchange    ActionType = "exchange"

	ActionBlockForeignAid  ActionType = "block_foreign_aid"
	ActionBlockAssassinate ActionType = "block_assassinate"
	ActionBlockSteal       ActionType = "block_steal"
)

var (
	ErrGameNotFound   = errors.New("game_not_found")
	ErrAlreadyStarted = errors.New("game_already_started")
	ErrNotStarted     = errors.New("game_not_started")
	ErrInvalidAction  = errors.New("invalid_action")
)

type Influence struct {
	Role     string       `json:"role"`
	Revealed bool         `json:"revealed"`
	Actions  []ActionType `json:"actions"`
}

type Player struct {
	ID         string      `json:"id"`
	Nickname   string      `json:"nickname"`
	Coins      int         `json:"coins"`
	Alive      bool        `json:"alive"`
	Influences []Influence `json:"influences"`
}

type Game struct {
	ID        string
	CreatedAt time.Time
	Players   []*Player
	TurnIndex int
	Started   bool
	Finished  bool
}

/*
⚠️ Warning:
- this is a public representation of the influence, so it should not contain the role if it is not revealed
*/
type PublicInfluence struct {
	Role     *string `json:"role,omitempty"`
	Revealed bool    `json:"revealed"`
}

type PublicPlayer struct {
	ID         string            `json:"id"`
	Nickname   string            `json:"nickname"`
	Coins      int               `json:"coins"`
	Alive      bool              `json:"alive"`
	Influences []PublicInfluence `json:"influences"`
}

type PublicState struct {
	GameID    string         `json:"gameID"`
	Started   bool           `json:"started"`
	Finished  bool           `json:"finished"`
	TurnIndex int            `json:"turnIndex"`
	Players   []PublicPlayer `json:"players"`
}

func newGame() *Game {
	return &Game{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		Players:   []*Player{},
	}
}

func (game *Game) getPublicState() *PublicState {
	publicPlayers := make([]PublicPlayer, 0, len(game.Players))

	for _, player := range game.Players {
		publicPlayers = append(publicPlayers, PublicPlayer{
			ID: player.ID, Nickname: player.Nickname, Coins: player.Coins, Alive: player.Alive,
		})
	}

	return &PublicState{
		GameID: game.ID, Started: game.Started, Finished: game.Finished,
		TurnIndex: game.TurnIndex, Players: publicPlayers,
	}
}

func (g *Game) handleAction(action ActionType, body json.RawMessage) error {
	switch action {
	case ActionStart:
		if g.Started {
			return ErrAlreadyStarted
		}
		if len(g.Players) < 2 {
			return errors.New("need_at_least_two_players")
		}
		g.Started = true
		for _, p := range g.Players {
			p.Coins = 2
			p.Alive = true
		}
		return nil
	case ActionIncome:
		if !g.Started {
			return ErrNotStarted
		}
		cur := g.Players[g.TurnIndex%len(g.Players)]
		if !cur.Alive {
			g.TurnIndex = (g.TurnIndex + 1) % len(g.Players)
			return nil
		}
		cur.Coins++
		g.TurnIndex = (g.TurnIndex + 1) % len(g.Players)
		return nil

		// TODO: Implement other actions
	default:
		return ErrInvalidAction
	}
}

type Store struct {
	mu    sync.RWMutex
	games map[string]*Game

	redis *redis.Client
}

func NewStore(redisClient *redis.Client) *Store {
	return &Store{
		games: make(map[string]*Game),
		redis: redisClient,
	}
}

func (store *Store) NewGame() (*Game, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	newGame := newGame()

	store.games[newGame.ID] = newGame

	serializedGame, err := json.Marshal(newGame)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize game.")
		return nil, err
	}

	redisKey := "game:" + newGame.ID
	ctx := context.Background()

	if err := store.redis.Set(ctx, redisKey, serializedGame, 0).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to save game to Redis.")
		return nil, err
	}

	return newGame, nil
}

func (s *Store) Get(id string) *Game {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.games[id]
}

func (s *Store) Join(gameID, nickname string) (*Game, *Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.games[gameID]
	if !ok {
		return nil, nil, ErrGameNotFound
	}
	if g.Started {
		return nil, nil, ErrAlreadyStarted
	}
	p := &Player{
		ID:       uuid.NewString(),
		Nickname: nickname,
		Coins:    0,
		Alive:    false,
	}
	g.Players = append(g.Players, p)
	return g, p, nil
}
