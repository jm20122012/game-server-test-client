package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	CLIENT_COUNT int = 1000
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < CLIENT_COUNT; i++ {
		wg.Add(1)
		r := rand.Intn(5) + 1
		slog.Info("Starting client", "clientID", i, "runDuration", r)
		go runClient(&wg, i, time.Duration(r)*time.Second)
	}

	wg.Wait()
}

func runClient(wg *sync.WaitGroup, id int, d time.Duration) {
	defer wg.Done()

	conn, err := net.Dial("tcp", "localhost:7000")
	if err != nil {
		slog.Error("error creating TCP connection in runClient", "error", err)
		return
	}
	defer func() {
		// Gracefully shut down the write side of the connection
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
		err = conn.Close()
		if err != nil {
			slog.Error("error closing tcp connection", "clientID", id, "error", err)
		} else {
			slog.Info("TCP connection closed", "clientID", id)
		}
	}()

	timer := time.NewTimer(d)
	defer timer.Stop()

	ticker := time.NewTicker(time.Millisecond * 250)

	for {
		select {
		case <-timer.C:
			slog.Info("Client timer done", "clientID", id)
			ticker.Stop()
			_, err := conn.Write([]byte("\x17\x04"))
			if err != nil {
				slog.Error("error sending data to server", "clientID", id, "error", err)
			}
			return
		case <-ticker.C:
			m := fmt.Sprintf("%d,hello,\x04", id)
			_, err := conn.Write([]byte(m))
			if err != nil {
				slog.Error("error sending data to server", "clientID", id, "error", err)
			}
		}
	}
}
