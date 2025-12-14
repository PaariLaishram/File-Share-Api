package handlers

import (
	"FileShare/models"
	"encoding/binary"
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

// var clients = make(map[string]*websocket.Conn)
var clients = make(map[string]*models.Client)
var broadcast = make(chan models.UploadSignal)

func HandleWSConnections(w http.ResponseWriter, r *http.Request) {
	ws_conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws_conn.Close()

	for {
		messageType, message, err := ws_conn.ReadMessage()
		if err != nil {
			log.Error("Read Error:", err)
			return
		}
		switch messageType {
		case websocket.BinaryMessage:
			handleBinaryChunk(message)
		case websocket.TextMessage:
			var signal models.UploadSignal
			err := json.Unmarshal(message, &signal)
			if err != nil {
				log.Error("Error decoding JSON: ", err.Error())
				continue
			}

			share_link := models.SanitizeData(signal.ShareLink)
			user_type := models.SanitizeData(signal.UserType)
			conn_key := fmt.Sprintf("%s:%s", share_link, user_type)
			clients[conn_key] = &models.Client{
				Conn:       ws_conn,
				ShareLink:  signal.ShareLink,
				UserType:   signal.UserType,
				LastSignal: signal,
			}
			broadcast <- signal
		}
	}
}

func handleBinaryChunk(data []byte) {
	if len(data) < 4 {
		log.Error("Binary message too short - missing header")
		return
	}
	metadataLen := binary.BigEndian.Uint32(data[0:4])

	if len(data) < 4+int(metadataLen) {
		log.Error("Binary message malformed: metadata length exceeds message size")
		return
	}

	metadataJSON := data[4 : 4+metadataLen]
	var signal models.UploadSignal
	if err := json.Unmarshal(metadataJSON, &signal); err != nil {
		log.Error("Failed to parse metadata JSON:", err)
		return
	}
	chunkData := data[4+metadataLen:]
	// log.Info(models.SanitizeData(signal.UserType), models.SanitizeData(signal.ShareLink), models.SanitizeData(signal.ActionType))
	receiver_key := fmt.Sprintf("%s:%s", models.SanitizeData(signal.ShareLink), "receiver")
	reciver := clients[receiver_key]
	processFileChunk(&signal, chunkData, reciver)
}
func processFileChunk(signal *models.UploadSignal, chunkData []byte, receiver *models.Client) {
	metadataJSON, _ := json.Marshal(signal)
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(metadataJSON)))

	message := append(header, metadataJSON...) // header + metadata
	message = append(message, chunkData...)
	receiver.Conn.WriteMessage(websocket.BinaryMessage, message)
}

func HandleWSMessages() {
	for {
		incoming_signal := <-broadcast
		var err error
		user_type := models.SanitizeData(incoming_signal.UserType)
		share_link := models.SanitizeData(incoming_signal.ShareLink)
		action_type := models.SanitizeData(incoming_signal.ActionType)

		is_valid_share_link := isValidShareLink(action_type, user_type, share_link)

		outgoing_signal := incoming_signal.DeepCopy()
		outgoing_signal.IsValidShareLink = &is_valid_share_link

		var outgoing_user_type string

		if !is_valid_share_link {
			log.Info("Share Link is not valid")
			writeWSJSON("sender", share_link, outgoing_signal)
		} else {
			switch action_type {
			case "initConn":
				outgoing_user_type = outgoing_signal.InitConn()
			case "startUpload":
				outgoing_user_type = outgoing_signal.StartUpload()
			case "confirmUpload":
				outgoing_user_type = outgoing_signal.UploadConfirmation()
			case "ackChunk":
				outgoing_user_type = outgoing_signal.AckChunk()
			case "uploadComplete":
				outgoing_user_type = outgoing_signal.UploadComplete()
			}
			err = writeWSJSON(outgoing_user_type, models.SanitizeString(outgoing_signal.ShareLink), outgoing_signal)
			if err != nil {
				log.Error("Error writing WS Message")
			}
		}
	}
}

func isValidShareLink(actionType, userType, shareLink string) bool {
	conn_key := fmt.Sprintf("%s:%s", shareLink, userType)
	if userType == "sender" && (actionType == "initConn" || actionType == "startUpload") {
		conn_key = fmt.Sprintf("%s:%s", shareLink, "receiver")
	}
	if _, exists := clients[conn_key]; exists {
		return true
	}
	return false
}

func writeWSJSON(user_type, share_link string, data *models.UploadSignal) error {
	if user_type == "" {

		log.Error("Invalid user type for action type: ", models.SanitizeData(data.ActionType))
		return errors.New("writeWSMessage error: Invalid user type")
	}
	if share_link == "" {
		log.Error("Invalid share link", models.SanitizeData(data.ActionType))
		return errors.New("writeWSMessage error: Invalid share link")
	}
	conn_key := fmt.Sprintf("%s:%s", share_link, user_type)
	client := clients[conn_key]

	client.WriteMutex.Lock()
	defer client.WriteMutex.Unlock()
	err := client.Conn.WriteJSON(data)
	if err != nil {
		log.Error("Error writing json: ", err.Error())
		client.Conn.Close()
		return err
	}
	return nil
}
