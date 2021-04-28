package client

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
)

const (
	NET_UDP = "udp"
	NET_TCP = "tcp"
	NET_TLS = "tcp-tls"
)

type Req struct {
	ID    uint64
	Query *dns.Msg
}

type Resp struct {
	ID  uint64
	Ans *dns.Msg
	Err error
}

type Config struct {
	Net           string
	RemoteAddr    string
	TlsServerName string
}

func (c *Config) Validate() error {
	switch c.Net {
	case NET_UDP, NET_TCP, NET_TLS:
	default:
		return fmt.Errorf("invalid network type, must be one of: %s, %s, %s", NET_UDP, NET_TCP, NET_TLS)
	}

	if c.Net == NET_TLS && c.TlsServerName == "" {
		return errors.New("no TLS server name specified")
	}

	return nil
}
