package client

type DnsClient interface {
	Concurrency() uint32
	Submit(Req)
	ResultsCh() <-chan *Resp
	Stop()
}

func NewDnsClient(config Config) DnsClient {
	return newClient(&config)
}
