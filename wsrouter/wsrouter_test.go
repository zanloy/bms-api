package wsrouter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/zanloy/bms-api/models"
	"gopkg.in/olahol/melody.v1"
)

type handleMessageFunc func(*melody.Session, []byte)

type testServer struct {
	m *melody.Melody
}

func newTestServerHandler(handler handleMessageFunc) *testServer {
	m := melody.New()
	m.HandleMessage(handler)
	return &testServer{m: m}
}

func newTestServer() *testServer {
	m := melody.New()
	return &testServer{m: m}
}

func (ts *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ts.m.HandleRequest(w, r)
}

func newDialer(url string) (*websocket.Conn, error) {
	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(strings.Replace(url, "http", "ws", 1), nil)
	return conn, err
}

func TestLoadFilters(t *testing.T) {
	var testData = []models.Filter{{
		Action: "ignore",
		Kind:   "daemonset",
		Name:   "daemonset3",
	}, {
		Action: "MUTE",
		Kind:   "DEPLOYMENT",
		Name:   "DEPLOYMENT1",
	}, {
		Action: "mute",
		Kind:   "NameSpace",
		Name:   "namespace->1",
	}}

	var badData = []models.Filter{{
		Action: "badActor",
		Kind:   "daemonset",
		Name:   "daemonset3",
	}, {
		Action: "mute",
		Kind:   "badKind",
		Name:   "yanni",
	}, {
		Action: "mute",
		Kind:   "",
		Name:   "laurel",
	}}

	viper.Set("")
	LoadFilters(testData)

	// Make sure our filters were put in the correct location.
	for _, testCase := range testData {
		assert.Contains(t, filters, strings.ToLower(testCase.Kind))
		assert.Contains(t, filters[strings.ToLower(testCase.Kind)], testCase.Name, "filters[kind] must contain %s", testCase.Name)
	}

	// anticases
	LoadFilters(badData)
	assert.NotContains(t, filters, "socrates", "uninitialized key should not exist.")
	assert.NotContains(t, filters["daemonset"], "daemonset666", "uninitialized key should not exits.")
	assert.NotContains(t, filters, "badKind", "invalid kind should be ignored.")
}

/*
func TestBroadcast(t *testing.T) {
	// Setup filters
	var myFilters = []models.Filter{{
		Action: "mute",
		Kind:   "deployment",
		Name:   "ignoreme",
	}, {
		Action: "ignore",
		Kind:   "deployment",
		Name:   "ignoremetooplz",
	}}
	LoadFilters(myFilters)

	// Setup mock http server
	listener := listeners["deployment"]
	update := models.HealthUpdate{
		Kind:    "deployment",
		Name:    "testing",
		Healthy: models.StatusHealthy,
	}

	var jsonedUpdate string
	msg := update.ToMsg()
	if bytes, err := json.Marshal(msg); err != nil {
		require.NoError(t, err, "must be able to marshal test data")
	} else {
		jsonedUpdate = string(bytes)
	}

	err := Broadcast(update)
	assert.NoError(t, err)
	assert.Equal(t, jsonedUpdate, listener.buffer)
}
*/
