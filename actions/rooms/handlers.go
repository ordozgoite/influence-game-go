package rooms

import (
	"fmt"
	"influence_game/internal/game"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/rs/zerolog/log"
)

var renderer = render.New(render.Options{})

type RoomsController struct {
	Store *game.Store
}

func NewRoomsController(store *game.Store) *RoomsController {
	return &RoomsController{Store: store}
}

func (controller *RoomsController) CreateRoom(ctx buffalo.Context) error {
	log.Info().Msg("Creating new game...")
	var dto CreateRoomDTO

	if err := ctx.Bind(&dto); err != nil {
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": "invalid_json",
		}))
	}

	if err := dto.Validate(); err != nil {
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}

	nickname := dto.Nickname

	newGamePublicInfo, err := controller.Store.CreateGameRoom(nickname)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new game.")
		return ctx.Render(500, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}
	return ctx.Render(200, renderer.JSON(newGamePublicInfo))
}

func (controller *RoomsController) JoinRoom(ctx buffalo.Context) error {
	log.Info().Msg("Joining room...")
	var dto JoinRoomDTO

	if err := ctx.Bind(&dto); err != nil {
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": "invalid_json",
		}))
	}

	if err := dto.Validate(); err != nil {
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}

	roomID := ctx.Param("gameID")
	fmt.Println("Joining room with gameID=", roomID)

	game, player, sessionToken, err := controller.Store.Join(roomID, dto.Nickname)
	if err != nil {
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}

	return ctx.Render(200, renderer.JSON(map[string]any{
		"gameID":       game.ID,
		"playerID":     player.ID,
		"token":        sessionToken,
		"playersCount": len(game.Players),
	}))
}
