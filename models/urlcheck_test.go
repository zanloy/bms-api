package models_test

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "github.com/zanloy/bms-api/models"
)

var jsonPayload = `{
	"booltrue": true,
	"boolfalse": false,
	"integer": 1269,
  "string": "healthy",
	"nested": {
		"light.color": "green",
		"nil": null
	}
}`

var urls = map[string]string{
	"200":            "http://test.test/200",
	"200json":        "http://test.test/200.json",
	"200invalidjson": "http://test.text/200.invalidjson",
	"404":            "http://test.test/404",
	"error":          "http://test.test/error",
}

func TestCheck(t *testing.T) {
	// Setup
	httpmock.ActivateNonDefault(RestyClient.GetClient())
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		urls["200"],
		httpmock.NewStringResponder(200, "This is a text body. Here is some more example text to match."),
	)

	httpmock.RegisterResponder(
		"GET",
		urls["200json"],
		httpmock.NewStringResponder(200, jsonPayload),
	)

	httpmock.RegisterResponder(
		"GET",
		urls["200invalidjson"],
		httpmock.NewStringResponder(200, "this is not json"),
	)

	httpmock.RegisterResponder(
		"GET",
		urls["404"],
		httpmock.NewStringResponder(404, "NOT FOUND"),
	)

	httpmock.RegisterResponder(
		"GET",
		urls["error"],
		httpmock.NewErrorResponder(fmt.Errorf("")),
	)

	testCases := []struct {
		desc    string
		subj    URLCheck
		healthy HealthyStatus
		text    string
		errors  bool
	}{{
		desc: "when expecting a connection error",
		subj: URLCheck{
			Name: "connerr",
			Url:  "http://connection.failure/",
		},
		healthy: StatusUnknown,
		errors:  true,
	}, {
		desc: "when result is 200",
		subj: URLCheck{
			Name: "test200",
			Url:  urls["200"],
		},
		healthy: StatusHealthy,
		text:    "200",
	}, {
		desc: "when result fails on 200",
		subj: URLCheck{
			Name:     "fail200",
			Url:      urls["200"],
			FailTrue: true,
		},
		healthy: StatusUnhealthy,
		text:    "200",
	}, {
		desc: "when result is 404",
		subj: URLCheck{
			Name: "test404",
			Url:  urls["404"],
		},
		healthy: StatusUnhealthy,
		text:    "404",
	}, {
		desc: "when expected response if 404",
		subj: URLCheck{
			Name:   "expect404",
			Url:    urls["404"],
			RegExp: "404",
		},
		healthy: StatusHealthy,
		text:    "404",
	}, {
		desc: "when checking json with nil path",
		subj: URLCheck{
			Name:   "jsonnopath",
			Url:    urls["200json"],
			Type:   RespTypeJSON,
			RegExp: "healthy",
		},
		healthy: StatusUnknown,
		errors:  true,
	}, {
		desc: "when checking json with invalid path",
		subj: URLCheck{
			Name:     "jsoninvalidpath",
			Url:      urls["200json"],
			Type:     RespTypeJSON,
			JSONPath: "$.string",
		},
		healthy: StatusUnknown,
		errors:  true,
	}, {
		desc: "when checking json with invalid regexp",
		subj: URLCheck{
			Name:     "jsoninvalidre",
			Url:      urls["200json"],
			Type:     RespTypeJSON,
			JSONPath: "string",
			RegExp:   `)`,
		},
		healthy: StatusUnknown,
		text:    "healthy", // This is the json value, not our healthy result
		errors:  true,
	}, {
		desc: "when checking json with invalid json in response body",
		subj: URLCheck{
			Name:     "jsoninvalidresp",
			Url:      urls["200invalidjson"],
			Type:     RespTypeJSON,
			JSONPath: "string",
			RegExp:   "healthy",
		},
		healthy: StatusUnknown,
		errors:  true,
	}, {
		desc: "when checking json with valid path but nil regexp",
		subj: URLCheck{
			Name:     "jsonnore",
			Url:      urls["200json"],
			Type:     RespTypeJSON,
			JSONPath: "string",
		},
		healthy: StatusHealthy,
		text:    "healthy",
	}, {
		desc: "when checking json string value",
		subj: URLCheck{
			Name:     "jsonstr",
			Url:      urls["200json"],
			Type:     RespTypeJSON,
			JSONPath: "string",
			RegExp:   "healthy",
		},
		healthy: StatusHealthy,
		text:    "healthy",
	}, {
		desc: "when checking json true bool value",
		subj: URLCheck{
			Name:     "jsonbool",
			Url:      urls["200json"],
			Type:     RespTypeJSON,
			JSONPath: "booltrue",
			RegExp:   "true",
		},
		healthy: StatusHealthy,
		text:    "true",
	}, {
		desc: "when checking json bool value for !true",
		subj: URLCheck{
			Name:     "jsonbool",
			Url:      urls["200json"],
			Type:     RespTypeJSON,
			JSONPath: "boolfalse",
			RegExp:   "true",
			FailTrue: true,
		},
		healthy: StatusHealthy,
		text:    "false",
	}, {
		desc: "when checking json int value",
		subj: URLCheck{
			Name:     "jsonint",
			Url:      urls["200json"],
			Type:     RespTypeJSON,
			JSONPath: "integer",
			RegExp:   "1269",
		},
		healthy: StatusHealthy,
		text:    "1269",
	}, {
		desc: "when checking nested json value",
		subj: URLCheck{
			Name:     "nestedjson",
			Url:      urls["200json"],
			Type:     RespTypeJSON,
			JSONPath: "nested.'light.color'",
			RegExp:   "green",
		},
		healthy: StatusHealthy,
		text:    "green",
	}, {
		desc: "when checking json map value",
		subj: URLCheck{
			Name:     "jsonmap",
			Url:      urls["200json"],
			Type:     RespTypeJSON,
			JSONPath: "nested",
		},
		healthy: StatusUnknown,
		errors:  true,
	}, {
		desc: "when checking http body and nil regexp",
		subj: URLCheck{
			Name: "bodynore",
			Url:  urls["200"],
			Type: RespTypeHTTPBody,
		},
		healthy: StatusUnknown,
		errors:  true,
	}, {
		desc: "when checking http body expecting a match",
		subj: URLCheck{
			Name:   "bodymatch",
			Url:    urls["200"],
			Type:   RespTypeHTTPBody,
			RegExp: "ma[^r]ch",
		},
		healthy: StatusHealthy,
		text:    "healthy",
	}, {
		desc: "when checking http body expecting no match",
		subj: URLCheck{
			Name:   "bodyunmatched",
			Url:    urls["200"],
			Type:   RespTypeHTTPBody,
			RegExp: "march",
		},
		healthy: StatusUnhealthy,
		text:    "unhealthy",
	}}

	for _, testcase := range testCases {
		testcase.subj.Check()

		require.NotNil(t, testcase.subj.Report, "Report can not be nil.")
		if testcase.errors {
			assert.NotEmpty(t, testcase.subj.Report.Errors, fmt.Sprintf("No errors occurred %s", testcase.desc))
		} else {
			assert.Empty(t, testcase.subj.Report.Errors, fmt.Sprintf("Errors occurred %s", testcase.desc))
		}
		assert.Equal(t, testcase.healthy, testcase.subj.Report.Healthy, fmt.Sprintf("Healthy assertion fails %s", testcase.desc))
		assert.Equal(t, testcase.text, testcase.subj.Report.Text, fmt.Sprintf("Text assertion fails %s", testcase.desc))
	}

	// Finally we do a one-off test if RestClient isn't initialized for some reason.
	check := URLCheck{}
	RestyClient = nil
	report := check.Check()
	assert.NotEmpty(t, report.Errors, "We should error if RestyClient is nil.")
}
