package server

import "net"

type plugin struct {
	Closer func(net.Conn)
}

func (p *plugin) HandleConnClose(c net.Conn) bool {
	if p.Closer != nil {
		p.Closer(c)
	}
	return true
}
