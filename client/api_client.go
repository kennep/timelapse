package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kennep/timelapse/api"
	"github.com/sirupsen/logrus"
)

const applicationJson = "application/json"

type ApiClient struct {
	credentials   *Credentials
	configuration *Configuration
	client        *http.Client
}

type HTTPResponseError struct {
	response *http.Response
	body     []byte
}

type RemoteError struct {
	Message string `json:"message"`
}

func (err *HTTPResponseError) Error() string {
	return fmt.Sprintf("%d: %s", err.response.StatusCode, err.body)
}

func (err *RemoteError) Error() string {
	return err.Message
}

func NewApiClient() (*ApiClient, error) {
	var apiClient ApiClient

	credentials, err := GetCredentials()
	if err != nil {
		return nil, err
	}

	configuration, err := GetConfiguration()
	if err != nil {
		return nil, err
	}

	apiClient.credentials = credentials
	apiClient.configuration = configuration
	apiClient.client = &http.Client{}

	return &apiClient, nil
}

func (c *ApiClient) requestFor(method string, path string, body []byte) (*http.Request, error) {
	url := c.configuration.BaseURL
	if strings.HasSuffix(url, "/") {
		url += path[1:]
	} else {
		url += path
	}

	request, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	token, err := c.getAuthorizationToken()
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+token)

	if logrus.GetLevel() >= logrus.TraceLevel {
		logrus.WithFields(logrus.Fields{
			"baseurl": c.configuration.BaseURL,
			"path":    path,
			"method":  method,
		}).Trace("HTTP Request")
		h := make(logrus.Fields)
		for k, v := range request.Header {
			h[k] = v
		}
		logrus.WithFields(h).Trace("HTTP Request headers")
		logrus.WithFields(logrus.Fields{
			"body": string(body),
		}).Trace("HTTP Request body")
	}

	return request, nil
}

func (c *ApiClient) getAuthorizationToken() (string, error) {
	creds := c.credentials.Credentials[c.credentials.DefaultProvider]
	if creds == nil {
		return "", errors.New("No credentials found. Please login first.")
	}
	token := creds.IDToken
	if token == "" {
		return "", errors.New("No token found. Please login first.")
	}
	return token, nil
}

func (c *ApiClient) getRefreshToken() (string, error) {
	creds := c.credentials.Credentials[c.credentials.DefaultProvider]
	if creds == nil {
		return "", errors.New("No credentials found. Please login first.")
	}
	token := creds.RefreshToken
	if token == "" {
		return "", errors.New("No token found. Please login first.")
	}
	return token, nil
}

func (c *ApiClient) jsonRequest(method string, path string, request interface{}, response interface{}) error {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return err
	}

	responseBody, err := c.doRequest(method, path, requestBody)
	if err != nil {
		return err
	}

	err = json.Unmarshal(responseBody, response)
	if err != nil {
		return err
	}

	return nil
}

func (c *ApiClient) doRequest(method string, path string, body []byte) ([]byte, error) {
	req, err := c.requestFor(method, path, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		refreshToken, err := c.getRefreshToken()
		err = refreshTokens(refreshToken)
		if err != nil {
			return nil, err
		}

		c.credentials, err = GetCredentials()
		if err != nil {
			return nil, err
		}

		req, err := c.requestFor(method, path, body)
		if err != nil {
			return nil, err
		}

		resp, err = c.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	responseBody := buf.Bytes()

	if logrus.GetLevel() >= logrus.TraceLevel {
		h := make(logrus.Fields)
		for k, v := range resp.Header {
			h[k] = v
		}
		logrus.WithFields(h).Trace("HTTP Response headers")
		logrus.WithFields(logrus.Fields{
			"body": string(responseBody),
		}).Trace("HTTP Response body")
	}

	if resp.StatusCode >= 300 {
		var remoteError RemoteError
		err = json.Unmarshal(responseBody, &remoteError)
		if err == nil && remoteError.Message != "" {
			return nil, &remoteError
		}

		return nil, &HTTPResponseError{resp, responseBody}
	}

	return responseBody, err
}

func (c *ApiClient) CreateProject(project *api.Project) (*api.Project, error) {
	var createdProject api.Project
	err := c.jsonRequest("POST", "/projects", project, &createdProject)
	if err != nil {
		return nil, err
	}
	return &createdProject, nil
}

func (c *ApiClient) GetProject(projectName string) (*api.Project, error) {
	var project api.Project
	err := c.jsonRequest("GET", fmt.Sprintf("/projects/%s", url.QueryEscape(projectName)), nil, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (c *ApiClient) UpdateProject(projectName string, project *api.Project) (*api.Project, error) {
	var updatedProject api.Project
	err := c.jsonRequest("PUT", fmt.Sprintf("/projects/%s", url.QueryEscape(projectName)), project, &updatedProject)
	if err != nil {
		return nil, err
	}

	return &updatedProject, nil
}

func (c *ApiClient) ListProjects() ([]*api.Project, error) {
	var projects []*api.Project
	err := c.jsonRequest("GET", "/projects", nil, &projects)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (c *ApiClient) GetTimeEntries(projectName string) ([]*api.TimeEntry, error) {
	var timeentries []*api.TimeEntry
	var path string
	if projectName == "" {
		path = "/entries"
	} else {
		path = fmt.Sprintf("/projects/%s/entries", url.QueryEscape(projectName))
	}
	err := c.jsonRequest("GET", path, nil, &timeentries)
	if err != nil {
		return nil, err
	}
	return timeentries, nil
}

func (c *ApiClient) AddTimeEntry(projectName string, entry *api.TimeEntry) (*api.TimeEntry, error) {
	path := fmt.Sprintf("/projects/%s/entries", url.QueryEscape(projectName))

	var entryResult api.TimeEntry
	err := c.jsonRequest("POST", path, entry, &entryResult)
	if err != nil {
		return nil, err
	}

	return &entryResult, nil
}

func (c *ApiClient) GetTimeEntry(projectName string, entryID string) (*api.TimeEntry, error) {
	path := fmt.Sprintf("/projects/%s/entries/%s", url.QueryEscape(projectName), url.QueryEscape(entryID))

	var entryResult api.TimeEntry
	err := c.jsonRequest("GET", path, nil, &entryResult)
	if err != nil {
		return nil, err
	}

	return &entryResult, nil
}

func (c *ApiClient) UpdateTimeEntry(projectName string, entry *api.TimeEntry) (*api.TimeEntry, error) {
	path := fmt.Sprintf("/projects/%s/entries/%s", url.QueryEscape(projectName), url.QueryEscape(entry.ID))

	var entryResult api.TimeEntry
	err := c.jsonRequest("PUT", path, entry, &entryResult)
	if err != nil {
		return nil, err
	}

	return &entryResult, nil
}
