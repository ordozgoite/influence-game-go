package game

import "time"

type WSMessageType string

const (
	WSActionDeclared   WSMessageType = "action_declared"
	WSActionResolved   WSMessageType = "action_resolved"
	WSActionCanceled   WSMessageType = "action_canceled"
	WSActionBlocked    WSMessageType = "action_blocked"
	WSActionContested  WSMessageType = "action_contested"
	WSGameStateUpdated WSMessageType = "game_state_updated"
)

type WSMessage struct {
	Type      WSMessageType    `json:"type"` // ex: "action_declared"
	GameID    string           `json:"gameId"`
	Timestamp time.Time        `json:"timestamp"`
	GameState *PublicGameState `json:"gameState,omitempty"`
	Payload   any              `json:"payload,omitempty"`
}
