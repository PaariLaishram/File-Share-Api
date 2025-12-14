package handlers

import (
	"FileShare/models"
	"FileShare/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func Login(ctx *fiber.Ctx) error {
	var login_body = new(models.LoginBody)
	err := ctx.BodyParser(&login_body)
	success := "Login successful"
	unathorized := 401
	if err != nil {
		log.Error("Error parsing body:", err)
		return utils.GetMessageResponse(ctx, false, "", "Invalid Credentials", unathorized)
	}
	flag, failure, result := login_body.Login()
	if flag {
		cookie := &fiber.Cookie{
			Name:     "refreshToken",
			Value:    models.SanitizeData(result.RefreshToken),
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			Path:     "/",
			MaxAge:   60 * 60 * 24 * 7,
		}
		return utils.GetSingleReponse(ctx, []interface{}{result}, len([]interface{}{result}), success, failure, cookie)
	}
	return utils.GetMessageResponse(ctx, flag, "", failure, unathorized)
}

func Logout(ctx *fiber.Ctx) error {
	user := ctx.Params("user")
	flag, failure := models.Logout(user)
	cookie := &fiber.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Unix(0, 0),
	}
	return utils.GetMessageResponse(ctx, flag, "Logout successful", failure, fiber.StatusOK, cookie)
}

func ValidateRefreshToken(ctx *fiber.Ctx) error {
	refresh_token := ctx.Cookies("refreshToken")
	if len(refresh_token) == 0 {
		log.Warn("Refresh token not found")
		return utils.GetMessageResponse(ctx, false, "", "Unathorized", 401)
	}
	flag, result := models.ValidateRefreshToken(refresh_token)
	if flag {
		cookie := &fiber.Cookie{
			Name:     "refreshToken",
			Value:    models.SanitizeData(result.RefreshToken),
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			Path:     "/",
			MaxAge:   60 * 60 * 24 * 7,
		}
		return utils.GetSingleReponse(ctx, []any{result}, len([]any{result}), "Validate succesfull", "", cookie)
	}
	return utils.GetMessageResponse(ctx, flag, "", "Unathorized", 401)
}
