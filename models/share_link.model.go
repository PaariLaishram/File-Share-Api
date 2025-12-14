package models

import (
	"FileShare/database"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

type ShareLink struct {
	ID   *int    `json:"id,omitempty"`
	Link *string `json:"link,omitempty"`
}

func (shareLink ShareLink) Get() (result []any, count int, error string) {
	fields := []any{&shareLink.ID, &shareLink.Link}
	result, err := database.RunQuery(get_share_link_query, fields, []any{}, &shareLink)
	if err != nil {
		log.Error("Error getting share links: ", err)
		return nil, 0, "Error getting File Share Link"
	}
	return result, len(result), ""
}

func (shareLink ShareLink) Add() (result []any, count int, error string) {
	link := GenerateLink()
	fields := []any{link}
	id, err := database.RunInsert(add_share_link_query, fields)
	if err != nil {
		log.Error("Error adding share link: ", err.Error())
		return nil, 0, "Error generating File Share Link"
	}
	last_inserted_token := ShareLink{
		ID:   &id,
		Link: &link,
	}
	var res []any = []any{last_inserted_token}
	return res, len(res), ""
}

var randsrc = rand.New(rand.NewSource(time.Now().Unix()))

func GenerateLink() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[randsrc.Intn(len(charset))]
	}
	return string(result)
}
