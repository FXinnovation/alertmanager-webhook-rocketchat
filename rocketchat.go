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
	defaultColor     = "#ffffff"
	severityLabel    = "severity"
	titleFormat      = "**[ %s ] %s from %s at %s**"
	attachmentFormat = "**%s**: %s\n"
)

// ChannelInfo - Channel configuration
type ChannelInfo struct {
	DefaultChannelName string `yaml:"default_channel_name"`
}

// RocketChatClient is the client interface to Rocket.Chat
type RocketChatClient interface {
	Login(credentials *models.UserCredentials) (*models.User, error)
	GetChannelId(name string) (string, error)
	SendMessage(message *models.Message) (*models.Message, error)
	NewMessage(channel *models.Channel, text string) *models.Message
}

// GetRocketChatClient returns the RocketChatClient
func GetRocketChatClient() (*realtime.Client, error) {

	rtClient, errClient := realtime.NewClient(&config.Endpoint, false)
	if errClient != nil {
		return nil, errClient
	}

	return rtClient, nil

}

// AuthenticateRocketChatClient performs login on the client
func AuthenticateRocketChatClient(rtClient RocketChatClient) error {
	_, errUser := rtClient.Login(&config.Credentials)
	return errUser
}

func formatMessage(rtClient RocketChatClient, channel *models.Channel, alert template.Alert, receiver string) *models.Message {
	severity := alert.Labels[severityLabel]
	title := fmt.Sprintf(titleFormat, alert.Status, alert.Labels["alertname"], receiver, alert.StartsAt)
	message := rtClient.NewMessage(channel, title)

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
func SendNotification(rtClient RocketChatClient, data template.Data) error {

	var channelName string
	if val, ok := data.CommonLabels["channel_name"]; ok {
		channelName = val
	} else {
		channelName = config.Channel.DefaultChannelName
	}

	if channelName == "" {
		log.Error("Exception: Channel name not found. Please specify a default_channel_name in the configuration.")
	} else {
		channelID, errRoom := rtClient.GetChannelId(channelName)
		if errRoom != nil {
			log.Errorf("Error to get room ID: %v", errRoom)
			return errRoom
		}
		channel := &models.Channel{ID: channelID}

		log.Infof("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
		for _, alert := range data.Alerts {

			message := formatMessage(rtClient, channel, alert, data.Receiver)
			_, errMessage := rtClient.SendMessage(message)
			if errMessage != nil {
				log.Infof("Error to send message: %v", errMessage)
				return errMessage
			}
		}
	}
	return nil
}
