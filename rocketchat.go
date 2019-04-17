package main

import (
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/prometheus/alertmanager/template"
)

// ChannelInfo - Channel configuration
type ChannelInfo struct {
	DefaultChannelName string `yaml:"default_channel_name"`
}

// Config - Rocket.Chat webhook configuration
type Config struct {
	Endpoint       url.URL                `yaml:"endpoint"`
	Credentials    models.UserCredentials `yaml:"credentials"`
	SeverityColors map[string]string      `yaml:"severity_colors"`
	Channel        ChannelInfo            `yaml:"channel"`
}

// RocketChatClient is the client interface to Rocket.Chat
type RocketChatClient interface {
	Login(credentials *models.UserCredentials) (*models.User, error)
	GetChannelId(name string) (string, error)
	SendMessage(message *models.Message) (*models.Message, error)
	NewMessage(channel *models.Channel, text string) *models.Message
}

// GetRocketChatClient returns the RocketChatClient
func GetRocketChatClient(config Config) (*realtime.Client, error) {

	rtClient, errClient := realtime.NewClient(&config.Endpoint, false)
	if errClient != nil {
		return nil, errClient
	}

	return rtClient, nil

}

// AuthenticateRocketChatClient performs login on the client
func AuthenticateRocketChatClient(rtClient RocketChatClient, config Config) error {
	_, errUser := rtClient.Login(&config.Credentials)
	return errUser
}

func formatMessage(rtClient RocketChatClient, channel *models.Channel, alert template.Alert, config Config) *models.Message {

	const (
		defaultColor = "#ffffff"
	)

	severity := alert.Labels["severity"]
	title := fmt.Sprintf("**[%s] %s: %s**", alert.Status, severity, alert.Annotations["summary"])
	message := rtClient.NewMessage(channel, title)

	color := defaultColor
	for k, v := range config.SeverityColors {
		if k == strings.ToLower(severity) {
			color = v
			break
		}
	}

	var keys []string
	for k := range alert.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	attachementText := fmt.Sprintf("**description**: %s\n**alert_timestamp**: %s\n", alert.Annotations["description"], alert.StartsAt)
	for _, k := range keys {
		attachementText += fmt.Sprintf("**%s**: %s\n", k, alert.Labels[k])
	}

	message.PostMessage.Attachments = []models.Attachment{
		models.Attachment{
			Color: color,
			Text:  attachementText,
		},
	}

	return message
}

// SendNotification connects to RocketChat server, authenticates the user and sends the notification
func SendNotification(rtClient RocketChatClient, data template.Data, config Config) error {

	var channelName string
	if val, ok := data.CommonLabels["channel_name"]; ok {
		channelName = val
	} else {
		channelName = config.Channel.DefaultChannelName
	}

	if channelName == "" {
		log.Print("Exception: Channel name not found. Please specify a default_channel_name in the configuration.")
	} else {
		channelID, errRoom := rtClient.GetChannelId(channelName)
		if errRoom != nil {
			log.Printf("Error to get room ID: %v", errRoom)
			return errRoom
		}
		channel := &models.Channel{ID: channelID}

		log.Printf("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
		for _, alert := range data.Alerts {

			message := formatMessage(rtClient, channel, alert, config)
			_, errMessage := rtClient.SendMessage(message)
			if errMessage != nil {
				log.Printf("Error to send message: %v", errMessage)
				return errMessage
			}
		}
	}
	return nil
}
