package client

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/share"
)

type Context struct {
	msg *protocol.Message
}

type connPlugin func(net.Conn)

func (c *connPlugin) ClientConnected(conn net.Conn) (net.Conn, error) {
	defer recover()
	(*c)(conn)
	return conn, nil
}

type closePlugin func(net.Conn)

func (c *closePlugin) ClientConnectionClose(conn net.Conn) error {
	defer recover()
	(*c)(conn)
	return nil
}

func (c *Context) Bind(v interface{}) error {
	return json.Unmarshal(c.msg.Payload, v)
}

type Client struct {
	cli     client.XClient
	routers map[string]func(*Context)
	receive chan *protocol.Message
}

func NewClient(addr string) *Client {
	d, _ := client.NewPeer2PeerDiscovery(addr, "")
	ch := make(chan *protocol.Message, 1000)
	c := client.NewBidirectionalXClient("starsea.platform", client.Failtry, client.RandomSelect, d, client.Option{
		Retries:             3,
		RPCPath:             share.DefaultRPCPath,
		ConnectTimeout:      time.Second,
		SerializeType:       protocol.JSON,
		CompressType:        protocol.None,
		BackupLatency:       10 * time.Millisecond,
		MaxWaitForHeartbeat: 30 * time.Second,
		TCPKeepAlivePeriod:  time.Minute,
	}, ch)

	return &Client{cli: c, routers: make(map[string]func(*Context)), receive: ch}
}

func (c *Client) Connect() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	var rsp string
	err := c.cli.Call(ctx, "_HELLO", nil, &rsp)
	if err != nil {
		return err
	}
	if rsp != "OK" {
		return errors.New("reply is invalid")
	}
	return nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) Handle(m string, fn func(*Context)) {
	c.routers[m] = fn
}

func (c *Client) HandleConnected(fn func(net.Conn)) {
	if fn != nil {
		c.cli.GetPlugins().Add((*connPlugin)(&fn))
	}
}

func (c *Client) HandleDisconnected(fn func(net.Conn)) {
	if fn != nil {
		c.cli.GetPlugins().Add((*closePlugin)(&fn))
	}
}

func (c *Client) Receive(quit chan struct{}) {
	for {
		select {
		case rev := <-c.receive:
			c.handleReceive(rev)

		case <-quit:
			return
		}
	}
}

func (c *Client) handleReceive(rev *protocol.Message) {
	if handler, ok := c.routers[rev.ServiceMethod]; ok {
		handler(&Context{msg: rev})
	}
}

func (c *Client) Authorize(token string) {
	c.cli.Auth(token)
}

func (c *Client) Call(m string, req, rsp interface{}) error {
	if c.cli == nil {
		return errors.New("client is nil")
	}
	var result struct {
		Code uint
		Msg  string
		Data json.RawMessage
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	err := c.cli.Call(ctx, m, req, &result)
	if err != nil {
		return err
	}
	if result.Code != 200 {
		return errors.New(result.Msg)
	}
	if rsp != nil && result.Data != nil {
		return json.Unmarshal(result.Data, rsp)
	}
	return nil
}
