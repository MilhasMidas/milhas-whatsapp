package meow

import (
	"log"

	"github.com/Unicorn-s-Club/whats-unicorn/config"
	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
)

var (
	logger *config.Logger
	client *whatsmeow.Client
)

type MyClient struct {
	WAClient       *whatsmeow.Client
	eventHandlerID uint32
}

func (mycli *MyClient) Register() {
	mycli.eventHandlerID = mycli.WAClient.AddEventHandler(mycli.myEventHandler)
}

func (mycli *MyClient) myEventHandler(evt interface{}) {
	//logger.Debug("Eventos customizados recebidos:")
//	logger.Debugf("Evento recebido: %v", evt)
	//logger.Info("Id do Client:", mycli.WAClient.Store.ID)

}

func StartSession() {
	logger = config.GetLogger("meow")

	deviceStore, err := GetDevice()
	if err != nil {
		log.Fatal(err)
	}
	client = CreateClient(deviceStore)
	ShowQrCode(client)

	myClient := &MyClient{WAClient: client}
	myClient.Register()

	ReceivedMessageHandler(client)

	err = client.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}

	// fmt.Println("Cliente conectado. Pressione Ctrl+C para sair.")

	// stop := make(chan os.Signal, 1)
	// signal.Notify(stop, os.Interrupt)
	// <-stop

	// fmt.Println("Encerrando...")
	// client.Disconnect()

}
