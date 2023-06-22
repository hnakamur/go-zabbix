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
	"sync"
	"sync/atomic"
)

const contentType = "application/json-rpc"
const jsonrpcVersion = "2.0"
const jsonrpcFilename = "api_jsonrpc.php"
const loginMethod = "user.login"

// Client represents a client for Zabbix JSON-RPC API.
type Client struct {
	httpClient *http.Client
	apiURL     string
	host       string

	requestID atomic.Uint64
	sessionID string

	apiVerOnce sync.Once
	apiVer     APIVersion
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

// NewClient creates a client for Zabbix JSON-RPC API.
// zabbixURL is something like http://example.com/zabbix/, and not like
// http://example.com/zabbix/index.php.
func NewClient(zabbixURL string, opts ...ClientOpt) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}

	u, err := url.Parse(zabbixURL)
	if err != nil {
		return nil, err
	}
	c.apiURL = u.JoinPath(jsonrpcFilename).String()
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	return c, nil
}

// Login sends a "user.login" request to the server.
// If the login is successful, the session ID will be returned from the server.
// It is kept in the Client and it will be set to requests created with Call
// method called after this call of Login method.
func (c *Client) Login(ctx context.Context, username, password string) error {
	apiVer, err := c.APIVersion(ctx)
	if err != nil {
		return err
	}

	var params any
	if apiVer.Compare(APIVersion{Major: 6, Minor: 4, Patch: 0,
		PreRelType: Beta, PreRelVer: 5}) >= 0 {
		// https://support.zabbix.com/browse/ZBXNEXT-8085
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
	if err := c.Call(ctx, loginMethod, params, &auth); err != nil {
		return err
	}
	if auth == "" {
		return errors.New("user.login API should have return a valid (non-empty) auth")
	}

	c.sessionID = auth
	return nil
}

// Call sends a JSON-RPC request to the server.
// The caller of this method must pass a pointer to the appropriate type of result.
// The appropriate type is different for method and params.
func (c *Client) Call(ctx context.Context, method string, params, result any) error {
	type responseCommon struct {
		Jsonrpc string    `json:"jsonrpc"`
		Error   *APIError `json:"error"`
		ID      uint64    `json:"id"`
	}

	var res struct {
		responseCommon
		Result any `json:"result"`
	}
	res.Result = result
	req, err := c.internalCall(ctx, method, params, &res)
	if err != nil {
		return &CallError{
			ID:     req.ID,
			Method: req.Method,
			Params: req.Params,
			Err:    err,
		}
	}
	if res.Error != nil {
		return &CallError{
			ID:     req.ID,
			Method: req.Method,
			Params: req.Params,
			Err:    res.Error,
		}
	}
	if res.ID != req.ID {
		return &CallError{
			ID:     req.ID,
			Method: req.Method,
			Params: req.Params,
			Err: fmt.Errorf("response ID (%d) does not match resquest ID (%d)",
				res.ID, req.ID),
		}
	}
	return nil
}

// CallError is the error type returned by Client.Call method.
// The concret type of the Err field is APIError or other error.
type CallError struct {
	ID     uint64 `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
	Err    error  `json:"error"`
}

var _ error = (*CallError)(nil)

func (e *CallError) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (e *CallError) Unwrap() error {
	return e.Err
}

type ErrorCode int

const (
	// ErrorCodeNone represents the error which is not an error object from
	// Zabbix JSON-RPC API but some other error occurred in the client side.
	ErrorCodeNone ErrorCode = 0

	ErrorCodeParse          ErrorCode = -32700
	ErrorCodeInvalidRequest ErrorCode = -32600
	ErrorCodeMethodNotFound ErrorCode = -32601
	ErrorCodeInvalidParams  ErrorCode = -32602
	ErrorCodeInternal       ErrorCode = -32603
	ErrorCodeApplication    ErrorCode = -32500
	ErrorCodeSystem         ErrorCode = -32400
	ErrorCodeTransport      ErrorCode = -32300
)

// GetErrorCode returns the Code field if err or a unwrapped error is APIError
// or ErrorCodeNone otherwise.
func GetErrorCode(err error) ErrorCode {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	return ErrorCodeNone
}

// APIError is an error object in responses from Zabbix JSON-RPC API.
// https://www.zabbix.com/documentation/current/en/manual/api#error-handling
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Data    string    `json:"data"`
}

var _ error = (*APIError)(nil)

func (e *APIError) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// APIVersion returns APIVersion.
// For the first call of this method, a request is sent to the server and
// the result will be cached.
// For subsequent call of this method, it returns the cached value.
func (c *Client) APIVersion(ctx context.Context) (APIVersion, error) {
	var err error
	c.apiVerOnce.Do(func() {
		c.apiVer, err = c.getAPIVersion(ctx)
	})
	if err != nil {
		return c.apiVer, err
	}
	return c.apiVer, nil
}

func (c *Client) getAPIVersion(ctx context.Context) (APIVersion, error) {
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
