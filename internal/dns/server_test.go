package dns

//func Test_Server_Run(t *testing.T) {
//	defer func() {
//		if err := recover(); err != nil {
//			t.Fatal(err)
//		}
//	}()
//	serverAddr := "127.0.0.1:11153"
//	net := "udp"
//	var f dns.HandlerFunc = func(writer dns.ResponseWriter, msg *dns.Msg) {
//		if err := writer.WriteMsg(msg); err != nil {
//			t.Fatal(err)
//		}
//	}
//	server := newServer(net, serverAddr, f)
//	ctx, cancel := context.WithCancel(context.Background())
//	var wg sync.WaitGroup
//	syncCh := make(chan struct{})
//	wg.Add(2)
//
//	go func() {
//		defer wg.Done()
//		go server.Run()
//		time.Sleep(500 * time.Millisecond)
//		syncCh <- struct{}{}
//		<-ctx.Done()
//		err := server.Shutdown()
//		if err != nil {
//			panic(err)
//		}
//	}()
//
//	go func() {
//		defer wg.Done()
//		cli := &dns.Client{
//			Net: net,
//		}
//		var wg2 sync.WaitGroup
//		<-syncCh
//		for _, name := range names {
//			msg := &dns.Msg{}
//			msg.SetQuestion(name+".", dns.TypeA)
//			wg2.Add(1)
//			go func() {
//				defer wg2.Done()
//				if _, _, err := cli.Exchange(msg, serverAddr); err != nil {
//					panic(err)
//				}
//			}()
//		}
//		wg2.Wait()
//		cancel()
//	}()
//	wg.Wait()
//}
