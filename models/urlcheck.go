package models // import github.com/zanloy/bms-api/models

import (
	"fmt"
	"regexp"

	"github.com/elgs/gojq"
	"github.com/go-resty/resty/v2"
)

type RespType string

const (
	RespTypeHTTPStatus RespType = "httpstatus"
	RespTypeHTTPBody   RespType = "httpbody"
	RespTypeJSON                = "json"
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
	Type     RespType      `json:"type,omitempty"`
	FailTrue bool          `json:"fail_true,omitempty"` // This will basically just invert our result.
	JSONPath string        `json:"jsonpath,omitempty"`  // This is the jsonpath used for json type.
	RegExp   string        `json:"regexp,omitempty"`
	Report   *HealthReport `json:"report,omitempty"`
}

// URLClientOpts is a struct to hold all configuration options needed for
// a resty client to make the connection to check the health of a URL. This
// is still a work in progress and not all options are currently implemented.
type URLClientOpts struct {
	Headers map[string]string
	Cert    string
	Key     string
}

func (uc *URLCheck) Check() HealthReport {
	report := NewHealthReport()

	// Need to nil check client
	if RestyClient == nil {
		report.Errors = append(report.Errors, "resty client has not been initialized")
		return report
	}

	// Make request
	req := RestyClient.NewRequest()
	if resp, err := req.Get(uc.Url); err == nil {
		uc.CheckResponse(&report, resp)
	} else {
		report.Errors = append(report.Errors, err.Error())
	}

	uc.Report = &report
	return report
}

// CheckHTTPBody will take the response body and check our RegExp against it.
func (uc *URLCheck) CheckHTTPBody(report *HealthReport, resp *resty.Response) {
	// Check RegExp
	if uc.RegExp == "" {
		report.Errors = append(report.Errors, "RegExp can not be null.")
		return
	}

	body := resp.Body()

	report.Healthy, report.Errors = checkValidity(string(body), uc.RegExp)
	switch report.Healthy {
	case StatusHealthy:
		report.Text = "healthy"
	case StatusUnhealthy:
		report.Text = "unhealthy"
	}
	return
}

// CheckHTTPStatus is the default validator. It will use sane defaults saying
// that any HTTP status code that is !4xx and !5xx. It will use the "Valid"
// field if you want a custom check. This will be treated as a comma separated
// list of checks.
func (uc *URLCheck) CheckHTTPStatus(report *HealthReport, resp *resty.Response) {
	// Load RegExp
	var expr string
	if uc.RegExp == "" {
		expr = `^[^(4|5)]\d\d`
	} else {
		expr = uc.RegExp
	}

	report.Text = resp.Status()
	report.Healthy, report.Errors = checkValidity(resp.Status(), expr)
}

// CheckJSON will attempt to parse the response body into json. If RegExp is
// empty then we only validate that the field exists in the response.
func (uc *URLCheck) CheckJSON(report *HealthReport, resp *resty.Response) {
	// Check Path
	if uc.JSONPath == "" {
		report.Errors = append(report.Errors, "Path can not be null.")
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
		report.Errors = append(report.Errors, err.Error())
		return
	}

	raw, err := parser.Query(uc.JSONPath)
	if err != nil {
		report.Errors = append(report.Errors, err.Error())
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
		report.Errors = append(report.Errors, "JSON value type assertion failure. Parser can only handle basic type of string, bool, and int/float.")
		return
	}

	report.Text = val
	report.Healthy, report.Errors = checkValidity(val, expr)
}

func (uc *URLCheck) CheckResponse(report *HealthReport, resp *resty.Response) {
	switch uc.Type {
	case RespTypeHTTPBody:
		uc.CheckHTTPBody(report, resp)
	case RespTypeJSON:
		uc.CheckJSON(report, resp)
	default:
		uc.CheckHTTPStatus(report, resp)
	}

	// Check if we should fail true
	if uc.FailTrue {
		switch report.Healthy {
		case StatusHealthy:
			report.Healthy = StatusUnhealthy
		case StatusUnhealthy:
			report.Healthy = StatusHealthy
		}
	}
}

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
