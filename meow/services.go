package meow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Unicorn-s-Club/whats-unicorn/config"
	"github.com/mdp/qrterminal/v3"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
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


type GroupBuffer struct {
	messages    []string  
	lastMessage string
	timer       *time.Timer
	mu          sync.Mutex
}

var (
	groupBuffers = make(map[string]*GroupBuffer)
	buffersMu    sync.Mutex
)





func ReceivedMessageHandler(client *whatsmeow.Client) {
	allowedGroups := []string{
		"120363047082894049@g.us",
		"120363393943954922@g.us", 
		"120363401595010116@g.us", 
	}

	client.AddEventHandler(func(evt interface{}) {
		if msgEvt, ok := evt.(*events.Message); ok {
			if !msgEvt.Info.IsGroup {
				return
			}

			groupID := msgEvt.Info.Chat.String()
			if !contains(allowedGroups, groupID) {
				return
			}

			var messageContent string

			if msgEvt.Message.GetConversation() != "" {
				messageContent = msgEvt.Message.GetConversation()
			}

			if msgEvt.Message.ImageMessage != nil {
				imagePath, err := downloadImage(client, msgEvt)
				if err != nil {
					fmt.Printf("[ERROR] Failed to download image: %v\n", err)
					return
				}
				messageContent = "[Image:" + imagePath + "]"
			}

			if messageContent == "" {
				return
			}

			fmt.Printf("[DEBUG] Received from %s: %s\n", groupID, messageContent)

			buffersMu.Lock()
			buf, exists := groupBuffers[groupID]
			if !exists {
				buf = &GroupBuffer{}
				groupBuffers[groupID] = buf
				fmt.Printf("[DEBUG] Created buffer for %s\n", groupID)
			}
			buffersMu.Unlock()

			buf.mu.Lock()

			buf.messages = append(buf.messages, messageContent)
			fmt.Printf("[DEBUG] Buffer for %s now has %d messages\n", groupID, len(buf.messages))

			if buf.lastMessage != "" && messageContent == buf.lastMessage {
				fmt.Printf("[DEBUG] Duplicate message — ending buffer for %s early\n", groupID)
				if buf.timer != nil {
					buf.timer.Stop()
				}

				msgs := append([]string(nil), buf.messages...)

				buffersMu.Lock()
				delete(groupBuffers, groupID)
				buffersMu.Unlock()

				buf.messages = nil
				buf.lastMessage = ""
				buf.timer = nil
				buf.mu.Unlock()

				go func(g string, m []string) {
					if err := sendToN8N("https://3337-milhasmidas-milhasserve-k7raaurwn35.ws-us121.gitpod.io/content",  m); err != nil {
						fmt.Printf("[ERROR] sendToN8N (duplicate) for %s: %v\n", g, err)
					} else {
						fmt.Printf("[DEBUG] sendToN8N (duplicate) succeeded for %s\n", g)
					}
				}(groupID, msgs)

				return
			}

			buf.lastMessage = messageContent

			if buf.timer != nil {
				buf.timer.Stop()
				fmt.Printf("[DEBUG] Reset timer for %s\n", groupID)
			} else {
				fmt.Printf("[DEBUG] Start timer for %s\n", groupID)
			}

			buf.timer = time.AfterFunc(10*time.Second, func() {
				buf.mu.Lock()
				msgs := append([]string(nil), buf.messages...)
				buf.messages = nil
				buf.lastMessage = ""
				if buf.timer != nil {
					buf.timer.Stop()
				}
				buf.timer = nil
				buf.mu.Unlock()

				buffersMu.Lock()
				delete(groupBuffers, groupID)
				buffersMu.Unlock()

				if err := sendToN8N("https://3337-milhasmidas-milhasserve-k7raaurwn35.ws-us121.gitpod.io/content", msgs); err != nil {
					fmt.Printf("[ERROR] sendToN8N (timer) for %s: %v\n", groupID, err)
				} else {
					fmt.Printf("[DEBUG] sendToN8N (timer) succeeded for %s\n", groupID)
				}
			})

			buf.mu.Unlock()
		}
	})
}

func sendToN8N(url string, messages []string) error {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	var textMsgs []string
	var attachedPaths []string

	for _, m := range messages {
		if strings.HasPrefix(m, "[Image:") && strings.HasSuffix(m, "]") {
			path := m[len("[Image:") : len(m)-1]
			f, err := os.Open(path)
			if err != nil {
				fmt.Printf("[WARN] could not open image %s: %v\n", path, err)
				continue
			}
			// renamed "files" -> "image"
			part, err := writer.CreateFormFile("image", filepath.Base(path))
			if err != nil {
				f.Close()
				fmt.Printf("[WARN] create form file failed for %s: %v\n", path, err)
				continue
			}
			_, err = io.Copy(part, f)
			f.Close()
			if err != nil {
				fmt.Printf("[WARN] copy file %s failed: %v\n", path, err)
			}
			attachedPaths = append(attachedPaths, path)
		} else {
			textMsgs = append(textMsgs, m)
		}
	}

	msgsJSON, _ := json.Marshal(textMsgs)
	if err := writer.WriteField("messages", string(msgsJSON)); err != nil {
		writer.Close()
		return fmt.Errorf("write messages field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("n8n returned status %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("[DEBUG] n8n response: %s\n", string(respBody))
	for _, p := range attachedPaths {
		if err := os.Remove(p); err != nil {
			fmt.Printf("[WARN] failed to remove %s: %v\n", p, err)
		} else {
			fmt.Printf("[DEBUG] removed %s\n", p)
		}
	}
	return nil
}



func downloadImage(client *whatsmeow.Client, msgEvt *events.Message) (string, error) {
	data, err := client.Download(msgEvt.Message.ImageMessage)
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("image_%d.jpg", time.Now().UnixNano())
	filePath := "./" + fileName

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", err
	}

	fmt.Printf("[DEBUG] Image saved to %s\n", filePath)
	return filePath, nil
}



type Message struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
	Type    string `json:"type,omitempty"` 
}

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
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
