package dns

import (
	"math/rand"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	miekg "github.com/miekg/dns"
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
		"webk.telegram.org",
	}
)

func newQuestion(name string) *miekg.Msg {
	if !strings.HasSuffix(name, ".") {
		name = name + "."
	}
	m := &miekg.Msg{}
	return m.SetQuestion(name, miekg.TypeA)
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
	test_submit(udpClient, t)
	tcpClient := newClient(configTCP)
	test_submit(tcpClient, t)
	tlsClient := newClient(configTLS)
	test_submit(tlsClient, t)
}

func test_submit(cli *client, t *testing.T) {
	for _, name := range names {
		cli.Submit(Query{
			ID:       id(rand.Uint64()),
			Question: newQuestion(name),
		})
		time.Sleep(time.Microsecond)
	}
	go func() {
		for atomic.LoadUint32(&cli.concurrency) != 0 {
			time.Sleep(time.Microsecond)
		}
		cli.Stop()
	}()

	go func() {
		for range cli.respCh {
		}
	}()

	for err := range cli.ErrCh() {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func Test_client_Stop(t *testing.T) {
	tcpClient := newClient(configTCP)
	if tcpClient.Stop() != true {
		t.Fatal()
	}
	if tcpClient.Stop() != false {
		t.Fail()
	}
}
