package main

import (
	"fmt"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/common/log"
	"strings"
)

const (
	defaultColor       = "#ffffff"
	severityLabel      = "severity"
	titleFormat        = "**[ %s ] %s from %s at %s**"
	attachmentFormat   = "**%s**: %s\n"
	alertNameFieldName = "alertname"
)

// RocketChatClient is the client interface to Rocket.Chat
type RocketChatClient interface {
	WrapperLogin(credentials *models.UserCredentials) (*models.User, error)
	WrapperGetChannelID(channelName string) (string, error)
	WrapperSendMessage(message *models.Message) (*models.Message, error)
	WrapperNewMessage(channel *models.Channel, text string) *models.Message
}

// RocketChatConnector connector and method base
type RocketChatConnector struct {
	Client *realtime.Client
}

// WrapperLogin wraps the Login method
func (connector RocketChatConnector) WrapperLogin(credentials *models.UserCredentials) (*models.User, error) {
	return connector.Client.Login(credentials)
}

// WrapperGetChannelID wraps the GetChannelId method
func (connector RocketChatConnector) WrapperGetChannelID(channelName string) (string, error) {
	return connector.Client.GetChannelId(channelName)
}

// WrapperSendMessage wraps SendMessage method
func (connector RocketChatConnector) WrapperSendMessage(message *models.Message) (*models.Message, error) {
	return connector.Client.SendMessage(message)
}

// WrapperNewMessage wraps the NewMessage method
func (connector RocketChatConnector) WrapperNewMessage(channel *models.Channel, text string) *models.Message {
	return connector.Client.NewMessage(channel, text)
}

// GetRocketChatClient returns the RocketChatClient
func GetRocketChatClient() (RocketChatConnector, error) {

	rtClient, errClient := realtime.NewClient(&config.Endpoint, false)
	if errClient != nil {
		return RocketChatConnector{}, errClient
	}

	return RocketChatConnector{Client: rtClient}, nil

}

// AuthenticateRocketChatClient performs login on the client
func AuthenticateRocketChatClient(connector RocketChatClient) error {
	_, errUser := connector.WrapperLogin(&config.Credentials)
	return errUser
}

func formatMessage(connector RocketChatClient, channel *models.Channel, alert template.Alert, receiver string) *models.Message {
	severity := alert.Labels[severityLabel]

	title := fmt.Sprintf(titleFormat, alert.Status, alert.Labels[alertNameFieldName], receiver, alert.StartsAt)
	message := connector.WrapperNewMessage(channel, title)

	var usedColor string
	if color, colorExists := config.SeverityColors[severity]; colorExists {
		usedColor = color
	} else {
		usedColor = defaultColor
	}

	var attachmentBuilder strings.Builder

	for _, label := range alert.Labels.SortedPairs() {
		attachmentBuilder.WriteString(fmt.Sprintf(attachmentFormat, label.Name, label.Value))
	}
	for _, annotation := range alert.Annotations.SortedPairs() {
		attachmentBuilder.WriteString(fmt.Sprintf(attachmentFormat, annotation.Name, annotation.Value))
	}

	message.PostMessage.Attachments = []models.Attachment{
		{
			Color: usedColor,
			Text:  attachmentBuilder.String(),
		},
	}

	return message
}

// SendNotification connects to RocketChat server, authenticates the user and sends the notification
func SendNotification(connector RocketChatClient, data template.Data) error {

	channelName := config.Channel.DefaultChannelName
	if val, ok := data.CommonLabels["channel_name"]; ok {
		channelName = val
	}

	if channelName == "" {
		log.Error("Exception: Channel name not found. Please specify a default_channel_name in the configuration.")
	} else {
		channelID, errRoom := connector.WrapperGetChannelID(channelName)
		if errRoom != nil {
			log.Errorf("Error to get room ID: %v", errRoom)
			return errRoom
		}
		channel := &models.Channel{ID: channelID}

		log.Infof("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
		for _, alert := range data.Alerts {

			message := formatMessage(connector, channel, alert, data.Receiver)
			_, errMessage := connector.WrapperSendMessage(message)
			if errMessage != nil {
				log.Infof("Error to send message: %v", errMessage)
				return errMessage
			}
		}
	}
	return nil
}
