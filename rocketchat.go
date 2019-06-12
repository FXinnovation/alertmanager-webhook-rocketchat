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

// RocketChat is the client interface to Rocket.Chat
type RocketChat interface {
	Login(credentials *models.UserCredentials) (*models.User, error)
	GetChannelID(channelName string) (string, error)
	SendMessage(message *models.Message) (*models.Message, error)
	NewMessage(channel *models.Channel, text string) *models.Message
}

// RocketChatConnector connector and method base
type RocketChatConnector struct {
	Client *realtime.Client
}

// Login wraps the Login method
func (connector RocketChatConnector) Login(credentials *models.UserCredentials) (*models.User, error) {
	return connector.Client.Login(credentials)
}

// GetChannelID wraps the GetChannelId method
func (connector RocketChatConnector) GetChannelID(channelName string) (string, error) {
	return connector.Client.GetChannelId(channelName)
}

// SendMessage wraps SendMessage method
func (connector RocketChatConnector) SendMessage(message *models.Message) (*models.Message, error) {
	return connector.Client.SendMessage(message)
}

// NewMessage wraps the NewMessage method
func (connector RocketChatConnector) NewMessage(channel *models.Channel, text string) *models.Message {
	return connector.Client.NewMessage(channel, text)
}

// GetRocketChat returns the RocketChat
func GetRocketChat() (RocketChatConnector, error) {

	rtClient, errClient := realtime.NewClient(&config.Endpoint, false)
	if errClient != nil {
		return RocketChatConnector{}, errClient
	}

	return RocketChatConnector{Client: rtClient}, nil

}

// AuthenticateRocketChatClient performs login on the client
func AuthenticateRocketChatClient(connector RocketChat) error {
	_, errUser := connector.Login(&config.Credentials)
	return errUser
}

func formatMessage(connector RocketChat, channel *models.Channel, alert template.Alert, receiver string) *models.Message {
	severity := alert.Labels[severityLabel]

	title := fmt.Sprintf(titleFormat, alert.Status, alert.Labels[alertNameFieldName], receiver, alert.StartsAt)
	message := connector.NewMessage(channel, title)

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
func SendNotification(connector RocketChat, data template.Data) error {

	channelName := config.Channel.DefaultChannelName
	if val, ok := data.CommonLabels["channel_name"]; ok {
		channelName = val
	}

	if channelName == "" {
		log.Error("Exception: Channel name not found. Please specify a default_channel_name in the configuration.")
	} else {
		channelID, errRoom := connector.GetChannelID(channelName)
		if errRoom != nil {
			log.Errorf("Error to get room ID: %v", errRoom)
			return errRoom
		}
		channel := &models.Channel{ID: channelID}

		log.Infof("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
		for _, alert := range data.Alerts {

			message := formatMessage(connector, channel, alert, data.Receiver)
			_, errMessage := connector.SendMessage(message)
			if errMessage != nil {
				log.Infof("Error to send message: %v", errMessage)
				return errMessage
			}
		}
	}
	return nil
}
