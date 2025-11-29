package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
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
	ErrGameNotFound        = errors.New("game_not_found")
	ErrAlreadyStarted      = errors.New("game_already_started")
	ErrNotStarted          = errors.New("game_not_started")
	ErrInvalidAction       = errors.New("invalid_action")
	ErrPlayerAlreadyJoined = errors.New("Player already joined with this nickname")
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
	AdminID   string
	JoinCode  string
	Players   []*Player
	TurnIndex int
	Started   bool
	Finished  bool
}

type PlayerSession struct {
	PlayerID string `json:"playerId"`
	GameID   string `json:"gameId"`
}

/*
⚠️ Warning:
- this is a public representation of the influence, so it should not contain the role if it is not revealed
*/
type PublicInfluence struct {
	Role     *string `json:"role,omitempty"`
	Revealed bool    `json:"revealed"`
}

type PlayerPublicInfo struct {
	ID         string            `json:"id"`
	Nickname   string            `json:"nickname"`
	Coins      int               `json:"coins"`
	Alive      bool              `json:"alive"`
	Influences []PublicInfluence `json:"influences"`
}

type PublicGameState struct {
	GameID    string             `json:"gameID"`
	JoinCode  string             `json:"joinCode"`
	Started   bool               `json:"started"`
	AdminID   string             `json:"adminID"`
	Finished  bool               `json:"finished"`
	TurnIndex int                `json:"turnIndex"`
	Players   []PlayerPublicInfo `json:"players"`
}

func (game *Game) GetPublicGameState() *PublicGameState {
	playersPublicInfo := make([]PlayerPublicInfo, 0, len(game.Players))

	for _, player := range game.Players {
		playersPublicInfo = append(playersPublicInfo, getPublicPlayerInfo(player))
		//  TODO: Retrieve user's own influences
	}

	return &PublicGameState{
		GameID:    game.ID,
		JoinCode:  game.JoinCode,
		Started:   game.Started,
		Finished:  game.Finished,
		TurnIndex: game.TurnIndex,
		Players:   playersPublicInfo,
		AdminID:   game.AdminID,
	}
}

func getPublicPlayerInfo(player *Player) PlayerPublicInfo {
	return PlayerPublicInfo{
		ID:         player.ID,
		Nickname:   player.Nickname,
		Coins:      player.Coins,
		Alive:      player.Alive,
		Influences: []PublicInfluence{}, // TODO: Add influences public info
	}
}

// func (g *Game) HandleAction(action ActionType, body json.RawMessage) error {
// 	switch action {
// 	case ActionStart:
// 		if g.Started {
// 			return ErrAlreadyStarted
// 		}
// 		if len(g.Players) < 2 {
// 			return errors.New("need_at_least_two_players")
// 		}
// 		g.Started = true
// 		for _, p := range g.Players {
// 			p.Coins = 2
// 			p.Alive = true
// 		}
// 		return nil
// 	case ActionIncome:
// 		if !g.Started {
// 			return ErrNotStarted
// 		}
// 		cur := g.Players[g.TurnIndex%len(g.Players)]
// 		if !cur.Alive {
// 			g.TurnIndex = (g.TurnIndex + 1) % len(g.Players)
// 			return nil
// 		}
// 		cur.Coins++
// 		g.TurnIndex = (g.TurnIndex + 1) % len(g.Players)
// 		return nil

// 		// TODO: Implement other actions
// 	default:
// 		return ErrInvalidAction
// 	}
// }

type Store struct {
	redis *redis.Client
}

func NewStore(redisClient *redis.Client) *Store {
	return &Store{
		redis: redisClient,
	}
}

type OnboardingResult struct {
	Game   *PublicGameState `json:"game"`
	Player *Player          `json:"player"`
	Token  string           `json:"token"`
}

func (store *Store) CreateGameRoom(adminNickname string) (*OnboardingResult, error) {
	adminPlayer := buildNewPlayer(adminNickname)

	newGame, err := store.buildNewGame(adminPlayer)
	if err != nil {
		return nil, err
	}

	if err := store.saveGameToRedis(newGame); err != nil {
		ctx := context.Background()
		_ = store.redis.Del(ctx, "joincode:"+newGame.JoinCode).Err()
		return nil, err
	}

	sessionToken, err := store.CreatePlayerSession(newGame.ID, adminPlayer.ID)
	if err != nil {
		return nil, err
	}

	publicState := newGame.GetPublicGameState()

	return &OnboardingResult{
		Game:   publicState,
		Player: adminPlayer,
		Token:  sessionToken,
	}, nil
}

func buildNewPlayer(nickname string) *Player {
	return &Player{
		ID:         uuid.NewString(),
		Nickname:   nickname,
		Coins:      2,
		Alive:      true,
		Influences: []Influence{},
	}
}

func (store *Store) buildNewGame(adminPlayer *Player) (*Game, error) {
	gameID := uuid.NewString()

	joinCode, err := store.reserveJoinCode(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique join code: %w", err)
	}

	game := &Game{
		ID:        gameID,
		CreatedAt: time.Now(),
		Players:   []*Player{adminPlayer},
		JoinCode:  joinCode,
		AdminID:   adminPlayer.ID,
		TurnIndex: 0,
		Started:   false,
		Finished:  false,
	}

	return game, nil
}

const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomJoinCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (store *Store) reserveJoinCode(gameID string) (string, error) {
	ctx := context.Background()

	for {
		code := randomJoinCode()
		key := "joincode:" + code

		ok, err := store.redis.SetNX(ctx, key, gameID, JoinCodeTTL).Result()
		if err != nil {
			return "", err
		}

		if ok {
			return code, nil
		}
	}
}

func (store *Store) saveGameToRedis(newGame *Game) error {
	serializedGame, err := json.Marshal(newGame)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize game.")
		return err
	}

	redisKey := "game:" + newGame.ID
	ctx := context.Background()

	if err := store.redis.Set(ctx, redisKey, serializedGame, 0).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to save game to Redis.")
		return err
	}

	return nil
}

func (store *Store) CreatePlayerSession(gameID string, playerID string) (string, error) {
	if store.redis == nil {
		return "", errors.New("redis_not_configured")
	}

	ctx := context.Background()

	sessionToken := uuid.NewString()

	session := PlayerSession{
		PlayerID: playerID,
		GameID:   gameID,
	}

	data, err := json.Marshal(session)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize session.")
		return "", err
	}

	redisKey := "session:" + sessionToken

	err = store.redis.Set(ctx, redisKey, data, SessionDuration).Err()
	if err != nil {
		log.Error().Err(err).Msg("Failed to save session to Redis.")
		return "", err
	}

	return sessionToken, nil
}

func (store *Store) Join(joinCode, nickname string) (*OnboardingResult, error) {
	ctx := context.Background()

	joinKey := "joincode:" + joinCode
	gameID, err := store.redis.Get(ctx, joinKey).Result()
	if err == redis.Nil {
		return nil, ErrGameNotFound
	}
	if err != nil {
		return nil, err
	}

	gameKey := "game:" + gameID

	var joinedPlayer *Player
	var finalGame Game

	for {
		err := store.redis.Watch(ctx, func(tx *redis.Tx) error {
			gameJSON, err := tx.Get(ctx, gameKey).Bytes()
			if err != nil {
				log.Error().Err(err).Msg("Failed to get game from Redis.")
				return err
			}

			if err := json.Unmarshal(gameJSON, &finalGame); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal game from Redis.")
				return err
			}

			if finalGame.Started {
				log.Error().Msg("Game already started.")
				return ErrAlreadyStarted
			}

			for _, p := range finalGame.Players {
				if p.Nickname == nickname {
					log.Error().Err(err).Msg("Player already joined with this nickname.")
					return ErrPlayerAlreadyJoined
				}
			}

			joinedPlayer = buildNewPlayer(nickname)
			finalGame.Players = append(finalGame.Players, joinedPlayer)

			updatedJSON, _ := json.Marshal(finalGame)

			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, gameKey, updatedJSON, 0)
				return nil
			})

			return err
		}, gameKey)

		if err == redis.TxFailedErr {
			continue
		}

		if err != nil {
			return nil, err
		}

		break
	}

	sessionToken, err := store.CreatePlayerSession(gameID, joinedPlayer.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create player session.")
		return nil, err
	}

	return &OnboardingResult{
		Game:   finalGame.GetPublicGameState(),
		Player: joinedPlayer,
		Token:  sessionToken,
	}, nil
}
