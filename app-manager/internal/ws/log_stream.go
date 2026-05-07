package ws

import (
	"bufio"
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func StreamFromReader(w http.ResponseWriter, r *http.Request, reader io.ReadCloser) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	defer conn.Close()
	defer reader.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		default:
		}

		line := stripDockerLogHeader(scanner.Bytes())
		if err := conn.WriteMessage(websocket.TextMessage, line); err != nil {
			return
		}
	}
}

func StreamFromChannel(w http.ResponseWriter, r *http.Request, ch <-chan string) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case line, ok := <-ch:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
				return
			}
		}
	}
}

func StreamDockerLogs(w http.ResponseWriter, r *http.Request, getReader func(ctx context.Context, tail int) (io.ReadCloser, error), tail int) {
	if tail <= 0 {
		tail = 100
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	reader, err := getReader(ctx, tail)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error: "+err.Error()))
		return
	}
	defer reader.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		default:
		}

		line := stripDockerLogHeader(scanner.Bytes())
		conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, line); err != nil {
			return
		}
	}
}

func stripDockerLogHeader(line []byte) []byte {
	if len(line) < 8 {
		return line
	}

	header := line[:8]
	if header[0] == 1 || header[0] == 2 {
		return line[8:]
	}
	return line
}

func SendHistoryAndStream(w http.ResponseWriter, r *http.Request, history []string, streamCh <-chan string) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	for _, line := range history {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			return
		}
	}

	if streamCh == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case line, ok := <-streamCh:
			if !ok {
				return
			}
			conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
				return
			}
		}
	}
}
