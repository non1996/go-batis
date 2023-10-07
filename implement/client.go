package implement

import (
	"context"
)

type Client struct {
	w Writer // 写client
	r Reader // 读client
}

func NewClient(w Writer, r Reader) Client {
	return Client{
		w: w,
		r: r,
	}
}

func (c *Client) R(session Session) Session {
	if _, ok := session.Value(txKey).(Reader); ok {
		return session
	}
	return context.WithValue(session, txKey, c.r)
}

func (c *Client) W(session Session) Session {
	if _, ok := session.Value(txKey).(Writer); ok {
		return session
	}
	return context.WithValue(session, txKey, c.w)
}
