package rooms

import (
	"influence_game/internal/game"
	"strings"

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

func (controller *RoomsController) StartGame(ctx buffalo.Context) error {
	log.Info().Msg("Starting game.")
	gameID := ctx.Param("gameID")

	authHeader := ctx.Request().Header.Get("Authorization")
	const prefix = "Bearer "

	if !strings.HasPrefix(authHeader, prefix) {
		log.Error().Msg("Missing or invalid Authorization header.")
		return ctx.Render(401, renderer.JSON(map[string]any{
			"error": "missing or invalid Authorization header",
		}))
	}

	sessionToken := strings.TrimPrefix(authHeader, prefix)
	if sessionToken == "" {
		log.Error().Msg("Empty bearer token.")
		return ctx.Render(401, renderer.JSON(map[string]any{
			"error": "empty bearer token",
		}))
	}

	updatedGameState, err := controller.Store.StartGame(gameID, sessionToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start game.")
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}

	log.Info().Msg("Game started successfully.")

	return ctx.Render(200, renderer.JSON(updatedGameState))
}

func (controller *RoomsController) DeclareAction(ctx buffalo.Context) error {
	log.Info().Msg("Declaring action.")
	gameID := ctx.Param("gameID")

	var dto DeclareActionDTO
	if err := ctx.Bind(&dto); err != nil {
		log.Error().Err(err).Msg("Failed to bind declare action request.")
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": "invalid_json",
		}))
	}

	if err := dto.Validate(); err != nil {
		log.Error().Err(err).Msg("Failed to validate declare action request.")
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}

	authHeader := ctx.Request().Header.Get("Authorization")
	const prefix = "Bearer "

	if !strings.HasPrefix(authHeader, prefix) {
		log.Error().Msg("Missing or invalid Authorization header.")
		return ctx.Render(401, renderer.JSON(map[string]any{
			"error": "missing or invalid Authorization header",
		}))
	}

	sessionToken := strings.TrimPrefix(authHeader, prefix)
	if sessionToken == "" {
		log.Error().Msg("Empty bearer token.")
		return ctx.Render(401, renderer.JSON(map[string]any{
			"error": "empty bearer token",
		}))
	}

	currentGameState, err := controller.Store.DeclareAction(
		gameID,
		game.DeclareActionPayload{
			ActionName:     dto.ActionName,
			TargetPlayerID: dto.TargetPlayerID,
		},
		sessionToken,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to declare action.")
		return ctx.Render(400, renderer.JSON(map[string]any{
			"error": err.Error(),
		}))
	}

	log.Info().Msg("Action declared successfully.")

	return ctx.Render(200, renderer.JSON(currentGameState))
}
