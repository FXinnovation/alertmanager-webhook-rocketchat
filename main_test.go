package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ConfigDataTest struct {
	input    Config
	expected error
}

var valuesCheckConfig = []ConfigDataTest{
	{
		input: Config{
			url.URL{
				Host:   "rocket.chat",
				Scheme: "https",
			},
			models.UserCredentials{
				Name:     "john",
				Email:    "123@123",
				Password: "1234",
			},
			map[string]string{},
			ChannelInfo{
				DefaultChannelName: "default",
			},
		},
		expected: nil,
	},
	{
		input: Config{
			url.URL{
				Scheme: "https",
			},
			models.UserCredentials{
				Name:     "john",
				Email:    "123@123",
				Password: "1234",
			},
			map[string]string{},
			ChannelInfo{
				DefaultChannelName: "default",
			},
		},
		expected: errors.New("rocket.chat host not provided"),
	},
	{
		input: Config{
			url.URL{
				Host: "rocket.chat",
			},
			models.UserCredentials{
				Name:     "john",
				Email:    "123@123",
				Password: "1234",
			},
			map[string]string{},
			ChannelInfo{
				DefaultChannelName: "default",
			},
		},
		expected: errors.New("rocket.chat scheme not provided"),
	},
	{
		input: Config{
			url.URL{
				Host:   "rocket.chat",
				Scheme: "https",
			},
			models.UserCredentials{
				Email:    "123@123",
				Password: "1234",
			},
			map[string]string{},
			ChannelInfo{
				DefaultChannelName: "default",
			},
		},
		expected: errors.New("rocket.chat name not provided"),
	},
	{
		input: Config{
			url.URL{
				Host:   "rocket.chat",
				Scheme: "https",
			},
			models.UserCredentials{
				Name:     "john",
				Password: "1234",
			},
			map[string]string{},
			ChannelInfo{
				DefaultChannelName: "default",
			},
		},
		expected: errors.New("rocket.chat email not provided"),
	},
	{
		input: Config{
			url.URL{
				Host:   "rocket.chat",
				Scheme: "https",
			},
			models.UserCredentials{
				Name:  "john",
				Email: "123@123",
			},
			map[string]string{},
			ChannelInfo{
				DefaultChannelName: "default",
			},
		},
		expected: errors.New("rocket.chat password not provided"),
	},
}

type MockedClient struct {
	mock.Mock
}

func TestReadRequestBodyOk(t *testing.T) {

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test_param_warning.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()

	dataReq, _ := readRequestBody(req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusOK)
	}

	// Check the returned data
	if reflect.DeepEqual(template.Data{}, dataReq) {
		t.Error("Struct shouldn't be empty")
	}
}

func TestReadRequestBodyError(t *testing.T) {
	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", nil)

	dataReq, err := readRequestBody(req)

	// Check the returned data
	if !reflect.DeepEqual(template.Data{}, dataReq) {
		t.Error("Struct should be empty")
	}

	// Check the response body
	expected := "EOF"
	if err.Error() != expected {
		t.Errorf("Unexpected body: got %v, want %v", err.Error(), expected)
	}
}

func initMockMessage(text, attachmentText, color, channelName string) {

	rocketChatMock := new(MockedClient)
	rocketChatClient = rocketChatMock

	rocketChatMock.On("WrapperGetChannelID", channelName).Return("test123")
	channel := &models.Channel{ID: "test123"}
	message := &models.Message{
		ID:     "123",
		RoomID: channel.ID,
		Msg:    text,
		PostMessage: models.PostMessage{
			Attachments: []models.Attachment{
				{
					Color: color,
					Text:  attachmentText,
				},
			},
		},
	}
	rocketChatMock.On("WrapperSendMessage", message).Return(message)

	*configFile = "config/rocketchat_example.yml"
	config = loadConfig(*configFile)
	user := &models.User{ID: "123", Name: "prometheus"}
	rocketChatMock.On("WrapperLogin", config).Return(user)
}

func TestWebhookHandlerWarning(t *testing.T) {

	text := "**[ firing ] something_happened from admins at 2019-03-14 17:05:37.903 +0000 UTC**"
	attachmentText := `**alertname**: something_happened
**env**: prod
**instance**: server01.int:9100
**job**: node
**service**: prometheus_bot
**severity**: warning
**supervisor**: runit
**summary**: Oops, something happened!
`
	color := "<warning_color_hexcode>"
	channelName := "prometheus-test-room"

	initMockMessage(text, attachmentText, color, channelName)

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test_param_warning.json")
	if err != nil {
		t.Fatal(err)
	}

	assertWebhookHandler(t, data)
}

func TestWebhookHandlerCritical(t *testing.T) {
	text := "**[ firing ] something_happened from admins at 2019-03-14 17:05:37.903 +0000 UTC**"
	attachmentText := `**alertname**: something_happened
**env**: prod
**instance**: server01.int:9100
**job**: node
**service**: prometheus_bot
**severity**: critical
**supervisor**: runit
**summary**: Oops, something happened!
`
	color := "<critical_color_hexcode>"
	channelName := "prometheus-test-room"

	initMockMessage(text, attachmentText, color, channelName)

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test_param_critical.json")
	if err != nil {
		t.Fatal(err)
	}

	assertWebhookHandler(t, data)
}

func TestWebhookHandlerUndefined(t *testing.T) {
	text := "**[ firing ] something_happened from admins at 2019-03-14 17:05:37.903 +0000 UTC**"
	attachmentText := `**alertname**: something_happened
**env**: prod
**instance**: server01.int:9100
**job**: node
**service**: prometheus_bot
**severity**: critic
**supervisor**: runit
**summary**: Oops, something happened!
`
	color := "#ffffff"
	channelName := "<default_channel_name>"

	initMockMessage(text, attachmentText, color, channelName)

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test_param_undefined.json")
	if err != nil {
		t.Fatal(err)
	}

	assertWebhookHandler(t, data)
}

func assertWebhookHandler(t *testing.T, data []byte) {

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `{"Status":200,"Message":"Success"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestWebhookHandlerError(t *testing.T) {
	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", nil)

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusBadRequest)
	}

	// Check the response body
	expected := `{"Status":400,"Message":"EOF"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestCheckConfig(t *testing.T) {

	for _, d := range valuesCheckConfig {
		configStatus := checkConfig(d.input)
		assert.Equal(t, d.expected, configStatus)
	}
}

func (mock *MockedClient) WrapperGetChannelID(channelName string) (string, error) {
	args := mock.Called(channelName)
	return args.String(0), nil
}

func (mock *MockedClient) WrapperSendMessage(message *models.Message) (*models.Message, error) {
	args := mock.Called(message)
	return args.Get(0).(*models.Message), nil
}

func (mock *MockedClient) WrapperNewMessage(channel *models.Channel, text string) *models.Message {
	return &models.Message{
		ID:     "123",
		RoomID: channel.ID,
		Msg:    text,
	}
}

func (mock *MockedClient) WrapperLogin(credentials *models.UserCredentials) (*models.User, error) {
	args := mock.Called(&config.Credentials)
	return args.Get(0).(*models.User), nil
}
