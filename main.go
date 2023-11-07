package main

import (
	"projectfiber/controllers"
	"projectfiber/repository"
	"projectfiber/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	repository.ConnectDatabase()

	app := fiber.New()
	api := app.Group("/api")

	// Akses todolist
	first := api.Group("/firsts")
	todolist := first.Group("/todolists")
	todolist.Get("/", controllers.GetTodoLists)
	todolist.Get("/:id", controllers.GetTodoList)
	todolist.Post("/", controllers.CreateTodoList)
	todolist.Patch("/:id", controllers.UpdateTodoList)
	todolist.Delete("/:id", controllers.DeleteTodoList)

	// Pass the same app instance to AuthRoute
	routes.AuthRoute(app)

	app.Listen(":8000")
}

// todolist.Get("/", middleware.AuthUserLogin(), controllers.GetTodoLists)
// todolist.Get("/:id", middleware.AuthUserLogin(), controllers.GetTodoList)
// todolist.Post("/", middleware.AuthUserLogin(), controllers.CreateTodoList)
// todolist.Patch("/:id", middleware.AuthUserLogin(), controllers.UpdateTodoList)
// todolist.Delete("/:id", middleware.AuthUserLogin(), controllers.DeleteTodoList)

//akses login
// first := api.Group("/firsts")
// first.Post("/signup", controllers.SignUpUser)
// first.Post("/login", controllers.LoginUser)
// first.Get("/login/:username", controllers.GetLoginUsername)
// first.Post("/logout", middleware.AuthUserLogin(), controllers.LogOutUser)
// first.Post("/refresh", middleware.RenewAccToken)
