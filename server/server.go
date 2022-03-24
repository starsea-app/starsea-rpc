package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
)

const AuthorizeKey = "__AUTHORIZE"

type OptionFn = server.OptionFn
type Context = server.Context

type Authorize interface {
	Authorize(prefix, token string) interface{}
}

type Server struct {
	svr  *server.Server
	auth Authorize

	mux   sync.RWMutex
	conns map[string]*client
}

func NewServer(options ...OptionFn) *Server {
	s := server.NewServer(options...)

	return &Server{svr: s, conns: make(map[string]*client)}
}

func (s *Server) Serve(addr string) error {
	s.svr.AuthFunc = s.authorize
	s.svr.Plugins.Add(s)

	if !strings.Contains(addr, "@") {
		return errors.New("addr is invalid")
	}
	strs := strings.Split(addr, "@")
	go s.svr.Serve(strs[0], strs[1])
	return nil
}

func (s *Server) Call(prefix, m string, req interface{}) error {
	s.mux.RLock()
	for k, cli := range s.conns {
		if k == prefix {
			s.mux.RUnlock()
			return cli.Call(m, req)
		}
	}
	s.mux.RUnlock()
	return errors.New("prefix not found")
}

func (s *Server) Handle(m string, fn interface{}) {
	s.svr.AddHandler("starsea.platform", m, func(ctx *server.Context) error {
		return s.handle(ctx, fn)
	})
}

func (s *Server) handle(c *server.Context, h interface{}) error {
	defer func() {
		if err := recover(); err != nil {
			c.Write(map[string]interface{}{"code": 500, "msg": err})
		}
	}()

	fmt.Printf("<-- %v.%v \n", c.ServicePath(), c.ServiceMethod())

	var result interface{}
	switch fn := h.(type) {
	case func(Client):
		if cli, ok := c.Get(AuthorizeKey).(Client); ok {
			fn(cli)
			return nil
		}
		return errors.New("authorize is emptry")

	case func(*server.Context):
		fn(c)
		return nil

	case func(*server.Context, Client):
		if cli, ok := c.Get(AuthorizeKey).(Client); ok {
			fn(c, cli)
			return nil
		}
		return errors.New("authorize is emptry")

	case func(Client) interface{}:
		if cli, ok := c.Get(AuthorizeKey).(Client); ok {
			result = fn(cli)
		} else {
			result = errors.New("authorize is emptry")
		}

	case func(*server.Context) interface{}:
		result = fn(c)

	case func(*server.Context, Client) interface{}:
		if cli, ok := c.Get(AuthorizeKey).(Client); ok {
			result = fn(c, cli)
		} else {
			result = errors.New("authorize is emptry")
		}

	}

	switch v := result.(type) {
	case error:
		c.Write(map[string]interface{}{"code": 500, "msg": v.Error()})

	case func(*server.Context):
		v(c)

	default:
		dict := map[string]interface{}{"code": 200}
		if v != nil {
			dict["data"] = v
		}
		byts, _ := json.Marshal(dict)
		fmt.Printf("--> %v.%v %v\n", c.ServicePath(), c.ServiceMethod(), string(byts))
		c.Write(dict)
	}
	return nil
}

func (s *Server) Authorize(auth Authorize) {
	s.auth = auth
}

func (s *Server) authorize(ctx context.Context, r *protocol.Message, token string) error {
	if !strings.Contains(token, "@") {
		return errors.New("token is invalid")
	}
	list := strings.Split(token, "@")
	s.mux.RLock()
	if cli, ok := s.conns[list[0]]; ok {
		s.mux.RUnlock()
		if sctx, ok := ctx.(*share.Context); ok {
			sctx.SetValue(AuthorizeKey, cli)
		}
		return nil
	}
	s.mux.RUnlock()

	if s.auth == nil {
		return errors.New("authorize is nil")
	}
	if v := s.auth.Authorize(list[0], list[1]); v != nil {
		conn := ctx.Value(server.RemoteConnContextKey).(net.Conn)
		s.mux.Lock()
		defer s.mux.Unlock()
		s.conns[list[0]] = &client{svr: s.svr, conn: conn, prefix: list[0], authorize: v, values: make(map[interface{}]interface{})}
		if sctx, ok := ctx.(*share.Context); ok {
			sctx.SetValue(AuthorizeKey, s.conns[list[0]])
		}
		return nil
	}
	return errors.New("authorize fail")
}

func (s *Server) HandleConnClose(conn net.Conn) bool {
	s.mux.Lock()
	for k, cli := range s.conns {
		if cli.conn == conn {
			delete(s.conns, k)
			s.mux.Unlock()
			cli.Close()
			return true
		}
	}
	s.mux.Unlock()
	return true
}
