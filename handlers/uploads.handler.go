package handlers

import (
	"FileShare/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func HandleFileUpload(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")

	if err != nil {
		log.Error("Error uploading file: ", err.Error())
		return utils.GetMessageResponse(ctx, false, "", "Error uploading file", 500)
	}
	save_path := fmt.Sprintf("./uploads/%s", file.Filename)
	if err := ctx.SaveFile(file, save_path); err != nil {
		log.Error("Error uploading file: ", err.Error())
		return utils.GetMessageResponse(ctx, false, "", "Error uploading file", 500)
	}
	return utils.GetMessageResponse(ctx, true, "File Uploaded Successfully", "", 200)
}
