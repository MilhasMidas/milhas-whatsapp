package meow

import (
	"context"
	"fmt"
	"os"

	waProto "go.mau.fi/whatsmeow/binary/proto"

	"github.com/Unicorn-s-Club/whats-unicorn/config"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

func GetDevice() (*store.Device, error) {
	container := config.GetSQLiteMeow()
	JID := types.NewJID("556581347018", "96@s.whatsapp.net")
	logger.Debugf("JID: %s", JID)
	container.GetDevice(JID)
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return nil, fmt.Errorf("device not found: %v", err)
	}

	return deviceStore, err
}

func CreateDevice() *store.Device {
	container := config.GetSQLiteMeow()

	deviceStore := container.NewDevice()

	return deviceStore
}

func CreateClient(deviceStore *store.Device) *whatsmeow.Client {
	client := whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "INFO", true))
	return client
}

func GetClient() *whatsmeow.Client {
	return client
}

func ShowQrCode(client *whatsmeow.Client) {
	qrChan, _ := client.GetQRChannel(context.Background())
	go func() {
		for evt := range qrChan {
			if evt.Code != "" {
				fmt.Println("Escaneie o QR Code abaixo para conectar:")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			}
		}
	}()
}
func SendToGroup(client *whatsmeow.Client, groupJID string, text string) error {
	// Build the JID
	jid := types.NewJID(groupJID, "g.us") // group JID always ends with g.us

	// Build the message
	msg := &waProto.Message{
		Conversation: proto.String(text),
	}

	// Send
	_, err := client.SendMessage(context.Background(), jid, msg)
	return err
}
func ReceivedMessageHandler(client *whatsmeow.Client) {
	allowedGroups := []string{
		"120363427657699137@g.us",
		"120363400000000000@g.us",
	}

	client.AddEventHandler(func(evt interface{}) {
		if msgEvt, ok := evt.(*events.Message); ok {
			if msgEvt.Info.IsGroup {
				groupID := msgEvt.Info.Chat.String()

				logger.Debugf("Group Message from %s: %s",
					groupID,
					msgEvt.Message.GetConversation(),
				)

				if msgEvt.Info.Sender.User == client.Store.ID.User {
					logger.Debugf("Ignoring message from self: %s", msgEvt.Info.Sender.User)
					return
				}

				if contains(allowedGroups, groupID) {
					//err := SendToGroup(client, groupID, "Hello group!")
					//Send the message to the API
					/*if err != nil {
						logger.Errorf("failed to send message: %v", err)
					} else {
						logger.Debugf("Message sent to group %s!", groupID)
					}*/
				}
			}
		}
	})
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func SendMessageHandler(client *whatsmeow.Client, sender types.JID, message string) {
	/*
	   	reply := &waProto.Message{
	   		Conversation: proto.String(message),
	   	}

	   _, err := client.SendMessage(context.Background(), sender, reply)

	   	if err != nil {
	   		fmt.Println("Erro ao enviar resposta:", err)
	   	} else {

	   		fmt.Println("[Auto-resposta]: pong enviada com sucesso.")
	   	}
	*/
}

func GetAllDevices() ([]*store.Device, error) {
	container := config.GetSQLiteMeow()
	devices, err := container.GetAllDevices()
	logger.Debugf("devices: %v", devices)
	if err != nil {
		logger.Errorf("failed to get devices: %v", err)
		return nil, fmt.Errorf("failed to get devices: %v", err)
	}
	return devices, nil
}

func PrintGroups(client *whatsmeow.Client) {
	groups, err := client.GetJoinedGroups()
	if err != nil {
		logger.Errorf("failed to get joined groups: %v", err)
		return
	}

	for _, g := range groups {
		fmt.Printf("Group Name: %s | Group ID: %s\n", g.Name, g.JID.String())
	}
}
