package hive

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/brutella/log"
)

const (
	baseHdr             = "x-governess-endpoint"
	httpAcceptHdr       = "Accept"
	httpContentHdr      = "Content-Type"
	omniaAccessTokenHdr = "X-Omnia-Access-Token"
	omniaClientHdr      = "X-Omnia-Client"
)

const (
	// MinTemp is the minimum temprature that can be set on the Hive Home unit
	MinTemp = 1.0

	// MaxTemp is the maximum temprature that can be set on the Hive Home unit
	MaxTemp = 32.0
)

// Some are var for tests
var (
	defaultRefreshInterval = 60 * time.Second
	avoidRetrySeconds      = int64(10)
	loginURL               = "https://api.hivehome.com/v5/login"
	nodeURL                = "https://api-prod.bgchprod.info/omnia/nodes/"
)

// Config holds configuration information for the Hive connection
type Config struct {
	Username        string
	Password        string
	RefreshInterval time.Duration
}

// StateChangeFunc is a function type for receiving state change notifications
type StateChangeFunc func(*State)

// Hive contains the data structures to communicate with the Hive Home web service
type Hive struct {
	config Config

	ticker      *time.Ticker
	updateNow   chan bool
	lastRefresh int64 // epoch
	refreshing  sync.RWMutex

	token   string
	baseURL string

	stateChangeHandler StateChangeFunc
	lastState          State
}

// Connect initiates the communication with the Hive Home web service
func Connect(config Config) (*Hive, error) {

	h := &Hive{
		config:    config,
		updateNow: make(chan bool, 10),
	}

	err := h.login()
	if err != nil {
		return nil, err
	}

	go h.startPolling()

	return h, nil
}

// HandleStateChange lets you register a callback to be notified when the Hive Home
// state changes. It is called with the most recent hive.State
func (h *Hive) HandleStateChange(stf StateChangeFunc) {
	h.refreshing.Lock()
	h.stateChangeHandler = stf
	h.refreshing.Unlock()
}

// GetState returns the last known state of the Hive
func (h *Hive) GetState() (s State) {
	h.refreshing.RLock()
	s = h.lastState
	h.refreshing.RUnlock()
	return
}

// SetTargetTemp sets the desired temperature on the Hive. If the Hive heating mode is set to
func (h *Hive) SetTargetTemp(temp float64) error {
	nodes := nodesReply{
		Nodes: []nodeInfo{
			nodeInfo{
				Attributes: nodeAttributes{
					TargetHeatTemperature: &nodeReportFloat{
						TargetValue: temp,
					},
				},
			},
		},
	}

	body, err := json.Marshal(nodes)
	if err != nil {
		return err
	}

	state := h.GetState()
	_, err = h.putHTTP(nodeURL+state.heatingNodeID, body)

	go (func() {
		<-time.After(5 * time.Second)
		h.updateNow <- true
	})()

	return err
}

// SetTargetHeatMode sets the desired heating mode on the Hive
func (h *Hive) SetTargetHeatMode(mode HeatCoolMode) error {
	log.Printf("[INFO] Changing HeatCoolMode is unsupported at this time\n")
	return nil
}

// ToggleHotWater either boosts the hotwater for a duration, or restores it to automatic mode
func (h *Hive) ToggleHotWater(on bool, onForLength time.Duration) error {
	var info nodeInfo
	if on == true {
		info = nodeInfo{
			Attributes: nodeAttributes{
				ActiveHeatCoolMode: &nodeReportString{
					TargetValue: apiBoost,
				},
				ScheduleLockDuration: &nodeReportInt{
					TargetValue: int32(onForLength.Minutes()),
				},
			},
		}
	} else {
		info = nodeInfo{
			Attributes: nodeAttributes{
				ActiveHeatCoolMode: &nodeReportString{
					TargetValue: apiHeat,
				},
				ActiveScheduleLock: &nodeReportBool{
					TargetValue: false,
				},
			},
		}
	}

	nodes := nodesReply{
		Nodes: []nodeInfo{info},
	}

	body, err := json.Marshal(nodes)
	if err != nil {
		return err
	}

	state := h.GetState()
	_, err = h.putHTTP(nodeURL+state.hotWaterNodeID, body)

	go (func() {
		<-time.After(5 * time.Second)
		h.updateNow <- true
	})()

	return err
}

// ToggleHeatingBoost will switch on or off the heating boost for the duration
func (h *Hive) ToggleHeatingBoost(on bool, duration time.Duration) error {
	var info nodeInfo
	if on == true {
		info = nodeInfo{
			Attributes: nodeAttributes{
				ActiveHeatCoolMode: &nodeReportString{
					TargetValue: apiBoost,
				},
				ScheduleLockDuration: &nodeReportInt{
					TargetValue: int32(duration.Minutes()),
				},
			},
		}
	} else {
		info = nodeInfo{
			Attributes: nodeAttributes{
				ActiveHeatCoolMode: &nodeReportString{
					TargetValue: apiHeat,
				},
				ActiveScheduleLock: &nodeReportBool{
					TargetValue: false,
				},
			},
		}
	}

	nodes := nodesReply{
		Nodes: []nodeInfo{info},
	}

	body, err := json.Marshal(nodes)
	if err != nil {
		return err
	}

	state := h.GetState()
	_, err = h.putHTTP(nodeURL+state.heatingNodeID, body)

	go (func() {
		<-time.After(5 * time.Second)
		h.updateNow <- true
	})()
	return err
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

	h.baseURL = "https://" + strings.Split(res.Header.Get(baseHdr), ":")[0]
	h.token = reply.Token

	return nil
}

func (h *Hive) startPolling() {
	if h.config.RefreshInterval < 1 {
		h.config.RefreshInterval = defaultRefreshInterval
	}

	h.ticker = time.NewTicker(h.config.RefreshInterval)
	go h.getStatus(true)
	for {
		select {
		case <-h.ticker.C:
			h.getStatus(false)
		case <-h.updateNow:
			h.getStatus(true)
		}
	}
}

func (h *Hive) getStatus(force bool) {
	log.Printf("[VERB] Refreshing Hive status")
	h.refreshing.Lock()
	defer h.refreshing.Unlock()

	now := time.Now().Unix()
	if force == false && h.lastRefresh > now-avoidRetrySeconds {
		return
	}

	res, err := h.getHTTP(h.baseURL + "/omnia/nodes")
	if err != nil {
		log.Printf("[WARN] Unable to get nodes info: %s", err)
		return
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var reply nodesReply
	err = decoder.Decode(&reply)

	if err != nil {
		log.Printf("[WARN] Unable to get nodes info: %s", err)
		return
	}

	state := newStateFromNodes(reply.Nodes)
	if err != nil {
		log.Printf("[WARN] Unable to extract state from reply: %s", err)
		return
	}

	h.lastState = *state
	if h.stateChangeHandler != nil {
		go h.stateChangeHandler(state)
	}

	h.lastRefresh = now
}

func (h *Hive) getHTTP(url string) (*http.Response, error) {
	req, err := h.createGetRequest(url)
	if err != nil {
		return nil, err
	}

	req.Header.Add(omniaAccessTokenHdr, h.token)
	req.Header.Add(httpAcceptHdr, "application/vnd.alertme.zoo-6.1+json")
	req.Header.Add(omniaClientHdr, "HiveKit")

	client := &http.Client{}
	res, err := client.Do(req)

	// It's possible our token is no longer valid
	if res != nil && res.StatusCode == http.StatusUnauthorized {
		h.login()
		req, _ := h.createGetRequest(url)
		res, err = client.Do(req)
	}
	return res, err
}

func (h *Hive) createGetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(omniaAccessTokenHdr, h.token)
	req.Header.Add(httpAcceptHdr, "application/vnd.alertme.zoo-6.1+json")
	req.Header.Add(omniaClientHdr, "HiveKit")
	return req, nil
}

func (h *Hive) putHTTP(url string, body []byte) (*http.Response, error) {
	req, err := h.createPutRequest(url, body)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Bad Hive!
	}
	client := &http.Client{
		Transport: tr,
	}
	res, err := client.Do(req)

	// It's possible our token is no longer valid
	if res != nil && res.StatusCode == http.StatusUnauthorized {
		h.login()
		req, _ := h.createPutRequest(url, body)
		res, err = client.Do(req)
	}
	return res, err
}

func (h *Hive) createPutRequest(url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add(omniaAccessTokenHdr, h.token)
	req.Header.Add(httpAcceptHdr, "application/vnd.alertme.zoo-6.1+json")
	req.Header.Add(omniaClientHdr, "HiveKit")
	req.Header.Add(httpContentHdr, "application/json")

	return req, nil
}
