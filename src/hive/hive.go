package hive

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const baseHeader = "x-governess-endpoint"

// Is var for tests
// loginURL = "https://api.hivehome.com/v5/login"
var (
	defaultRefreshInterval = 60 * time.Second
	avoidRetrySeconds      = int64(10)
	loginURL               = "https://api-hivehome-com-k2b7u8lhpl4e.runscope.net/v5/login"
)

// Config holds configuration information for the Hive connection
type Config struct {
	Username        string
	Password        string
	RefreshInterval time.Duration
}

// Hive contains the data structures to communicate with the Hive Home web service
type Hive struct {
	config Config

	ticker      *time.Ticker
	lastRefresh int64 // epoch

	token   string
	baseURL string
}

// Connect initiates the communication with the Hive Home web service
func Connect(config Config) (*Hive, error) {
	values := url.Values{}
	values.Set("username", config.Username)
	values.Set("password", config.Password)

	res, err := http.PostForm(loginURL, values)
	if err != nil {
		return nil, fmt.Errorf("Unable to login: %s", err)
	}

	decoder := json.NewDecoder(res.Body)
	var reply loginReply
	err = decoder.Decode(&reply)

	if err != nil && res.StatusCode == http.StatusOK {
		return nil, fmt.Errorf("Unable to login: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unable to login: Incorrect status code %s: %s", res.Status, reply.Error.Reason)
	}

	if reply.Token == "" {
		return nil, errors.New("Unable to login: Invalid session token returned")
	}

	h := &Hive{
		config:  config,
		token:   reply.Token,
		baseURL: "https://" + strings.Split(res.Header.Get(baseHeader), ":")[0],
	}

	go h.startPolling()

	return h, nil
}

func (h *Hive) startPolling() {
	if h.config.RefreshInterval < 1 {
		h.config.RefreshInterval = defaultRefreshInterval
	}

	h.ticker = time.NewTicker(h.config.RefreshInterval)
	go h.getStatus()
	for {
		<-h.ticker.C
		go h.getStatus()
	}
}

func (h *Hive) getStatus() {
	now := time.Now().Unix()
	if h.lastRefresh > now-avoidRetrySeconds {
		return
	}

	h.lastRefresh = now

	req, err := http.NewRequest("GET", h.baseURL+"/omnia/nodes", nil)
	req.Header.Add("X-Omnia-Access-Token", h.token)
	req.Header.Add("Accept", "application/vnd.alertme.zoo-6.1+json")
	req.Header.Add("X-Omnia-Client", "HiveKit")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Unable to get nodes info: %s", err)
		return
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var reply nodesReply
	err = decoder.Decode(&reply)

	if err != nil {
		fmt.Printf("Unable to get nodes info: %s", err)
		return
	}

	for _, info := range reply.Nodes {
		fmt.Println(info.ID)
	}
}
