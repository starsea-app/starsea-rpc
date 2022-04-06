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

func (p *Plaza) Router(notice string, rsp interface{}, args ...string) error {
	values := H{"notice": notice}
	for i := 0; i < len(args)/2; i++ {
		values[args[i*2]] = args[i*2+1]
	}
	return p.Call("router", values, rsp)
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

func (p *Plaza) Onlines(typ uint8, uid uint, time int64, rsp interface{}) error {
	return p.Call("onlines", H{"type": typ, "uid": uid, "time": time}, rsp)
}
