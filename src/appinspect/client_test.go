package appinspect

import (
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

const TESTING_URL = "http://foo-bar"

func getClient() *Client {
	client := New()
	client.Client.SetHostURL(TESTING_URL)
	httpmock.ActivateNonDefault(client.GetClient())
	return client
}

func TestClientImplementsInterface(t *testing.T) {
	var _ ClientInterface = (*Client)(nil)
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

	responder, _ = httpmock.NewJsonResponder(401, nil)
	httpmock.RegisterResponder("GET", TESTING_URL+"/validate/status/"+requestId, responder)
	status, err := client.Status(requestId)
	assert.Error(err)
	assert.Equal(status.RequestID, requestId)
	assert.Equal(status.StatusCode, 401)
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

	responder, _ = httpmock.NewJsonResponder(402, nil)
	httpmock.RegisterResponder("GET", TESTING_URL+"/validate/status/"+requestId, responder)
	status, err := client.Status(requestId)
	assert.Error(err)
	assert.Equal(status.RequestID, requestId)
	assert.Equal(status.StatusCode, 402)
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

	responder, _ = httpmock.NewJsonResponder(401, nil)
	httpmock.RegisterResponder("GET", TESTING_URL+"/report/"+requestId, responder)
	report, err := client.ReportJSON(requestId)
	assert.Error(err)
	assert.Nil(report)
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

	httpmock.RegisterResponder("GET", TESTING_URL+"/report/"+requestId, httpmock.NewBytesResponder(
		402, nil))
	report, err := client.ReportHTML(requestId)
	assert.Error(err)
	assert.Nil(report)
}

func TestAddIdToRequest(t *testing.T) {
	assert := assert.New(t)
	request := getClient().Client.R()

	request.URL = "/foo/"
	result, err := addIdToRequest(request, "bar")
	assert.Nil(err)
	assert.Empty(result.Sha)
	assert.Equal("bar", result.RequestID)
	assert.Equal("/foo/bar", request.URL)

	request.URL = "/foo/"
	result, err = addIdToRequest(request, ShaId{"baz", []string{"test-tag"}})
	assert.Nil(err)
	assert.Empty(result.RequestID)
	assert.Equal("baz", result.Sha)
	assert.Equal("/foo/baz", request.URL)

	_, err = addIdToRequest(request, 1)
	assert.Error(err)
}
