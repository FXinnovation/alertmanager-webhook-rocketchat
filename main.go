package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	configFile       = kingpin.Flag("config.file", "RocketChat configuration file.").Default("config/rocketchat.yml").String()
	listenAddress    = kingpin.Flag("listen.address", "The address to listen on for HTTP requests.").Default(":9876").String()
	config           Config
	rocketChatClient RocketChatClient
)

// JSONResponse is the webhook http response
type JSONResponse struct {
	Status  int
	Message string
}

func webhook(w http.ResponseWriter, r *http.Request) {

	data, err := readRequestBody(r)
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var errAuthentication error

	errSend := retry(1, 2*time.Second, func() (err error) {
		errSend := SendNotification(rocketChatClient, data, config)
		if errSend != nil {
			errAuthentication = AuthenticateRocketChatClient(rocketChatClient, config)
		}

		if errAuthentication != nil {
			log.Printf("Error authenticating RocketChat client: %v", errAuthentication)
			log.Print("No notification was sent")
		}

		return errSend

	})

	if errSend != nil {
		log.Printf("Error sending notifications to RocketChat : %v", errSend)
		// Returns a 403 if the user can't authenticate
		sendJSONResponse(w, http.StatusUnauthorized, errAuthentication.Error())
	} else {
		// Returns a 200 if everything went smoothly
		sendJSONResponse(w, http.StatusOK, "Success")
	}
}

func retry(retries int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return nil
		}

		if i >= retries {
			break
		}

		time.Sleep(sleep)

		log.Println("retrying after error:", err)
	}
	return fmt.Errorf("after %d retries, last error: %s", retries, err)
}

// Starts 2 listeners
// - one to give a status on the receiver itself
// - one to actually process the data
func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	config = loadConfig(*configFile)

	var errClient error
	rocketChatClient, errClient = GetRocketChatClient(config)
	if errClient != nil {
		log.Fatalf("Error getting RocketChat client: %v", errClient)
	}

	errAuthentication := AuthenticateRocketChatClient(rocketChatClient, config)
	if errAuthentication != nil {
		log.Printf("Error authenticating RocketChat client: %v", errAuthentication)
	}

	http.HandleFunc("/webhook", webhook)
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("listening on: %v", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func sendJSONResponse(w http.ResponseWriter, status int, message string) {
	data := JSONResponse{
		Status:  status,
		Message: message,
	}

	w.WriteHeader(status)

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error writing body: %v", err.Error())
	} else {
		w.Write(bytes)
	}
}

func readRequestBody(r *http.Request) (template.Data, error) {

	// Do not forget to close the body at the end
	defer r.Body.Close()

	// Extract data from the body in the Data template provided by AlertManager
	data := template.Data{}
	err := json.NewDecoder(r.Body).Decode(&data)

	return data, err
}

func loadConfig(configFile string) Config {
	config := Config{}

	// Load the config from the file
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	errYAML := yaml.Unmarshal([]byte(configData), &config)
	if errYAML != nil {
		log.Fatalf("Error: %v", errYAML)
	}

	return config

}
