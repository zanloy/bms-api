package models // import github.com/zanloy/bms-api/models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/elgs/gojq"
	"github.com/go-resty/resty/v2"
)

type RespType string

const (
	RespTypeHTTPStatus RespType = "httpstatus"
	RespTypeHTTPBody   RespType = "httpbody"
	RespTypeJSON       RespType = "json"
)

var (
	RestyClient *resty.Client
)

func init() {
	RestyClient = resty.New()
}

// A URLCheck will treat a match against RegExp field as healthy unless FailTrue
// is set which inverts the result.
type URLCheck struct {
	Name     string        `json:"name"`
	Desc     string        `json:"description,omitempty"`
	Url      string        `json:"url"`
	Type     RespType      `json:"type"`
	FailTrue bool          `json:"fail_true,omitempty"` // True will invert our result.
	JSONPath string        `json:"jsonpath,omitempty"`  // Used for json Type.
	RegExp   string        `json:"regexp,omitempty"`
	Date     time.Time     `json:"date,omitempty"`
	Healthy  HealthyStatus `json:"healthy"`
	Text     string        `json:"text,omitempty"`
	Errors   []string      `json:"errors,omitempty"`
}

// URLClientOpts is a struct to hold all configuration options needed for
// a resty client to make the connection to check the health of a URL. This
// is still a work in progress and not all options are currently implemented.
type URLClientOpts struct {
	Headers map[string]string
	Cert    string
	Key     string
}

// Check will verify health of the URLCheck and populate the relevent fields.
func (uc *URLCheck) Check() {
	// Reset any previous results
	uc.Date = time.Now()
	uc.Healthy = StatusUnknown
	uc.Errors = make([]string, 0)

	// nil check client
	if RestyClient == nil {
		uc.Errors = append(uc.Errors, "resty client has not been initialized")
		return
	}

	// Make request
	req := RestyClient.NewRequest()
	if resp, err := req.Get(uc.Url); err == nil {
		uc.checkResponse(resp)
	} else {
		uc.Errors = append(uc.Errors, err.Error())
	}
}

// CheckHTTPBody will take the response body and check our RegExp against it.
func (uc *URLCheck) checkHTTPBody(resp *resty.Response) {
	// Check RegExp
	if uc.RegExp == "" {
		uc.Errors = append(uc.Errors, "RegExp can not be null.")
		return
	}

	body := resp.Body()

	uc.Healthy, uc.Errors = checkValidity(string(body), uc.RegExp)
	switch uc.Healthy {
	case StatusHealthy:
		uc.Text = "healthy"
	case StatusUnhealthy:
		uc.Text = "unhealthy"
	}
	return
}

// CheckHTTPStatus is the default validator. It will use sane defaults saying
// that any HTTP status code that is !4xx and !5xx is considered healthy. It
// will use the "RegExp" field if you want a custom check.
func (uc *URLCheck) checkHTTPStatus(resp *resty.Response) {
	// Load RegExp
	var expr string
	if uc.RegExp == "" {
		expr = `^[^(4|5)]\d\d` // Check for any non 4xx/5xx status codes
	} else {
		expr = uc.RegExp
	}

	uc.Healthy, uc.Errors = checkValidity(resp.Status(), expr)
	uc.Text = resp.Status()
}

// CheckJSON will attempt to parse the response body into json. If RegExp is
// empty then we only validate that the field exists in the response.
func (uc *URLCheck) checkJSON(resp *resty.Response) {
	// Check Path
	if uc.JSONPath == "" {
		uc.Errors = append(uc.Errors, "Path can not be null for a json check.")
		return
	}

	// Load RegExp
	var expr string
	if uc.RegExp == "" {
		expr = "."
	} else {
		expr = uc.RegExp
	}

	// Parse JSON body
	body := resp.Body()

	parser, err := gojq.NewStringQuery(string(body))
	if err != nil {
		uc.Errors = append(uc.Errors, err.Error())
		return
	}

	raw, err := parser.Query(uc.JSONPath)
	if err != nil {
		uc.Errors = append(uc.Errors, err.Error())
		return
	}

	// We need to process booleans like string for our comparison
	var val string // To hold our stringified result
	switch typed := raw.(type) {
	case string:
		val = typed
	case bool:
		if typed == true {
			val = "true"
		} else {
			val = "false"
		}
	case int, int8, int16, int32, int64, float32, float64:
		val = fmt.Sprint(typed)
	default:
		// We don't know what to do with this.
		uc.Errors = append(uc.Errors, "JSON value type assertion failure. Parser can only handle basic type of string, bool, and int/float.")
		return
	}

	uc.Healthy, uc.Errors = checkValidity(val, expr)
	uc.Text = val
}

// checkResponse will look at the check and use the appropriate validator
// based on the Type field.
func (uc *URLCheck) checkResponse(resp *resty.Response) {
	switch uc.Type {
	case RespTypeHTTPBody:
		uc.checkHTTPBody(resp)
	case RespTypeJSON:
		uc.checkJSON(resp)
	default:
		uc.checkHTTPStatus(resp)
	}

	// Check if we should fail true
	if uc.FailTrue {
		switch uc.Healthy {
		case StatusHealthy:
			uc.Healthy = StatusUnhealthy
		case StatusUnhealthy:
			uc.Healthy = StatusHealthy
		}
	}
}

// checkValidity takes a value and a regexp and return a HealthyStatus if the
// value matches the regexp.
func checkValidity(value string, expr string) (healthy HealthyStatus, errors []string) {
	healthy = StatusUnknown
	errors = make([]string, 0)

	// Validate RegExp
	re, err := regexp.Compile(expr)
	if err != nil {
		errors = append(errors, fmt.Sprintf("RegExp [%s] failed syntax check.", expr))
		return
	}

	// Check for match
	if re.MatchString(value) {
		healthy = StatusHealthy
	} else {
		healthy = StatusUnhealthy
	}

	return
}
