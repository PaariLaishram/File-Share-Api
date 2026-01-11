package handlers

import (
	"FileShare/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2/log"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[string]*models.Client)
var messages = make(chan models.UploadSignal)

func HandleWSConnections(w http.ResponseWriter, r *http.Request) {
	ws_conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	var client *models.Client

	defer func() {
		if client != nil {
			delete(clients, models.SanitizeString(client.ConnKey))
			ws_conn.Close()
			log.Info("Client disconnected: ", models.SanitizeData(client.ConnKey))
		}
	}()
	for {
		_, message, err := ws_conn.ReadMessage()
		if err != nil {
			log.Error("Read Error:", err)
			return
		}
		var signal models.UploadSignal
		err = json.Unmarshal(message, &signal)
		if err != nil {
			log.Error("error unmarshalling: ", err.Error())
			return
		}
		if client == nil {
			share_link := models.SanitizeData(signal.ShareLink)
			user_type := models.SanitizeData(signal.UserType)
			conn_key := fmt.Sprintf("%s:%s", share_link, user_type)
			client = &models.Client{
				Conn:       ws_conn,
				ShareLink:  signal.ShareLink,
				UserType:   signal.UserType,
				LastSignal: signal,
				ConnKey:    &conn_key,
			}
			clients[conn_key] = client
		}
		messages <- signal
	}
}

func HandleWSMessages() {
	for {
		signal := <-messages
		user_type := models.SanitizeData(signal.UserType)
		action_type := models.SanitizeData(signal.ActionType)
		share_link := models.SanitizeData(signal.ShareLink)

		is_valid_share_link := isValidShareLink(action_type, user_type, share_link)
		signal.IsValidShareLink = &is_valid_share_link
		// outgoing_signal := incoming_signal.DeepCopy()
		// outgoing_signal.IsValidShareLink = &is_valid_share_link

		if !is_valid_share_link {
			log.Error("Invalid share link")
			writeWSJSON("sender", share_link, &signal)
		}
		switch action_type {
		case "initConn":
			if user_type == "sender" {
				writeWSJSON("sender", share_link, &signal)
			}
		case "createOffer":
			if user_type == "sender" {
				writeWSJSON("receiver", share_link, &signal)
			}
		case "answerOffer":
			if user_type == "receiver" {
				writeWSJSON("sender", share_link, &signal)
			}
		case "iceCandidate":
			msg_for := "sender"
			if user_type == "sender" {
				msg_for = "receiver"
			}
			writeWSJSON(msg_for, share_link, &signal)
		}
	}
}

func isValidShareLink(actionType, userType, shareLink string) bool {
	conn_key := fmt.Sprintf("%s:%s", shareLink, userType)
	if userType == "sender" && actionType == "initConn" {
		conn_key = fmt.Sprintf("%s:%s", shareLink, "receiver")
	}
	if _, exist := clients[conn_key]; exist {
		return true
	}
	return false
}

func writeWSJSON(user_type, share_link string, signal *models.UploadSignal) error {
	if user_type == "" {
		log.Error("invalid user type for action type: ", models.SanitizeData(signal.ActionType))
		return errors.New("error: invalid user type")
	}
	if share_link == "" {
		log.Error("invalid share link ", models.SanitizeData(signal.ActionType))
		return errors.New("error: invalid share link")
	}
	conn_key := fmt.Sprintf("%s:%s", share_link, user_type)
	client := clients[conn_key]
	client.WriteMutex.Lock()
	defer client.WriteMutex.Unlock()

	err := client.Conn.WriteJSON(signal)
	if err != nil {
		log.Error("error in writing json: ", err.Error())
		client.Conn.Close()
		return err
	}
	return nil
}
