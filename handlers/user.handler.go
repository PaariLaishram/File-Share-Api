package handlers

import (
	"FileShare/models"
	"FileShare/utils"

	"github.com/gofiber/fiber/v2"
)

func GetProfile(ctx *fiber.Ctx) error {
	var profile models.Profile
	user := ctx.Params("user")
	result, err := profile.GetProfile(user)
	return utils.GetSingleReponse(ctx, result, len(result), "Profile fetched", err.Error())
}
