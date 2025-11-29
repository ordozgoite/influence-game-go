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
	log.Info().Msg("Creating new game room.")
	var dto CreateRoomDTO

	if err := ctx.Bind(&dto); err != nil {
		log.Error().Err(err).Msg("Failed to bind create room request.")
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": "invalid_json",
		}))
	}

	if err := dto.Validate(); err != nil {
		log.Error().Err(err).Msg("Failed to validate create room request.")
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}

	nickname := dto.Nickname

	newGamePublicInfo, err := controller.Store.CreateGameRoom(nickname)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new game room.")
		return ctx.Render(500, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}

	log.Info().Msg("Created new game room successfully.")

	return ctx.Render(200, renderer.JSON(newGamePublicInfo))
}

func (controller *RoomsController) JoinRoom(ctx buffalo.Context) error {
	log.Info().Msg("Joining game room.")
	var dto JoinRoomDTO

	if err := ctx.Bind(&dto); err != nil {
		log.Error().Err(err).Msg("Failed to bind join room request.")
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": "invalid_json",
		}))
	}

	if err := dto.Validate(); err != nil {
		log.Error().Err(err).Msg("Failed to validate join room request.")
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}

	joinCode := ctx.Param("joinCode")

	onboardingResult, err := controller.Store.Join(
		joinCode,
		dto.Nickname,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to join game room.")
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}

	log.Info().Msg("Joined game room successfully.")

	return ctx.Render(200, renderer.JSON(onboardingResult))
}
