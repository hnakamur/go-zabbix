package zabbix

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

const contentType = "application/json-rpc"
const jsonrpcVersion = "2.0"
const jsonrpcEndpoint = "api_jsonrpc.php"
const loginMethod = "user.login"

type Client struct {
	client      http.Client
	host        string
	endpointURL string
	logger      Logger

	mu        sync.Mutex
	requestID uint64
	auth      string
}

func NewClient(zabbixURL, zabbixHost string, logger Logger) *Client {
	client := new(Client)
	if strings.HasSuffix(zabbixURL, "/") {
		client.endpointURL = zabbixURL + jsonrpcEndpoint
	} else {
		client.endpointURL = zabbixURL + "/" + jsonrpcEndpoint
	}
	client.host = zabbixHost
	client.logger = logger
	return client
}

type rpcRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      uint64      `json:"id"`
	Auth    interface{} `json:"auth"`
}

func (c *Client) newRPCRequest(method string, params interface{}) *rpcRequest {
	c.mu.Lock()
	c.requestID++
	c.mu.Unlock()

	r := &rpcRequest{
		Jsonrpc: jsonrpcVersion,
		Method:  method,
		Params:  params,
		ID:      c.requestID,
	}
	if c.auth != "" && method != loginMethod {
		r.Auth = c.auth
	}
	return r
}

func (c *Client) newHTTPRequest(r *rpcRequest) (*http.Request, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	if c.logger != nil {
		c.logger.Log("request:" + string(b))
	}
	req, err := http.NewRequest(http.MethodPost, c.endpointURL, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if c.host != "" {
		req.Host = c.host
	}
	req.Header.Set("Content-Type", contentType)
	return req, nil
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
	Method  string `json:"-"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s method=%s, code=%d, data=%s", e.Message, e.Method, e.Code, e.Data)
}

type rpcResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Error   *Error      `json:"error"`
	Result  interface{} `json:"result,string"`
	ID      uint64      `json:"id"`
}

func decodeResponse(r io.Reader, result interface{}) (*rpcResponse, error) {
	res := new(rpcResponse)
	res.Result = result
	err := json.NewDecoder(r).Decode(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type rpcCountResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Error   *Error `json:"error"`
	Result  int64  `json:"result,string"`
	ID      uint64 `json:"id"`
}

func decodeCountResponse(r io.Reader) (*rpcCountResponse, error) {
	res := new(rpcCountResponse)
	err := json.NewDecoder(r).Decode(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) Login(user, password string) error {
	params := struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}{
		User:     user,
		Password: password,
	}

	var auth string
	err := c.Call(loginMethod, params, &auth)
	if err != nil {
		return err
	}

	if auth == "" {
		// NOTE: When a error happens, rpcResponse.Error becomes non-null,
		// so this should not happen.
		return errors.New("user.login API should have return a valid (non-empty) auth.")
	}

	c.mu.Lock()
	c.auth = auth
	c.mu.Unlock()

	return nil
}

type responseCommon struct {
	Jsonrpc string `json:"jsonrpc"`
	Error   *Error `json:"error"`
	ID      uint64 `json:"id"`
}

func (c *Client) internalCall(method string, params interface{}, result interface{}) (req *rpcRequest, err error) {
	req = c.newRPCRequest(method, params)
	httpReq, err := c.newHTTPRequest(req)
	if err != nil {
		return
	}
	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return
	}
	defer httpRes.Body.Close()

	var buf *bytes.Buffer
	var reader io.Reader
	if c.logger != nil {
		buf = bytes.NewBuffer(make([]byte, 4096))
		reader = io.TeeReader(httpRes.Body, buf)
	} else {
		reader = httpRes.Body
	}
	err = json.NewDecoder(reader).Decode(result)
	if err != nil {
		return
	}
	if c.logger != nil {
		c.logger.Log("response:" + buf.String())
	}
	return
}

func (c *Client) Call(method string, params interface{}, result interface{}) error {
	var res struct {
		responseCommon
		Result interface{} `json:"result,string"`
	}
	res.Result = result
	req, err := c.internalCall(method, params, &res)
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

func (c *Client) CallForCount(method string, params interface{}) (int64, error) {
	var res struct {
		responseCommon
		Result int64 `json:"result,string"`
	}
	req, err := c.internalCall(method, params, &res)
	if err != nil {
		return 0, err
	}
	if res.Error != nil {
		res.Error.Method = method
		return 0, res.Error
	}
	if res.ID != req.ID {
		return 0, fmt.Errorf("response ID (%d) does not match resquest ID (%d)", res.ID, req.ID)
	}
	return res.Result, nil
}
