package hive

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	//loginURL   = "https://api.hivehome.com/v5/login"
	loginURL   = "https://api-hivehome-com-k2b7u8lhpl4e.runscope.net/v5/login"
	baseHeader = "x-governess-endpoint"
)

// Config holds configuration information for the Hive connection
type Config struct {
	Username string
	Password string
}

// Hive contains the data structures to communicate with the Hive Home web service
type Hive struct {
	config Config

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

	if err != nil {
		return nil, fmt.Errorf("Unable to login: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unable to login: Incorrect status code %s: %s", res.Status, reply.Error.Reason)
	}

	if reply.Token == "" {
		return nil, errors.New("Unable to login: Invalid session token returned")
	}

	baseURL := res.Header.Get(baseHeader)
	fmt.Println("hello", baseURL, reply.Token)

	return &Hive{
		config:  config,
		token:   reply.Token,
		baseURL: "https://" + baseURL,
	}, nil
}
