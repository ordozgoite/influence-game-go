package actions

import (
	"influence_game/actions/rooms"
	"influence_game/internal/game"
	"influence_game/locales"
	"sync"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/middleware/contenttype"
	"github.com/gobuffalo/middleware/forcessl"
	"github.com/gobuffalo/middleware/i18n"
	"github.com/gobuffalo/middleware/paramlogger"
	"github.com/gobuffalo/x/sessions"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

var ENV = envy.Get("GO_ENV", "development")

var (
	app     *buffalo.App
	appOnce sync.Once
	T       *i18n.Translator
)

func App() *buffalo.App {
	appOnce.Do(func() {
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionStore: sessions.Null{},
			PreWares: []buffalo.PreWare{
				cors.Default().Handler,
			},
			SessionName: "_influence_game_session",
		})

		// Middleware
		app.Use(forceSSL())
		app.Use(paramlogger.ParameterLogger)
		app.Use(contenttype.Set("application/json"))

		// Rotas básicas
		app.GET("/", HomeHandler)
		app.GET("/healthz", func(ctx buffalo.Context) error {
			return ctx.Render(200, r.JSON(map[string]string{
				"status": "ok",
			}))
		})

		// ============================================================
		// 🔥 Redis Client
		// ============================================================
		redisAddr := envy.Get("REDIS_ADDR", "localhost:6379")
		redisPassword := envy.Get("REDIS_PASSWORD", "")
		redisClient := redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       0,
		})

		// ============================================================
		// 🔥 Store + RoomsController
		// ============================================================
		store := game.NewStore(redisClient)
		roomsController := rooms.NewRoomsController(store)

		// Registrar rotas da feature /rooms
		rooms.Register(app, roomsController)
		// ============================================================
	})

	return app
}

func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(locales.FS(), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     false,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}
