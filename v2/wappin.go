package v2

import (
	"context"
)

type Client interface {
	SendMessage(ctx context.Context, reqMsg *RequestMessage) (res *ResponseMessage, err error)
}

type client struct {
	opt *Option
}

func (c *client) Assign(o *Option) *client {
	if o == nil {
		return c
	}

	c.opt = o.Clone()
	return c
}
