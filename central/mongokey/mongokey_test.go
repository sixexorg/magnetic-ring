package mongokey

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestChan(t *testing.T)  {
	ch := make(chan string,1)

	fmt.Printf("begin......\n")
	go func() {
		ticker := time.NewTicker(time.Second*5)

		for  {
			select {
			case <- ticker.C:
				ch <- time.Now().Format("15:04:05")
			}
		}

	}()

	for c := range ch {
		fmt.Printf("c=%s\n",c)
	}



	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-c

}


