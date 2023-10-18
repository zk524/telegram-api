package main

import (
	"fmt"

	"github.com/zelenin/go-tdlib/client"
)

// Telegram phonenumber, china: +86...
var phoneNumber = ""

// Application identifier for Telegram API access, which can be obtained at https://my.telegram.org
var appid = int32(0)
var appHash = ""

// Target chat
var chatID = int64(0)

// Sent Text Message
var msg = "test"

func main() {
	var cli *client.Client
	var err error
	authorizer := client.ClientAuthorizer()
	go func() {
		for {
			if currentState, ok := <-authorizer.State; !ok {
				return
			} else {
				switch currentState.AuthorizationStateType() {
				case client.TypeAuthorizationStateWaitPhoneNumber:
					if phoneNumber == "" {
						fmt.Print("Enter phoneNumber: ")
						fmt.Scanln(&phoneNumber)
					}
					authorizer.PhoneNumber <- phoneNumber
				case client.TypeAuthorizationStateWaitCode:
					var code string
					fmt.Print("Enter code: ")
					fmt.Scanln(&code)
					authorizer.Code <- code
				case client.TypeAuthorizationStateWaitPassword:
					var password string
					fmt.Print("Enter password: ")
					fmt.Scanln(&password)
					authorizer.Password <- password
				case client.TypeAuthorizationStateReady:
					return
				}
			}
		}
	}()
	if appid == 0 {
		fmt.Print("Enter appid: ")
		fmt.Scanln(&appid)
	}
	if appHash == "" {
		fmt.Print("Enter appHash: ")
		fmt.Scanln(&appHash)
	}
	authorizer.TdlibParameters <- &client.TdlibParameters{
		ApiId:                  appid,
		ApiHash:                appHash,
		SystemLanguageCode:     "en",
		DeviceModel:            "SEND",
		ApplicationVersion:     "1.8.0",
		DatabaseDirectory:      "./tdlib-db",
		FilesDirectory:         "./tdlib-files",
		UseFileDatabase:        true,
		UseChatInfoDatabase:    true,
		UseMessageDatabase:     true,
		EnableStorageOptimizer: true,
	}
	if cli, err = client.NewClient(authorizer, client.WithLogVerbosity(&client.SetLogVerbosityLevelRequest{NewVerbosityLevel: 0})); err != nil {
		panic(err)
	}
	go func() {
		listener := cli.GetListener()
		defer listener.Close()
		for update := range listener.Updates {
			if update.GetClass() == client.ClassUpdate && update.GetType() == client.TypeUpdateChatLastMessage {
				msg := update.(*client.UpdateChatLastMessage)
				if chatID == msg.ChatId && chatID == msg.LastMessage.SenderId.(*client.MessageSenderUser).UserId {
					content := msg.LastMessage.Content
					if data, ok := content.(*client.MessageText); ok {
						fmt.Println(data.Text.Text)
					} else if data, ok := content.(*client.MessagePhoto); ok {
						fmt.Println(data.Photo)
					}
				}
			}
		}
	}()
	cli.SendMessage(&client.SendMessageRequest{
		ChatId:              chatID,
		InputMessageContent: &client.InputMessageText{Text: &client.FormattedText{Text: msg}, DisableWebPagePreview: true, ClearDraft: true},
	})
}
