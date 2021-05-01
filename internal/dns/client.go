package dns

import (
	"crypto/tls"
	"errors"
	"fmt"
	miekg "github.com/miekg/dns"
	"sync"
	"sync/atomic"
)

const (
	NET_UDP = "udp"
	NET_TCP = "tcp"
	NET_TLS = "tcp-tls"

	CH_SIZE = 100
)

type id uint64

type Query struct {
	ID       id
	Question *miekg.Msg
}

type Answer struct {
	ID         id
	ServerAddr string
	Ans        *miekg.Msg
	Err        error
}

type Client interface {
	Concurrency() uint32
	Submit(query Query)
	ResultsCh() <-chan *Answer
	ErrCh() <-chan error
	Stop() bool
	//TODO: Cap()
	//TODO: Len()
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

func NewDnsClient(config Config) (Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return newClient(&config), nil
}

type client struct {
	remoteAddr  string
	cli         *miekg.Client
	concurrency uint32
	pool        sync.Pool
	respCh      chan *Answer
	errCh       chan error

	stopped      uint32
	stoppingLock sync.RWMutex
}

func newClient(config *Config) *client {
	c := &client{
		remoteAddr: config.RemoteAddr,
		respCh:     make(chan *Answer, CH_SIZE),
		errCh:      make(chan error, CH_SIZE),
	}
	c.cli = &miekg.Client{
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
			return err
		}
		return conn
	}}

	return c
}

func (c *client) Concurrency() uint32 {
	return atomic.LoadUint32(&c.concurrency)
}

func (c *client) ResultsCh() <-chan *Answer {
	return c.respCh
}

func (c *client) ErrCh() <-chan error {
	return c.errCh
}

func (c *client) Stop() bool {
	if atomic.CompareAndSwapUint32(&c.stopped, 0, 1) {
		c.stoppingLock.Lock()
		close(c.respCh)
		close(c.errCh)
		c.stoppingLock.Unlock()
		return true
	}
	return false
}

func (c *client) Submit(req Query) {
	atomic.AddUint32(&c.concurrency, 1)
	go func() {
		defer atomic.AddUint32(&c.concurrency, ^uint32(0))

		c.stoppingLock.RLock()
		defer c.stoppingLock.RUnlock()

		connOrErr := c.pool.Get()
		if err, ok := connOrErr.(error); ok {
			c.errCh <- err
			c.respCh <- &Answer{
				ID:         req.ID,
				ServerAddr: c.remoteAddr,
				Ans:        nil,
				Err:        err,
			}
			return
		}
		conn := connOrErr.(*miekg.Conn)
		ans, _, err := c.cli.ExchangeWithConn(req.Question, conn)
		if err != nil {
			c.errCh <- err
			conn.Close()
		} else {
			c.pool.Put(conn)
		}

		select {
		case c.respCh <- &Answer{ID: req.ID, ServerAddr: c.remoteAddr, Ans: ans, Err: err}:
		default:
			c.errCh <- errors.New("worker result chan is full")
		}
	}()
}

func (c *client) dial() (*miekg.Conn, error) {
	return c.cli.Dial(c.remoteAddr)
}
