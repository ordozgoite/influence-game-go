package rooms

import "github.com/gobuffalo/buffalo"

func Register(app *buffalo.App, controller *RoomsController) {
	app.POST("/rooms", controller.CreateRoom)
	app.POST("/rooms/{joinCode}/join", controller.JoinRoom)
	app.POST("/rooms/{gameID}/start", controller.StartGame)
	// app.DELETE("/rooms/{joinCode}/leave", controller.DeleteRoom)
	// app.POST("/rooms/{joinCode}/leave", controller.LeaveRoom)

	// In-game routes
	app.POST("/rooms/{gameID}/actions/declare", controller.DeclareAction)
}
