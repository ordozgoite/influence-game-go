package game

import (
	"encoding/json"
	"influence_game/internal/realtime"
	"time"

	"github.com/rs/zerolog/log"
)

type ServerEvent struct {
	EventType string           `json:"eventType"`         // "player_joined",
	GameID    string           `json:"gameID"`            // pra debug / cliente
	Timestamp time.Time        `json:"timestamp"`         // quando o evento foi emitido
	GameState *PublicGameState `json:"state,omitempty"`   // snapshot opcional
	Payload   map[string]any   `json:"payload,omitempty"` // dados extras da ação
}

func BroadcastEvent(
	state *PublicGameState,
	eventType string,
	payload map[string]any,
) {
	if state == nil {
		return
	}

	ev := ServerEvent{
		EventType: eventType,
		GameID:    state.GameID,
		Timestamp: time.Now().UTC(),
		GameState: state,
		Payload:   payload,
	}

	data, err := json.Marshal(ev)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal event.")
		return
	}

	realtime.Manager.Broadcast(state.GameID, data)
}
