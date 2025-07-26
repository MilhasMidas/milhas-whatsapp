package schemas

import "go.mau.fi/whatsmeow/types"

type SendMessageResponse struct {
	Message string    `json:"message"`
	Sender  types.JID `json:"sender"`
}
