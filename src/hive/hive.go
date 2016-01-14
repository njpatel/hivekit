package hive

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const baseHeader = "x-governess-endpoint"

// Is var for tests
// loginURL = "https://api.hivehome.com/v5/login"
var (
	defaultRefreshInterval = 60 * time.Second
	avoidRetrySeconds      = int64(10)
	loginURL               = "https://api.hivehome.com/v5/login"
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
	refreshing  sync.Mutex

	token   string
	baseURL string
}

// Connect initiates the communication with the Hive Home web service
func Connect(config Config) (*Hive, error) {

	h := &Hive{
		config: config,
	}

	err := h.login()
	if err != nil {
		return nil, err
	}

	go h.startPolling()

	return h, nil
}

func (h *Hive) login() error {
	values := url.Values{}
	values.Set("username", h.config.Username)
	values.Set("password", h.config.Password)

	res, err := http.PostForm(loginURL, values)
	if err != nil {
		return fmt.Errorf("Unable to login: %s", err)
	}

	decoder := json.NewDecoder(res.Body)
	var reply loginReply
	err = decoder.Decode(&reply)

	if err != nil && res.StatusCode == http.StatusOK {
		return fmt.Errorf("Unable to login: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Unable to login: Incorrect status code %s: %s", res.Status, reply.Error.Reason)
	}

	if reply.Token == "" {
		return errors.New("Unable to login: Invalid session token returned")
	}

	h.baseURL = "https://" + strings.Split(res.Header.Get(baseHeader), ":")[0]
	h.token = reply.Token

	return nil
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
	h.refreshing.Lock()
	defer h.refreshing.Unlock()

	now := time.Now().Unix()
	if h.lastRefresh > now-avoidRetrySeconds {
		return
	}

	res, err := h.getHTTP(h.baseURL+"/omnia/nodes", nil)
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

	state := newStateFromNodes(reply.Nodes)
	if err != nil {
		fmt.Printf("Unable to extract state from reply: %s", err)
		return
	}

	fmt.Printf("%v\n", state.CurrentTemp)

	h.lastRefresh = now
}

func (h *Hive) getHTTP(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Omnia-Access-Token", h.token)
	req.Header.Add("Accept", "application/vnd.alertme.zoo-6.1+json")
	req.Header.Add("X-Omnia-Client", "HiveKit")

	client := &http.Client{}
	res, err := client.Do(req)

	// It's possible our token is no longer valid
	if res.StatusCode == http.StatusUnauthorized {
		h.login()
		req.Header.Set("X-Omnia-Access-Token", h.token)
		res, err = client.Do(req)
	}
	return res, err
}
