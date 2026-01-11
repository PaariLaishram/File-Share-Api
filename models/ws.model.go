package models

import (
	"FileShare/utils"
	"sync"

	"github.com/gofiber/fiber/v2/log"
	"github.com/gorilla/websocket"
)

type UploadSignal struct {
	UserType         *string                `json:"userType,omitempty"`
	ShareLink        *string                `json:"shareLink,omitempty"`
	ActionType       *string                `json:"actionType,omitempty"`
	IsValidShareLink *bool                  `json:"isValidShareLink,omitempty"`
	ConfirmUpload    *bool                  `json:"confirmUpload,omitempty"`
	ChunkIndex       *int                   `json:"chunkIndex,omitempty"`
	TotalChunks      *int                   `json:"totalChunks,omitempty"`
	FileName         *string                `json:"fileName,omitempty"`
	Offer            *RTCSessionDescription `json:"offer,omitempty"`
	Answer           *RTCSessionDescription `json:"answer,omitempty"`
	Candidate        *RTCIceCandidate       `json:"candidate,omitempty"`
}

type RTCIceCandidate struct {
	Address          *string `json:"address,omitepmpty"`
	Candidate        *string `json:"candidate,omitempty"`
	Component        *string `json:"component,omitempty"`
	Foundation       *string `json:"foundation,omitempty"`
	Port             *int    `json:"port,omitempty"`
	Priority         *int    `json:"priority,omitempty"`
	Protocol         *string `json:"protocol,omitempty"`
	RelatedAddress   *string `json:"relatedAddress,omitempty"`
	RelatedPort      *int    `json:"relatedPort,omitempty"`
	SdpMid           *string `json:"sdpMid,omitempty"`
	SdpMLineIndex    *int    `json:"int,omitempty"`
	TcpType          *string `json:"tcpType,omitempty"`
	Type             *string `json:"type,omitempty"`
	UsernameFragment *string `json:"usernameFragment,omitempty"`
}

type RTCSessionDescription struct {
	Type *string `json:"type,omitempty"`
	SDP  *string `json:"sdp,omitempty"`
}

type Client struct {
	Conn       *websocket.Conn
	ShareLink  *string
	UserType   *string
	LastSignal UploadSignal
	WriteMutex sync.Mutex
	ConnKey    *string
}

func (data *UploadSignal) DeepCopy() *UploadSignal {
	if data == nil {
		return nil
	}
	return &UploadSignal{
		UserType:         utils.CopyDataPtr(data.UserType),
		ShareLink:        utils.CopyDataPtr(data.ShareLink),
		ActionType:       utils.CopyDataPtr(data.ActionType),
		IsValidShareLink: utils.CopyDataPtr(data.IsValidShareLink),
		ConfirmUpload:    utils.CopyDataPtr(data.ConfirmUpload),
		FileName:         utils.CopyDataPtr(data.FileName),
		TotalChunks:      utils.CopyDataPtr(data.TotalChunks),
	}
}

func (data *UploadSignal) StartUpload() (user_type string) {
	data_user_type := SanitizeData(data.UserType)
	switch data_user_type {
	case "sender":
		return "receiver"
	case "receiver":
		return "sender"
	default:
		return ""
	}

}

func (data *UploadSignal) UploadConfirmation() (user_type string) {
	data_user_type := SanitizeData(data.UserType)

	switch data_user_type {
	case "receiver":
		// continue_upload := SanitizeBoolean(data.ConfirmUpload)
		return "sender"
	default:
		return ""
	}
}

func (data *UploadSignal) InitConn() (user_type string) {
	data_user_type := SanitizeData(data.UserType)

	switch data_user_type {
	case "sender":
		return SanitizeData(data.UserType)
	default:
		return ""
	}

}

func (data *UploadSignal) SendChunk() (user_type string) {
	data_user_type := SanitizeData(data.UserType)

	switch data_user_type {
	case "sender":
		return "receiver"
	case "receiver":
		return "sender"
	default:
		return ""
	}
}

func (data *UploadSignal) AckChunk() (user_type string) {
	data_user_type := SanitizeData(data.UserType)

	switch data_user_type {
	case "receiver":
		return "sender"
	default:
		return ""
	}

}

func (data *UploadSignal) UploadComplete() (user_type string) {
	data_user_type := SanitizeData(data.UserType)
	switch data_user_type {
	case "sender":
		log.Info(SanitizeData(data.FileName))
		return "receiver"
	default:
		return ""
	}
}
