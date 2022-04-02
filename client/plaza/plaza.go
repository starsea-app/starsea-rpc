package plaza

import "github.com/starsea-app/starsea-rpc/client"

type (
	H       = client.H
	Context = client.Context
)

type Plaza struct {
	*client.Client
}

func NewPlaza(addr string) *Plaza {
	return &Plaza{Client: client.NewClient(addr)}
}

func (p *Plaza) Login(token string, rsp interface{}) error {
	return p.Call("login", H{"token": token}, rsp)
}

func (p *Plaza) Assets(address string, froms []string, rsp interface{}) error {
	return p.Call("assets", H{"address": address, "tokens": froms}, &rsp)
}

func (p *Plaza) Router(notice string, rsp interface{}) error {
	return p.Call("router", H{"notice": notice}, rsp)
}

func (p *Plaza) Listen(addr string) error {
	return p.Call("listen", H{"listen": addr}, nil)
}

func (p *Plaza) Universes(rsp interface{}) error {
	return p.Call("univs", H{}, rsp)
}

func (p *Plaza) Subscribe(typ uint8, token string, publish string, rsp interface{}) error {
	return p.Call("subscribe", H{"type": typ, "publish": publish, "data": H{"token": token}}, &rsp)
}

func (p *Plaza) Onlines(typ uint8, time int64, rsp interface{}) error {
	return p.Call("onlines", H{"type": typ, "time": time}, rsp)
}
