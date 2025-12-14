package handlers

import (
	"FileShare/models"
	"FileShare/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func GetProfile(ctx *fiber.Ctx) error {
	var profile models.Profile
	user := ctx.Params("user")
	log.Info(user)
	result, err := profile.GetProfile(user)
	return utils.GetSingleReponse(ctx, result, len(result), "Profile fetched", err.Error())
}
