package zabbix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync/atomic"
)

const contentType = "application/json-rpc"
const jsonrpcVersion = "2.0"
const jsonrpcEndpoint = "api_jsonrpc.php"
const loginMethod = "user.login"

// Client represents a client for Zabbix API. NewClient() to create a Client.
type Client struct {
	httpClient *http.Client
	apiURL     string
	host       string

	requestID  atomic.Uint64
	sessionID  string
	apiVersion APIVersion
}

type ClientOpt func(c *Client)

func WithHost(host string) ClientOpt {
	return func(c *Client) {
		c.host = host
	}
}

func WithHTTPClient(httpClient *http.Client) ClientOpt {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func NewClient(zabbixURL string, opts ...ClientOpt) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}

	u, err := url.Parse(zabbixURL)
	if err != nil {
		return nil, err
	}
	c.apiURL = u.JoinPath(jsonrpcEndpoint).String()
	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}
	return c, nil
}

// Login to Zabbix API
func (c *Client) Login(ctx context.Context, username, password string) error {
	if err := c.getAPIInfoVersionOnce(ctx); err != nil {
		return err
	}

	var params any
	if c.apiVersion.Compare(APIVersion{Major: 5, Minor: 4}) >= 0 {
		params = &struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: username, Password: password,
		}
	} else {
		params = &struct {
			User     string `json:"user"`
			Password string `json:"password"`
		}{
			User: username, Password: password,
		}
	}

	var auth string
	err := c.Call(ctx, loginMethod, params, &auth)
	if err != nil {
		return err
	}

	if auth == "" {
		// NOTE: When a error happens, rpcResponse.Error becomes non-null,
		// so this should not happen.
		return errors.New("user.login API should have return a valid (non-empty) auth")
	}

	c.sessionID = auth

	return nil
}

// Call calls a Zabbix API and gets the result.
// See https://github.com/hnakamur/go-zabbix/blob/46d9f81a6406cecd04ff2f9d41b29efb475a58e9/cmd/example/main.go#L113-L142
// for an example.
func (c *Client) Call(ctx context.Context, method string, params, result any) error {
	type responseCommon struct {
		Jsonrpc string `json:"jsonrpc"`
		Error   *Error `json:"error"`
		ID      uint64 `json:"id"`
	}

	var res struct {
		responseCommon
		Result any `json:"result,string"`
	}
	res.Result = result
	req, err := c.internalCall(ctx, method, params, &res)
	if err != nil {
		return err
	}
	if res.Error != nil {
		res.Error.Method = method
		return res.Error
	}
	if res.ID != req.ID {
		return fmt.Errorf("response ID (%d) does not match resquest ID (%d)", res.ID, req.ID)
	}
	return nil
}

// Error represents an error from Zabbix API
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
	Method  string `json:"-"`
}

// Returns a string for an error from Zabbix API
func (e Error) Error() string {
	return fmt.Sprintf("%s method=%s, code=%d, data=%s", e.Message, e.Method, e.Code, e.Data)
}

func (c *Client) getAPIInfoVersionOnce(ctx context.Context) error {
	if !c.apiVersion.IsZero() {
		return nil
	}
	ver, err := c.apiInfoVersion(ctx)
	if err != nil {
		return err
	}
	c.apiVersion = ver
	return nil
}

func (c *Client) apiInfoVersion(ctx context.Context) (APIVersion, error) {
	var v APIVersion
	var ver string
	if err := c.Call(ctx, "apiinfo.version", []string{}, &ver); err != nil {
		return v, err
	}
	v, err := ParseAPIVersion(ver)
	if err != nil {
		return v, err
	}
	return v, nil
}

func (c *Client) APIVersion() APIVersion {
	return c.apiVersion
}

func (c *Client) internalCall(ctx context.Context, method string, params, result any) (req *rpcRequest, err error) {
	req = c.newRPCRequest(method, params)
	httpReq, err := c.newHTTPRequestWithContext(ctx, req)
	if err != nil {
		return req, err
	}
	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return req, err
	}
	defer httpRes.Body.Close()

	bodyBytes, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return req, err
	}
	// log.Printf("response: %s", string(bodyBytes))
	if err := json.Unmarshal(bodyBytes, result); err != nil {
		return req, err
	}
	return req, nil
}

type rpcRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      uint64 `json:"id"`
	Auth    any    `json:"auth"`
}

func (c *Client) newRPCRequest(method string, params any) *rpcRequest {
	reqID := c.requestID.Add(1)

	r := &rpcRequest{
		Jsonrpc: jsonrpcVersion,
		Method:  method,
		Params:  params,
		ID:      reqID,
	}
	if c.sessionID != "" && method != loginMethod {
		r.Auth = c.sessionID
	}
	return r
}

func (c *Client) newHTTPRequestWithContext(ctx context.Context, r *rpcRequest) (*http.Request, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	// log.Printf("request:%s", string(b))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if c.host != "" {
		req.Host = c.host
	}
	req.Header.Set("Content-Type", contentType)
	return req, nil
}
