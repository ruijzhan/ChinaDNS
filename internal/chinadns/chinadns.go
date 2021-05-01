package chinadns

import (
	"errors"

	"github.com/miekg/dns"
)

type ChinaDNS struct {
	listenAddr string
	network    string
	server     *dns.Server
}

func NewChinaDNS(listenAddr string, network string) *ChinaDNS {
	return &ChinaDNS{
		listenAddr: listenAddr,
		network:    network,
	}
}

func (c *ChinaDNS) ListenAndServe() error {
	if c.server != nil {
		return errors.New("server already listening")
	}
	c.server = &dns.Server{Addr: c.listenAddr,
		Net:     c.network,
		Handler: dns.HandlerFunc(c.Handle)}
	return c.server.ListenAndServe()
}

func (c *ChinaDNS) Shutdown() error {
	defer func() {
		c.server = nil
	}()
	return c.server.Shutdown()
}

func (c *ChinaDNS) Handle(w dns.ResponseWriter, r *dns.Msg) {
	w.WriteMsg(r)
}
