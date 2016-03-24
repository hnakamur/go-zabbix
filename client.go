package zabbix

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const contentType = "application/json-rpc"
const jsonrpcVersion = "2.0"
const jsonrpcEndpoint = "api_jsonrpc.php"
const loginMethod = "user.login"

type Client struct {
	client      http.Client
	endpointURL string
	Logger      *log.Logger

	mu        sync.Mutex
	requestID uint64
	auth      string
}

func NewClient(zabbixURL string) *Client {
	client := new(Client)
	if strings.HasSuffix(zabbixURL, "/") {
		client.endpointURL = zabbixURL + jsonrpcEndpoint
	} else {
		client.endpointURL = zabbixURL + "/" + jsonrpcEndpoint
	}
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
	if c.Logger != nil {
		c.Logger.Printf("request:%s", string(b))
	}
	req, err := http.NewRequest(http.MethodPost, c.endpointURL, bytes.NewReader(b))
	if err != nil {
		return nil, err
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

func (c *Client) Call(method string, params interface{}, result interface{}) error {
	req := c.newRPCRequest(method, params)
	httpReq, err := c.newHTTPRequest(req)
	if err != nil {
		return err
	}
	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()

	var buf *bytes.Buffer
	var reader io.Reader
	if c.Logger != nil {
		buf = bytes.NewBuffer(make([]byte, 4096))
		reader = io.TeeReader(httpRes.Body, buf)
	} else {
		reader = httpRes.Body
	}
	res, err := decodeResponse(reader, result)
	if err != nil {
		return err
	}
	if c.Logger != nil {
		c.Logger.Printf("response:%s", buf.String())
	}
	if err := res.Error; err != nil {
		err.Method = method
		return err
	}
	if res.ID != req.ID {
		return fmt.Errorf("response ID (%d) does not match resquest ID (%d)", res.ID, req.ID)
	}
	return nil
}

func (c *Client) CallForCount(method string, params map[string]interface{}) (int64, error) {
	if params["countOutput"] != true {
		params = shallowCopyParams(params)
		params["countOutput"] = true
	}
	var value string
	err := c.Call(method, params, &value)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(value, 10, 64)
}

func shallowCopyParams(src map[string]interface{}) map[string]interface{} {
	dest := make(map[string]interface{})
	for k, v := range src {
		dest[k] = v
	}
	return dest
}
