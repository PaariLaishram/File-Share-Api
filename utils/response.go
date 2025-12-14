package utils

import "github.com/gofiber/fiber/v2"

type Response struct {
	Success *bool   `json:"success,omitempty"`
	Message *string `json:"message,omitempty"`
	Count   *int    `json:"count,omitempty"`
	Result  *any    `json:"result,omitempty"`
}

func FormatMessageResponse(flag bool, message string) Response {
	response := Response{
		Success: &flag,
		Message: &message,
	}
	return response
}

func GetMessageResponse(ctx *fiber.Ctx, flag bool, success, failure string, status int, cookies ...*fiber.Cookie) error {
	for _, cookie := range cookies {
		ctx.Cookie(cookie)
	}
	response := FormatMessageResponse(flag, success)
	if !flag {
		response.Message = &failure
		return ctx.Status(status).JSON(response)
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func GetArrayListReponse(ctx *fiber.Ctx, result any, count int, success, err string) error {
	flag := true
	response := FormatMessageResponse(flag, success)
	if err != "" {
		flag = false
		response.Success = &flag
		response.Message = &err
		return ctx.Status(fiber.StatusInternalServerError).JSON(response)
	}
	response.Count = &count
	response.Result = &result
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func GetSingleReponse(ctx *fiber.Ctx, result any, count int, success, err string, cookies ...*fiber.Cookie) error {
	for _, cookie := range cookies {
		ctx.Cookie(cookie)
	}
	flag := true
	response := FormatMessageResponse(flag, success)
	if err != "" {
		flag = false
		response.Success = &flag
		response.Message = &err
		return ctx.Status(fiber.StatusInternalServerError).JSON(response)
	}
	var response_result any
	if slice, ok := result.([]any); ok {
		if len(slice) == 0 {
			response_result = struct {
			}{}
		} else {
			response_result = slice[0]
		}
	}
	response.Count = &count
	response.Result = &response_result
	return ctx.Status(fiber.StatusOK).JSON(response)

}
