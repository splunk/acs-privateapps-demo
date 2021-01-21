package appinspect

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/splunk/acs-privateapps-demo/src/appinspect"
)

const (
	packageFile = "app-package.tar.gz"
)

func TestClient(t *testing.T) {

	u := os.Getenv("SPLUNK_COM_USERNAME")
	p := os.Getenv("SPLUNK_COM_PASSWORD")

	if u == "" || p == "" {
		t.Skipf("splunk.com creds not provided, skipping test...")
	}

	assert := assert.New(t)
	c := appinspect.New()
	assert.NotNil(c)
	assert.NoError(c.Login(u, p))

	file, err := ioutil.ReadFile(packageFile)
	assert.NoError(err)

	submitResp, err := c.Submit(packageFile, bytes.NewReader(file))
	assert.NoError(err)
	assert.NotNil(submitResp)
	t.Logf("submitted package for inspection, requestID=%s", submitResp.RequestID)

	var status *appinspect.StatusResult
	assert.Eventually(func() bool {
		status, err = c.Status(submitResp.RequestID)
		assert.NoError(err)
		if status.Status != "SUCCESS" {
			t.Logf("waiting for inspection to finish, current status: %s", status.Status)
			return false
		}
		return true
	}, 30*time.Minute, 2*time.Second)
	t.Logf("inspection complete, status: %v", *status)

	report, err := c.ReportJSON(submitResp.RequestID)
	assert.NoError(err)
	assert.NotNil(report)
	t.Logf("inspection report(summary): %v", report.Summary)
}
