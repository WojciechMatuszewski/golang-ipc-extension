package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// RegisterResponse is the body of the response for /register
type RegisterResponse struct {
	FunctionName    string `json:"functionName"`
	FunctionVersion string `json:"functionVersion"`
	Handler         string `json:"handler"`
}

// NextEventResponse is the response for /event/next
type NextEventResponse struct {
	EventType          EventType `json:"eventType"`
	DeadlineMs         int64     `json:"deadlineMs"`
	RequestID          string    `json:"requestId"`
	InvokedFunctionArn string    `json:"invokedFunctionArn"`
}

// EventType represents the type of events received from /event/next
type EventType string

const (
	// Invoke is a lambda invoke
	Invoke EventType = "INVOKE"

	// Shutdown is a shutdown event for the environment
	Shutdown EventType = "SHUTDOWN"

	extensionNameHeader      = "Lambda-Extension-Name"
	extensionIdentiferHeader = "Lambda-Extension-Identifier"
	extensionErrorType       = "Lambda-Extension-Function-Error-Type"
)

type Client struct {
	baseURL     string
	extensionID string
	httpClient  *http.Client
}

func NewClient(runtimeURL string) *Client {
	return &Client{
		baseURL:    fmt.Sprintf("http://%s/2020-01-01/extension", runtimeURL),
		httpClient: &http.Client{},
	}
}

func (c *Client) Register(ctx context.Context) (*RegisterResponse, error) {
	const action = "/register"
	url := c.baseURL + "/register"

	reqBody, err := json.Marshal(map[string][]EventType{
		"events": {Invoke, Shutdown},
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set(extensionNameHeader, "main")

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if httpRes.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with status %s", httpRes.Status)
	}
	defer httpRes.Body.Close()
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}
	res := RegisterResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	c.extensionID = httpRes.Header.Get(extensionIdentiferHeader)
	print(c.extensionID)
	return &res, nil
}

// NextEvent blocks while long polling for the next lambda invoke or shutdown
func (c *Client) NextEvent(ctx context.Context) (*NextEventResponse, error) {
	const action = "/event/next"
	url := c.baseURL + action

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set(extensionIdentiferHeader, c.extensionID)
	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if httpRes.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with status %s", httpRes.Status)
	}
	defer httpRes.Body.Close()
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}
	res := NextEventResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
