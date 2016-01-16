package hive

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/brutella/log"
)

type HiveTestSuite struct {
	suite.Suite
}

func TestHiveTestSuite(t *testing.T) {
	log.Verbose = false
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

func (suite *HiveTestSuite) TestPut() {
	assert := assert.New(suite.T())
	token := "ofmyappreciation"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(token, r.Header.Get(omniaAccessTokenHdr))
		assert.NotEmpty(r.Header.Get(omniaClientHdr))
		assert.NotEmpty(r.Header.Get(httpContentHdr))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h := &Hive{
		token: token,
	}

	b, _ := json.Marshal(nodeReportInt{TargetValue: 1})
	res, err := h.putHTTP(ts.URL, b)
	assert.NotNil(res)
	assert.Nil(err)
}

func (suite *HiveTestSuite) TestPutRetries() {
	assert := assert.New(suite.T())
	token := "ofmyappreciation"
	times := 0

	lts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		str, _ := json.Marshal(loginReply{
			Token: token,
		})
		w.Write(str)
	}))
	defer lts.Close()

	loginURL = lts.URL

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		times++
		if times == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		assert.Equal(token, r.Header.Get(omniaAccessTokenHdr))
		assert.NotEmpty(r.Header.Get(omniaClientHdr))
		assert.NotEmpty(r.Header.Get(httpContentHdr))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h := &Hive{
		token: token,
	}

	res, err := h.putHTTP(ts.URL, []byte{'h', 'e', 'l', 'l', 'o'})
	assert.NotNil(res)
	assert.Nil(err)
	assert.Equal(2, times)
}
