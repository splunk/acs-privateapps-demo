package appinspect

import "testing"
import "github.com/jarcoal/httpmock"
import "github.com/stretchr/testify/assert"

const TESTING_URL = "http://foo-bar"

func getClient() *Client {
	client := New()
	client.Client.SetHostURL(TESTING_URL)
	httpmock.ActivateNonDefault(client.GetClient())
	return client
}

func TestStatusById(t *testing.T) {
	assert := assert.New(t)
	requestId := "Foo"
	client := getClient()

	responder, _ := httpmock.NewJsonResponder(200, StatusResult{
		RequestID: requestId,
	})
	httpmock.RegisterResponder("GET", TESTING_URL+"/validate/status/"+requestId, responder)

	status, _ := client.Status(requestId)
	assert.Equal(status.RequestID, requestId)
}

func TestStatusByHash(t *testing.T) {
	assert := assert.New(t)
	requestId := "Foo"
	client := getClient()

	responder, _ := httpmock.NewJsonResponder(200, StatusResult{
		RequestID: requestId,
	})
	httpmock.RegisterResponder("GET", TESTING_URL+"/validate/status/Baz?included_tags=test-tag", responder)

	status, _ := client.Status(ShaId{"Baz", []string{"test-tag"}})
	assert.Equal(status.RequestID, requestId)
}

func TestReportJSON(t *testing.T) {
	assert := assert.New(t)
	requestId := "Bar"
	client := getClient()

	responder, _ := httpmock.NewJsonResponder(200, ReportJSONResult{
		RequestID: requestId,
	})
	httpmock.RegisterResponder("GET", TESTING_URL+"/report/"+requestId, responder)

	report, _ := client.ReportJSON(requestId)
	assert.Equal(report.RequestID, requestId)
}

func TestReportHTML(t *testing.T) {
	assert := assert.New(t)
	requestId := "Baz"
	client := getClient()

	response := []byte("<i>Hello test</i>")

	httpmock.RegisterResponder("GET", TESTING_URL+"/report/"+requestId, httpmock.NewBytesResponder(
		200, response))

	report, _ := client.ReportHTML(requestId)
	assert.Equal(report, response)
}
