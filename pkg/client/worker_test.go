package client

import (
	"github.com/miekg/dns"
	"math/rand"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var (
	configUDP = &Config{
		Net:           NET_UDP,
		RemoteAddr:    "114.114.114.114:53",
		TlsServerName: "",
	}

	configTLS = &Config{
		Net:           NET_TLS,
		RemoteAddr:    "8.8.8.8:853",
		TlsServerName: "dns.google",
	}

	configTCP = &Config{
		Net:        NET_TCP,
		RemoteAddr: "208.67.222.222:5353",
	}

	names = []string{
		"www.google.com.hk",
		"www.163.com",
		"news.163.com",
		"www.shef.ac.uk",
		"www.twitter.com",
		"www.taobao.com",
		"www.pornhub.com",
		"www.github.com",
		"www.6park.com",
		"www.youtube.com",
		"www.126.com",
		"www.ox.ac.uk",
		"www.bbc.co.uk",
		"www.facebook.com",
		"www.apple.com",
		"web.telegram.org",
	}
)

func newQuestion(name string) *dns.Msg {
	if !strings.HasSuffix(name, ".") {
		name = name + "."
	}
	m := &dns.Msg{}
	return m.SetQuestion(name, dns.TypeA)
}

func Test_newClient(t *testing.T) {
	_ = newClient(configUDP)
	_ = newClient(configTLS)
}

func Test_client_dial(t *testing.T) {
	udpClient := newClient(configUDP)
	if _, err := udpClient.dial(); err != nil {
		t.Fatal(err)
	}

	tlsClient := newClient(configTLS)
	if _, err := tlsClient.dial(); err != nil {
		t.Fatal(err)
	}

	tcpClient := newClient(configTCP)
	if _, err := tcpClient.dial(); err != nil {
		t.Fatal(err)
	}
}

func Test_client_Submit(t *testing.T) {
	udpClient := newClient(configUDP)
	tcpClient := newClient(configTCP)
	tlsClient := newClient(configTLS)
	test_submit(udpClient, t)
	test_submit(tcpClient, t)
	test_submit(tlsClient, t)
}

func test_submit(cli *client, t *testing.T) {
	for _, name := range names {
		cli.Submit(Req{
			ID:    rand.Uint64(),
			Query: newQuestion(name),
		})
		time.Sleep(time.Microsecond)
	}
	go func() {
		for atomic.LoadUint32(&cli.concurrency) != 0 {
			time.Sleep(time.Microsecond)
		}
		cli.Stop()
	}()
	for resp := range cli.ResultsCh() {
		if resp.Err != nil {
			t.Fatal(resp.Err)
		}
	}
}
