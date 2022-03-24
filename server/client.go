package server

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/smallnest/rpcx/server"
)

type Client interface {
	Set(k, v interface{})
	Get(k interface{}) interface{}
	Delete(k interface{}) interface{}

	Prefix() string
	Conn() net.Conn
	Authorize() interface{}
	Call(m string, req interface{}) error
}

type client struct {
	svr *server.Server

	conn      net.Conn
	prefix    string
	authorize interface{}

	mux    sync.RWMutex
	values map[interface{}]interface{}
}

func (c *client) Set(k, v interface{}) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.values[k] = v
}

func (c *client) Get(k interface{}) interface{} {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.values[k]
}

func (c *client) Delete(k interface{}) interface{} {
	c.mux.Lock()
	defer c.mux.Unlock()
	defer delete(c.values, k)
	return c.values[k]
}

func (c *client) Prefix() string {
	return c.prefix
}

func (c *client) Conn() net.Conn {
	return c.conn
}

func (c *client) Authorize() interface{} {
	return c.authorize
}

func (c *client) Call(m string, req interface{}) error {
	byts, err := json.Marshal(req)
	if err != nil {
		return errors.New("request not serialization")
	}
	return c.svr.SendMessage(c.conn, "", m, nil, byts)
}

func (c *client) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	for _, v := range c.values {
		if cl, ok := v.(io.Closer); ok {
			cl.Close()
		}
	}
	c.values = nil
	return nil
}
