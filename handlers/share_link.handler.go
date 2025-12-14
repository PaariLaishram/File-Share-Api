package handlers

import (
	"FileShare/models"
	"FileShare/utils"

	"github.com/gofiber/fiber/v2"
)

func GetShareLink(ctx *fiber.Ctx) error {
	var shareToken models.ShareLink
	result, count, err := shareToken.Get()
	return utils.GetArrayListReponse(ctx, result, count, "Share Token Successfully fetched", err)
}

func GenerateShareLink(ctx *fiber.Ctx) error {
	var shareToken models.ShareLink
	result, count, err := shareToken.Add()
	return utils.GetSingleReponse(ctx, result, count, "Share Token Successfull Added", err)
}
