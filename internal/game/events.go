package game

import (
	"encoding/json"
	"influence_game/internal/realtime"
	"time"

	"github.com/rs/zerolog/log"
)

type ServerEvent struct {
	EventType string           `json:"eventType"`
	GameID    string           `json:"gameID"`
	Timestamp time.Time        `json:"timestamp"`
	GameState *PublicGameState `json:"state,omitempty"`
	Payload   map[string]any   `json:"payload,omitempty"`
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
