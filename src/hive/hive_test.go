package hive

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HiveTestSuite struct {
	suite.Suite
}

func TestHiveTestSuite(t *testing.T) {
	suite.Run(t, new(HiveTestSuite))
}

func (suite *HiveTestSuite) TestConnectRequest() {
	assert := assert.New(suite.T())
	username := "bobby@charlton.com"
	password := "england66"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		assert.Equal(username, r.PostFormValue("username"))
		assert.Equal(password, r.PostFormValue("password"))
		w.Write([]byte("computersaysno"))
	}))
	defer ts.Close()

	loginURL = ts.URL

	Connect(Config{
		Username: username,
		Password: password,
	})
}

func (suite *HiveTestSuite) TestConnectSuccess() {
	assert := assert.New(suite.T())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		str, _ := json.Marshal(loginReply{
			Token: "kn32j4ioj23j4i2j4",
		})
		w.Write(str)
	}))
	defer ts.Close()

	loginURL = ts.URL

	h, err := Connect(Config{
		Username: "glenn@hoddle.com",
		Password: "magicIsReal",
	})
	assert.NotNil(h)
	assert.Nil(err)
}

func (suite *HiveTestSuite) TestConnectFailedRequest() {
	assert := assert.New(suite.T())
	reason := "incorrect username or password"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		str, _ := json.Marshal(loginReply{
			Error: errorReply{
				Reason: reason,
			},
		})
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(str)
	}))
	defer ts.Close()

	loginURL = ts.URL

	h, err := Connect(Config{
		Username: "glenn@hoddle.com",
		Password: "magicIsReal",
	})
	assert.Nil(h)
	assert.Contains(err.Error(), reason)
}

func (suite *HiveTestSuite) TestConnectFailedServer() {
	assert := assert.New(suite.T())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	loginURL = ts.URL

	h, err := Connect(Config{
		Username: "glenn@hoddle.com",
		Password: "magicIsReal",
	})
	assert.Nil(h)
	assert.Contains(err.Error(), "Internal")
}
