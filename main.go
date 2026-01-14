package main

import (
	db "FileShare/database"
	"FileShare/handlers"
	"FileShare/utils"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

type User struct {
	Name *string `json:"name,omitempty"`
	Age  *int    `json:"age,omitempty"`
}

type ShareToken struct {
	ID    *int    `json:"id,omitempty"`
	Token *string `json:"token,omitempty"`
}

func authenticate(ctx *fiber.Ctx) error {
	//validate access token
	authHeader := ctx.Get("Authorization")
	log.Info(authHeader)
	if authHeader == "" {
		return utils.GetMessageResponse(ctx, false, "", "Unathorised", 401)
	}
	return ctx.Next()
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Error("Env file not loaded: ", err.Error())
		return
	}
	db.ConnectToDatabase()
	var API_PORT = os.Getenv("API_PORT")
	var WS_PORT = os.Getenv("WS_PORT")

	go func() {
		log.Info("Starting ws server on:", WS_PORT)
		http.HandleFunc("/ws", handlers.HandleWSConnections)
		go handlers.HandleWSMessages()
		if err := http.ListenAndServe(WS_PORT, nil); err != nil {
			log.Fatal("Error listening to ws port: ", err)
		}
	}()

	app := fiber.New()
	app.Use(cors.New())

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Welcome to file share api")
	})
	api := app.Group("/api")
	v1 := api.Group("/v1")

	var auth fiber.Router = v1.Group("/auth")
	auth.Post("/login", handlers.Login)
	auth.Delete("/logout/:user", handlers.Logout)
	auth.Post("/refresh", handlers.ValidateRefreshToken)

	var profile fiber.Router = v1.Group("/profiles", authenticate)
	profile.Get("/:user", handlers.GetProfile)

	var share_links fiber.Router = v1.Group("/share-links")
	share_links.Get("/", handlers.GetShareLink)
	share_links.Post("/", handlers.GenerateShareLink)

	var uploads fiber.Router = v1.Group("/uploads")
	uploads.Post("/", handlers.HandleFileUpload)

	log.Info("Api listening on port: ", API_PORT)
	err := app.Listen(API_PORT)
	if err != nil {
		log.Fatal("Fiber failed to start: ", err)
	}
}
