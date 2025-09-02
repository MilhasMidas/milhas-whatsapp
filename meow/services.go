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
	"google.golang.org/genai"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func GetDevice() (*store.Device, error) {
	container := config.GetSQLiteMeow()
	if container == nil {
		return nil, fmt.Errorf("container is nil - database not initialized")
	}

	JID := types.NewJID("556581347018", "96@s.whatsapp.net")
	logger.Debugf("JID: %s", JID)

	// Try to get device by JID first
	deviceStore, err := container.GetDevice(JID)
	if err == nil && deviceStore != nil {
		return deviceStore, nil
	}

	// If not found by JID, try to get the first available device
	deviceStore, err = container.GetFirstDevice()
	if err != nil {
		return nil, fmt.Errorf("device not found: %v", err)
	}

	return deviceStore, nil
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
	// Global flag to track if offline sync has completed
	offlineSyncCompleted bool
	syncMu               sync.RWMutex
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
					if err := sendToN8N("https://3337-milhasmidas-milhasserve-k7raaurwn35.ws-us121.gitpod.io/content", m); err != nil {
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

// setOfflineSyncCompleted sets the offline sync completed flag
func setOfflineSyncCompleted() {
	syncMu.Lock()
	defer syncMu.Unlock()
	offlineSyncCompleted = true
	logger.Debugf("Offline sync completed flag set to true")
}

// isOfflineSyncCompleted checks if offline sync has completed
func isOfflineSyncCompleted() bool {
	syncMu.RLock()
	defer syncMu.RUnlock()
	return offlineSyncCompleted
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
	// Check if offline sync has completed
	if !isOfflineSyncCompleted() {
		logger.Errorf("cannot print groups: offline sync has not completed yet, groups are not loaded")
		return
	}

	groups, err := client.GetJoinedGroups()
	if err != nil {
		logger.Errorf("failed to get joined groups: %v", err)
		return
	}

	for _, g := range groups {
		fmt.Printf("Group Name: %s | Group ID: %s\n", g.Name, g.JID.String())
	}
}

// FindGroupByName finds a WhatsApp group by its name
// This function can only be executed after the OfflineSyncCompleted event occurs,
// otherwise the groups will not be loaded
func FindGroupByName(client *whatsmeow.Client, groupName string) (*types.GroupInfo, error) {
	// Check if offline sync has completed
	if !isOfflineSyncCompleted() {
		return nil, fmt.Errorf("cannot find group '%s': offline sync has not completed yet, groups are not loaded", groupName)
	}

	groups, err := client.GetJoinedGroups()
	logger.Debugf("groups: %v", groups)
	if err != nil {
		return nil, fmt.Errorf("failed to get joined groups: %v", err)
	}

	for _, group := range groups {
		if strings.EqualFold(group.Name, groupName) {
			return group, nil
		}
	}

	return nil, fmt.Errorf("group '%s' not found", groupName)
}

// Gemini client instance
var geminiClient *genai.Client

// Flight data structures matching the required schema
type Origin struct {
	City        string `json:"city"`
	AirportCode string `json:"airportCode"`
}

type Destination struct {
	City        string `json:"city"`
	AirportCode string `json:"airportCode"`
}

type Connection struct {
	City        string `json:"city"`
	AirportCode string `json:"airportCode"`
}

type Fees struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

type LoyaltyProgram struct {
	Name  string      `json:"name"`
	Miles float64     `json:"miles"`
	Fees  interface{} `json:"fees"` // Can be Fees struct or string
}

type Availability struct {
	SearchDate     string   `json:"searchDate"`
	AvailableDates []string `json:"availableDates"`
}

type FlightData struct {
	Origin          Origin           `json:"origin"`
	Destination     Destination      `json:"destination"`
	Airline         string           `json:"airline"`
	ServiceClass    string           `json:"serviceClass"`
	LoyaltyPrograms []LoyaltyProgram `json:"loyaltyPrograms"`
	Availability    Availability     `json:"availability"`
	Connections     []Connection     `json:"connections"`
}

// InitializeGeminiClient initializes the Gemini client with the API key
func InitializeGeminiClient(apiKey string) error {
	ctx := context.Background()
	config := &genai.ClientConfig{
		APIKey: apiKey,
	}
	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %v", err)
	}
	geminiClient = client
	return nil
}

// SendImageToGemini sends an image to Gemini API and returns structured flight data
func SendImageToGemini(imagePath string) (*FlightData, error) {
	if geminiClient == nil {
		return nil, fmt.Errorf("gemini client not initialized, call InitializeGeminiClient first")
	}

	ctx := context.Background()

	// Read image data
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %v", err)
	}

	// Create the schema for structured output
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"origin": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"city": {
						Type:        genai.TypeString,
						Description: "The departure city.",
					},
					"airportCode": {
						Type:        genai.TypeString,
						Description: "The IATA airport code for the origin.",
					},
				},
				Required: []string{"city", "airportCode"},
			},
			"destination": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"city": {
						Type:        genai.TypeString,
						Description: "The arrival city.",
					},
					"airportCode": {
						Type:        genai.TypeString,
						Description: "The IATA airport code for the destination.",
					},
				},
				Required: []string{"city", "airportCode"},
			},
			"connections": {
				Type:        genai.TypeArray,
				Description: "A list of connection points along the route.",
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"city": {
							Type:        genai.TypeString,
							Description: "The city of the connection.",
						},
						"airportCode": {
							Type:        genai.TypeString,
							Description: "The IATA airport code for the connection.",
						},
					},
					Required: []string{"city", "airportCode"},
				},
			},
			"airline": {
				Type:        genai.TypeString,
				Description: "The name of the airline.",
			},
			"serviceClass": {
				Type:        genai.TypeString,
				Description: "The class of service (e.g., 'Executive', 'Economy').",
			},
			"loyaltyPrograms": {
				Type:        genai.TypeArray,
				Description: "A list of loyalty programs and their associated costs.",
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"name": {
							Type:        genai.TypeString,
							Description: "The name of the loyalty program (e.g., 'Aeroplan').",
						},
						"miles": {
							Type:        genai.TypeNumber,
							Description: "The number of miles or points required for the flight.",
						},
						"fees": {
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"value": {
									Type:        genai.TypeNumber,
									Description: "The value of the fee.",
								},
								"currency": {
									Type:        genai.TypeString,
									Description: "The currency of the fee.",
								},
							},
						},
					},
					Required: []string{"name", "miles", "fees"},
				},
			},
			"availability": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"searchDate": {
						Type:        genai.TypeString,
						Description: "The date the search was performed.",
					},
					"availableDates": {
						Type:        genai.TypeArray,
						Description: "A list of available dates for the flight in DD/MM/YYYY format.",
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
				},
				Required: []string{"searchDate", "availableDates"},
			},
		},
		Required: []string{"origin", "destination", "airline", "serviceClass", "loyaltyPrograms", "availability"},
	}

	// Configure the generation settings
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	// Create content with text and image
	textPart := genai.NewPartFromText("Analyze this flight booking image and extract the flight information. Return the data in the exact JSON schema provided.")
	imagePart := genai.NewPartFromBytes(imageData, "image/jpeg")

	content := genai.NewContentFromParts([]*genai.Part{textPart, imagePart}, genai.RoleUser)

	// Generate content
	result, err := geminiClient.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		[]*genai.Content{content},
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}

	// Parse the JSON response
	var flightData FlightData
	if err := json.Unmarshal([]byte(result.Text()), &flightData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flight data: %v", err)
	}

	return &flightData, nil
}

// ProcessAlertGroupImages processes images from the "Alert" group and sends them to Gemini
func ProcessAlertGroupImages(client *whatsmeow.Client) error {
	var alertGroup *types.GroupInfo
	var alertGroupErr error

	// Set up event handler for the Alert group
	client.AddEventHandler(func(evt interface{}) {
		if _, ok := evt.(*events.OfflineSyncCompleted); ok {
			logger.Debugf("offline sync completed")
			setOfflineSyncCompleted()

			// Now that sync is completed, find the Alert group
			alertGroup, alertGroupErr = FindGroupByName(client, "Alert")
			if alertGroupErr != nil {
				logger.Errorf("failed to find Alert group after sync: %v", alertGroupErr)
				return
			}
			logger.Debugf("alertGroup found after sync: %v", alertGroup)
			fmt.Printf("[DEBUG] Found Alert group: %s (ID: %s)\n", alertGroup.Name, alertGroup.JID.String())
		}
		if msgEvt, ok := evt.(*events.Message); ok {
			// Only process messages if we have the alert group and sync is completed
			if alertGroup == nil || !isOfflineSyncCompleted() {
				return
			}

			// Check if message is from Alert group
			if !msgEvt.Info.IsGroup || msgEvt.Info.Chat.String() != alertGroup.JID.String() {
				return
			}

			// Check if message contains an image
			if msgEvt.Message.ImageMessage == nil {
				return
			}

			fmt.Printf("[DEBUG] Received image from Alert group\n")

			// Download the image
			imagePath, err := downloadImage(client, msgEvt)
			if err != nil {
				fmt.Printf("[ERROR] Failed to download image from Alert group: %v\n", err)
				return
			}

			// Process image with Gemini
			flightData, err := SendImageToGemini(imagePath)
			if err != nil {
				fmt.Printf("[ERROR] Failed to process image with Gemini: %v\n", err)
				// Clean up the image file
				os.Remove(imagePath)
				return
			}

			// Print the extracted flight data
			fmt.Printf("[SUCCESS] Extracted flight data from Alert group image:\n")
			flightDataJSON, _ := json.MarshalIndent(flightData, "", "  ")
			fmt.Printf("%s\n", string(flightDataJSON))

			// Clean up the image file
			if err := os.Remove(imagePath); err != nil {
				fmt.Printf("[WARN] Failed to remove image file %s: %v\n", imagePath, err)
			} else {
				fmt.Printf("[DEBUG] Removed image file %s\n", imagePath)
			}
		}
	})

	return nil
}

// StartAlertGroupProcessing starts monitoring the "Alert" WhatsApp group for images
// and processes them with Gemini API to extract flight booking information.
//
// Usage example:
//
//	err := StartAlertGroupProcessing("your-gemini-api-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// The function will:
// 1. Initialize the Gemini client with the API key using the official Google GenAI SDK
// 2. Find the WhatsApp group named "Alert"
// 3. Monitor for incoming images in that group
// 4. Download and process images with Gemini API using structured output
// 5. Extract structured flight data according to the specified schema
// 6. Print the extracted data to console
//
// This implementation uses the official Google GenAI Go SDK (google.golang.org/genai)
// which provides better error handling, type safety, and follows Google's best practices.
func StartAlertGroupProcessing(client *whatsmeow.Client, geminiAPIKey string) error {
	logger.Debugf("Starting Alert Group Processing")
	if client == nil {
		return fmt.Errorf("whatsapp client not initialized")
	}

	if geminiAPIKey == "" {
		return fmt.Errorf("gemini API key is required")
	}

	// Initialize Gemini client
	if err := InitializeGeminiClient(geminiAPIKey); err != nil {
		return fmt.Errorf("failed to initialize Gemini client: %v", err)
	}

	return ProcessAlertGroupImages(client)
}
