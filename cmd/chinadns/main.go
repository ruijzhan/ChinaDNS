package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ruijzhan/chinaDNS/internal/chinadns"
)

func main() {
	s := chinadns.NewChinaDNS("0.0.0.0:1153", "udp")
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		if err := s.Shutdown(); err != nil {
			log.Fatal(err)
		}
	}()
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
