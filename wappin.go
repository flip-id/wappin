package wappin

import (
	"context"
	"encoding/json"
	"github.com/fairyhunter13/pool"
	"github.com/flip-id/valuefirst/manager"
	"github.com/flip-id/valuefirst/storage"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"time"
)

var _ manager.TokenClient = new(client)

type Client interface {
	// TokenClient implements all TokenClient interface from the valuefirst package.
	manager.TokenClient
	SendMessage(ctx context.Context, reqMsg *RequestWhatsappMessage) (res *ResponseMessage, err error)
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

func (c *client) SendMessage(ctx context.Context, reqMsg *RequestWhatsappMessage) (res *ResponseMessage, err error) {
	if reqMsg == nil {
		err = ErrNilArguments
		return
	}

	res, err = c.postToWappin(ctx, EndpointSendHSM, reqMsg.Default(c.opt))
	// create new token if we get invalid credential
	if res.Status == "401" {
		var tokenResp manager.ResponseGenerateToken
		var respToken *storage.Token

		tokenResp, err = c.GenerateToken(ctx)
		if err != nil {
			return
		}

		respToken, err = tokenResp.ToToken()
		if err != nil {
			return
		}

		err = c.opt.Storage.Save(ctx, c.opt.TokenCacheKey, respToken.SetHalfExpiredDate(time.Now()))

		// re-hit send message to Wappin
		return c.postToWappin(ctx, EndpointSendHSM, reqMsg.Default(c.opt))
	}

	return
}

func (c *client) prepareRequest(ctx context.Context, req *http.Request) *http.Request {
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return req.WithContext(ctx)
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

	token, err := c.opt.manager.Get(ctx)
	if err != nil {
		return
	}

	req.Header.Set(fiber.HeaderAuthorization, TokenBearer+token)
	resp, err := c.opt.client.Do(c.prepareRequest(ctx, req))
	if err != nil {
		return
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	byteBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(byteBody, &res)
	if err != nil {
		return
	}

	res.HttpStatusCode = resp.StatusCode
	res.RawData = convertByteToString(byteBody)
	err = getError(res.HttpStatusCode, res.Status, res.Message)
	return
}

// GenerateToken generates a token for Wappin.
func (c *client) GenerateToken(ctx context.Context) (res manager.ResponseGenerateToken, err error) {
	url := c.opt.BaseURL + EndpointToken
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return
	}

	req.SetBasicAuth(c.opt.ClientID, c.opt.SecretKey)
	resp, err := c.opt.client.Do(c.prepareRequest(ctx, req))
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
