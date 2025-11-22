package rooms

import "github.com/gobuffalo/buffalo"

func Register(app *buffalo.App, controller *RoomsController) {
	app.POST("/rooms", controller.CreateRoom)
}
