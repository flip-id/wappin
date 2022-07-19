package wappin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/fairyhunter13/pool"
	"github.com/flip-id/valuefirst/manager"
	"github.com/gofiber/fiber/v2"
)

var _ manager.TokenClient = new(client)

type Client interface {
	// Implement all TokenClient interface from the valuefirst package.
	manager.TokenClient
	SendMessage(ctx context.Context, reqMsg *RequestWhatsappMessage) (res *ResponseMessage, err error)
	// 	TODO: Add method in here.
}

type client struct {
	opt *Option
}

// New initialize a new client for Wappin.
func New(opts ...FnOption) (c Client) {
	o := (new(Option)).Assign(opts...).Default()

	c = &client{o.Clone()}
	return
}

func (c *client) SendMessage(ctx context.Context, reqMsg *RequestWhatsappMessage) (res *ResponseMessage, err error) {
	if reqMsg == nil {
		err = ErrNilArguments
		return
	}

	res, err = c.postToWappin(ctx, EndpointSendHSM, reqMsg)
	return
}

func (c *client) postToWappin(ctx context.Context, endpoint string, body interface{}) (res *ResponseMessage, err error) {
	buff := pool.GetBuffer()
	defer pool.Put(buff)

	err = json.NewEncoder(buff).Encode(body)
	if err != nil {
		return
	}

	url := c.opt.BaseURL + endpoint
	req, err := http.NewRequest(http.MethodPost, url, buff)
	if err != nil {
		return
	}

	// TODO: Add get token here.
	// req.Header.Set(fiber.HeaderAuthorization, TokenBearer+c.opt.)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req = req.WithContext(ctx)
	resp, err := c.opt.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	res.HttpStatusCode = resp.StatusCode
	byteBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	res.RawData = convertByteToString(byteBody)
	err = getError(res.HttpStatusCode, res.Status, res.Message)
	if err != nil {
		return
	}

	err = json.Unmarshal(byteBody, &res)
	return
}

// GenerateToken generates a token for Wappin.
func (c *client) GenerateToken(ctx context.Context) (res manager.ResponseGenerateToken, err error) {
	url := c.opt.BaseURL + EndpointToken
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return
	}

	req = req.WithContext(ctx)
	req.SetBasicAuth(c.opt.ClientID, c.opt.SecretKey)
	resp, err := c.opt.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	var accessToken AccessToken
	err = json.NewDecoder(resp.Body).Decode(&accessToken)
	if err != nil {
		return
	}

	res, err = accessToken.ToResponseGenerateToken(resp.StatusCode)
	return
}

// EnableToken enables a token for Wappin.
func (c *client) EnableToken(ctx context.Context, token string) (res manager.ResponseEnableToken, err error) {
	return
}

// DisableToken disables a token for Wappin.
func (c *client) DisableToken(ctx context.Context, token string) (res manager.ResponseEnableToken, err error) {
	return
}

// DeleteToken deletes the token.
func (c *client) DeleteToken(ctx context.Context, token string) (res manager.ResponseEnableToken, err error) {
	return
}
