package appinspect

import (
	"fmt"
	"io"
	"net/url"

	"github.com/go-resty/resty/v2"
)

const (
	appInspectBaseURL = "https://appinspect.splunk.com/v1/app"
)

// Client to interface with the appinspect service
type Client struct {
	*resty.Client
	token string
}

// Error ...
type Error struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Description)
}

// New client to interface with the appinspect service
func New() *Client {
	client := &Client{
		Client: resty.New().SetHostURL(appInspectBaseURL).SetError(&Error{}).SetAuthScheme("Bearer"),
	}
	client.Client = client.Client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		if client.token != "" {
			req.SetAuthToken(client.token)
		}
		return nil
	})
	return client
}

// NewWithToken ...
func NewWithToken(token string) *Client {
	c := New()
	c.token = token
	return c
}

// AuthenticateResult ...
type AuthenticateResult struct {
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"`
	Msg        string `json:"msg"`
	Data       struct {
		Token string `json:"token"`
		User  struct {
			Name     string   `json:"name"`
			Email    string   `json:"email"`
			Username string   `json:"username"`
			Groups   []string `json:"groups"`
		} `json:"user"`
	} `json:"data"`
}

// Authenticate ...
func Authenticate(username, password string) (*AuthenticateResult, error) {
	type erro struct {
		StatusCode int    `json:"status_code"`
		Status     string `json:"status"`
		Msg        string `json:"msg"`
		Errors     string `json:"errors"`
	}
	resp, err := resty.New().R().SetBasicAuth(username, password).SetResult(&AuthenticateResult{}).SetError(&erro{}).
		Get("https://api.splunk.com/2.0/rest/login/splunk")
	if err != nil {
		return nil, fmt.Errorf("error while login: %s", err)
	}
	if resp.IsError() {
		if e, ok := resp.Error().(*erro); ok {
			return nil, fmt.Errorf("error while login: %s-%s", e.Status, e.Msg)
		}
		return nil, fmt.Errorf("error while login: %s", resp.Status())
	}
	var r *AuthenticateResult
	var ok bool
	if r, ok = resp.Result().(*AuthenticateResult); !ok {
		return nil, fmt.Errorf("error while login: failed to parse response")
	}
	return r, nil
}

// Login to appinspect service
func (c *Client) Login(username, password string) error {
	r, err := Authenticate(username, password)
	if err != nil {
		return err
	}
	c.token = r.Data.Token
	return nil
}

// SubmitResult ...
type SubmitResult struct {
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
	Links     []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

// Submit an app-package for inspection
func (c *Client) Submit(filename string, file io.Reader) (*SubmitResult, error) {

	formdata := url.Values{
		"included_tags": []string{"cloud", "self-service", "private_app"},
	}
	resp, err := c.R().SetAuthToken(c.token).SetFormDataFromValues(formdata).
		SetFileReader("app_package", filename, file).SetResult(&SubmitResult{}).Post("/validate")
	if err != nil {
		return nil, fmt.Errorf("error while submit: %s", err)
	}
	if resp.IsError() {
		if e, ok := resp.Error().(*Error); ok {
			return nil, fmt.Errorf("error while submit: %s", e)
		}
		return nil, fmt.Errorf("error while submit: %s", resp.Status())
	}
	var r *SubmitResult
	var ok bool
	if r, ok = resp.Result().(*SubmitResult); !ok {
		return nil, fmt.Errorf("error while submit: failed to parse response")
	}
	return r, nil
}

// StatusResult ...
type StatusResult struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Info      struct {
		Error         int `json:"error"`
		Failure       int `json:"failure"`
		Skipped       int `json:"skipped"`
		ManualCheck   int `json:"manual_check"`
		NotApplicable int `json:"not_applicable"`
		Warning       int `json:"warning"`
		Success       int `json:"success"`
	} `json:"info"`
	Links []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

// Status of an app-package inspection
func (c *Client) Status(requestID string) (*StatusResult, error) {
	resp, err := c.R().SetAuthToken(c.token).SetResult(&StatusResult{}).Get("/validate/status/" + requestID)
	if err != nil {
		return nil, fmt.Errorf("error while getting status: %s", err)
	}
	if resp.IsError() {
		if e, ok := resp.Error().(*Error); ok {
			return nil, fmt.Errorf("error while getting status: %s", e)
		}
		return nil, fmt.Errorf("error while getting status: %s", resp.Status())
	}
	var r *StatusResult
	var ok bool
	if r, ok = resp.Result().(*StatusResult); !ok {
		return nil, fmt.Errorf("error while getting status: failed to parse response")
	}
	return r, nil
}

// ReportJSONResult ...
type ReportJSONResult struct {
	RequestID string `json:"request_id"`
	Cloc      string `json:"cloc"`
	Reports   []struct {
		AppAuthor      string `json:"app_author"`
		AppDescription string `json:"app_description"`
		AppHash        string `json:"app_hash"`
		AppName        string `json:"app_name"`
		AppVersion     string `json:"app_version"`
		Metrics        struct {
			StartTime     string  `json:"start_time"`
			EndTime       string  `json:"end_time"`
			ExecutionTime float64 `json:"execution_time"`
		} `json:"metrics"`
		RunParameters struct {
			APIRequestID      string   `json:"api_request_id"`
			Identity          string   `json:"identity"`
			SplunkbaseID      string   `json:"splunkbase_id"`
			Version           string   `json:"version"`
			SplunkVersion     string   `json:"splunk_version"`
			StackID           string   `json:"stack_id"`
			APITimestamp      string   `json:"api_timestamp"`
			PackageLocation   string   `json:"package_location"`
			AppinspectVersion string   `json:"appinspect_version"`
			IncludedTags      []string `json:"included_tags"`
			ExcludedTags      []string `json:"excluded_tags"`
		} `json:"run_parameters"`
		Groups []struct {
			Checks []struct {
				Description string `json:"description"`
				Messages    []struct {
					Code            string      `json:"code"`
					Filename        string      `json:"filename"`
					Line            int         `json:"line"`
					Message         string      `json:"message"`
					Result          string      `json:"result"`
					MessageFilename string      `json:"message_filename"`
					MessageLine     interface{} `json:"message_line"`
				} `json:"messages"`
				Name   string   `json:"name"`
				Tags   []string `json:"tags"`
				Result string   `json:"result"`
			} `json:"checks"`
			Description string `json:"description"`
			Name        string `json:"name"`
		} `json:"groups"`
		Summary struct {
			Error         int `json:"error"`
			Failure       int `json:"failure"`
			Skipped       int `json:"skipped"`
			ManualCheck   int `json:"manual_check"`
			NotApplicable int `json:"not_applicable"`
			Warning       int `json:"warning"`
			Success       int `json:"success"`
		} `json:"summary"`
	} `json:"reports"`
	Summary struct {
		Error         int `json:"error"`
		Failure       int `json:"failure"`
		Skipped       int `json:"skipped"`
		ManualCheck   int `json:"manual_check"`
		NotApplicable int `json:"not_applicable"`
		Warning       int `json:"warning"`
		Success       int `json:"success"`
	} `json:"summary"`
	Metrics struct {
		StartTime     string  `json:"start_time"`
		EndTime       string  `json:"end_time"`
		ExecutionTime float64 `json:"execution_time"`
	} `json:"metrics"`
	RunParameters struct {
		APIRequestID      string   `json:"api_request_id"`
		Identity          string   `json:"identity"`
		SplunkbaseID      string   `json:"splunkbase_id"`
		Version           string   `json:"version"`
		SplunkVersion     string   `json:"splunk_version"`
		StackID           string   `json:"stack_id"`
		APITimestamp      string   `json:"api_timestamp"`
		PackageLocation   string   `json:"package_location"`
		AppinspectVersion string   `json:"appinspect_version"`
		IncludedTags      []string `json:"included_tags"`
		ExcludedTags      []string `json:"excluded_tags"`
	} `json:"run_parameters"`
	Links []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

// ReportJSON of an app-package inspection
func (c *Client) ReportJSON(requestID string) (*ReportJSONResult, error) {
	resp, err := c.R().SetAuthToken(c.token).SetResult(&ReportJSONResult{}).Get("/report/" + requestID)
	if err != nil {
		return nil, fmt.Errorf("error while getting json report: %s", err)
	}
	if resp.IsError() {
		if e, ok := resp.Error().(*Error); ok {
			return nil, fmt.Errorf("error while getting json report: %s", e)
		}
		return nil, fmt.Errorf("error while getting json report: %s", resp.Status())
	}
	var r *ReportJSONResult
	var ok bool
	if r, ok = resp.Result().(*ReportJSONResult); !ok {
		return nil, fmt.Errorf("error while getting json report: failed to parse response")
	}
	return r, nil
}
