package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/martian/log"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"time"
)

const (
	RequestId             = "request_id"
	headerContentType     = "Content-Type"
	headerApplicationJSON = "application/json"
	headerAuthorization   = "Authorization"
	headerBearer          = "Bearer "
)

type Client interface {
	SendMessage(ctx context.Context, reqMsg *RequestMessage) (res *ResponseMessage, err error)
}

type client struct {
	opt *Option
}

// New initialize a new client for Wappin.
func New(opts ...FnOption) (c Client) {
	cl := new(client)
	o := (new(Option)).Assign(opts...).
		setWappinClient(cl).
		Default()
	c = cl.Assign(o)
	return
}

func (c *client) Assign(o *Option) *client {
	if o == nil {
		return c
	}

	c.opt = o.Clone()
	return c
}

// SendMessage for sending Whatsapp message to Wappin, currently we only support for message type with template, we can add components are image, video and text in the message.
func (c *client) SendMessage(ctx context.Context, reqMsg *RequestMessage) (res *ResponseMessage, err error) {
	if reqMsg == nil {
		err = errors.New("Request nil arguments")
		return
	}

	res, err = c.postToWappin(ctx, c.opt.MessagesURL, reqMsg)
	if err != nil {
		return
	}

	return res, err
}

func (c *client) postToWappin(ctx context.Context, endpoint string, body interface{}) (res *ResponseMessage, err error) {
	requestId := c.getRequestId(ctx)

	var buff bytes.Buffer
	err = json.NewEncoder(&buff).Encode(body)
	if err != nil {
		log.Errorf("Error encoding request body with request_id = %s and payload = %s and error = %v", requestId, buff, err)
		return
	}

	// getting token
	token, err := c.getToken(ctx)
	if err != nil {
		log.Errorf("Error get token with request_id = %s and payload = %s and error = %v", requestId, buff, err)
		return
	}

	// prepare the request
	url := c.opt.BaseURL + endpoint
	req, err := http.NewRequest(http.MethodPost, url, &buff)
	if err != nil {
		log.Errorf("Error create HTTP request with request_id = %s and payload = %s and error = %v", requestId, buff, err)
		return
	}

	// set token to header and do request to Wapppin
	req.Header.Set(headerAuthorization, headerBearer+token)
	resp, err := c.opt.client.Do(c.prepareRequest(ctx, req))
	if err != nil {
		log.Errorf("Error HTTP request with request_id = %s and payload = %s and error = %v", requestId, buff, err)
		return
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	byteBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error get body response with request_id = %s and payload = %s and error = %v", requestId, buff, err)
		return
	}

	err = json.Unmarshal(byteBody, &res)
	if err != nil {
		log.Errorf("Error unmarshalling response with request_id = %s and payload = %s and error = %v", requestId, buff, err)
		return
	}

	// casting error from wappin if any error response
	err = getError(resp.StatusCode, res.Errors)
	if err != nil {
		return nil, err
	}

	return
}

func (c *client) getToken(ctx context.Context) (token string, err error) {
	// looking for token from cache
	tokenInterface, err := c.opt.Storage.Get(ctx, c.opt.TokenCacheKey)
	if err != nil && err != redis.Nil {
		return "", err
	}

	// convert token if not empty
	if tokenInterface != nil {
		tokenConv := fmt.Sprintf("%v", tokenInterface)
		return tokenConv, nil
	}

	url := c.opt.BaseURL + c.opt.LoginURL
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return
	}

	req.SetBasicAuth(c.opt.Username, c.opt.Password)
	resp, err := c.opt.client.Do(c.prepareRequest(ctx, req))
	if err != nil {
		return
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	var responseLogin ResponseLogin
	err = json.NewDecoder(resp.Body).Decode(&responseLogin)
	if err != nil {
		return
	}

	if len(responseLogin.Users) > 0 {
		token = responseLogin.Users[0].Token
		expiredStr := responseLogin.Users[0].ExpiredAfter

		ttlToken, err := c.getTTLToken(expiredStr)
		if err != nil {
			return "", err
		}

		err = c.opt.Storage.Save(ctx, c.opt.TokenCacheKey, token, ttlToken)
		if err != nil {
			return "", err
		}

		return token, err
	}

	err = getError(resp.StatusCode, responseLogin.Errors)
	return "", err
}

func (c *client) getTTLToken(expiredStr string) (time.Duration, error) {
	myTime, err := time.Parse("2006-01-02T15:04:05+07:00", expiredStr)
	if err != nil {
		panic(err)
	}

	// reducing the token for one and a half days
	now := time.Now().Add(time.Hour * 36)
	return myTime.Sub(now), err
}

func (c *client) prepareRequest(ctx context.Context, req *http.Request) *http.Request {
	req.Header.Set(headerContentType, headerApplicationJSON)
	return req.WithContext(ctx)
}

func (c *client) getRequestId(ctx context.Context) string {
	var reqID = ""
	if ctx == nil {
		return reqID
	}

	temp := ctx.Value(RequestId)
	if reqIDStr, ok := temp.(string); ok {
		reqID = reqIDStr
	}
	return reqID
}
