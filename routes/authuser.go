package routes

import (
	"projectfiber/controllers"
	"projectfiber/middleware"
	"projectfiber/repository"

	"github.com/gofiber/fiber/v2"
)

func AuthRoute(app *fiber.App) {
	repository.ConnectDatabase()
	//app := fiber.New()
	api := app.Group("/api")

	first := api.Group("/firsts")
	// first.Post("/signup", controllers.SignUpUser)
	// first.Post("/login", controllers.LoginUser)
	// first.Get("/:username", middleware.AuthUserLogin(), controllers.GetLoginUsername)
	// first.Post("/logout", middleware.AuthUserLogin(), controllers.LogOutUser)
	// first.Post("/refresh", middleware.RenewAccToken)

	// first.Post("/jobpredict", middleware.AuthUserLogin(),
	// 	controllers.NewJobUserController(repository.DB).UploadDocument)

	first.Post("/signup", controllers.SignUpUser)
	first.Post("/login", controllers.LoginUser)
	first.Get("/:username", controllers.GetLoginUsername)
	first.Post("/logout", controllers.LogOutUser)
	first.Post("/refresh", middleware.RenewAccToken)
	first.Post("/jobpredict", controllers.NewJobUserController(repository.DB).UploadDocument)
	//first.Post("/jobpredict", middleware.AuthUserLogin(), controllers.NewJobUserController(repository.DB).UploadDocument)
}

///api/firsts/login
