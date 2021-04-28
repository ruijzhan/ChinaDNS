package client

import (
	"crypto/tls"
	"errors"
	"github.com/miekg/dns"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const RES_CH_SIZE = 100

type client struct {
	config      *Config
	cli         *dns.Client
	concurrency uint32
	pool        sync.Pool
	respCh      chan *Resp
}

func newClient(config *Config) *client {
	if err := config.Validate(); err != nil {
		log.Fatal(err)
	}
	c := &client{config: config,
		respCh: make(chan *Resp, RES_CH_SIZE),
	}
	c.cli = &dns.Client{
		Net: config.Net,
		TLSConfig: func() *tls.Config {
			return &tls.Config{
				ServerName: config.TlsServerName,
			}
		}(),
	}

	c.pool = sync.Pool{New: func() interface{} {
		conn, err := c.dial()
		if err != nil {
			log.Println("dial error: ", err) //TODO: 替换 log 模块
			return nil
		}
		return conn
	}}

	return c
}

func (c *client) Concurrency() uint32 {
	return atomic.LoadUint32(&c.concurrency)
}

func (c *client) ResultsCh() <-chan *Resp {
	return c.respCh
}

func (c *client) Stop() {
	close(c.respCh)
}

func (c *client) Submit(req Req) {
	atomic.AddUint32(&c.concurrency, 1)
	go func() {
		defer atomic.AddUint32(&c.concurrency, ^uint32(0))
		conn := c.pool.Get().(*dns.Conn)
		if conn == nil {
			c.respCh <- &Resp{
				ID:  req.ID,
				Ans: nil,
				Err: errors.New("remote server connection failure"),
			}
			return
		}
		ans, _, err := c.cli.ExchangeWithConn(req.Query, conn)
	Loop:
		for {
			select {
			case c.respCh <- &Resp{ID: req.ID, Ans: ans, Err: err}:
				break Loop
			default:
				log.Println("warning: worker result chan is full")
			}
			time.Sleep(time.Microsecond)
		}
		if err == nil {
			c.pool.Put(conn)
		}
	}()
}

func (c *client) dial() (*dns.Conn, error) {
	return c.cli.Dial(c.config.RemoteAddr)
}
