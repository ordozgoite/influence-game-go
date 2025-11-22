package rooms

import (
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
	log.Info().Msg("Creating new game.")

	newGame, err := controller.Store.NewGame()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new game.")
		return ctx.Render(500, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}
	return ctx.Render(200, renderer.JSON(map[string]any{
		"gameID": newGame.ID,
	}))
}
