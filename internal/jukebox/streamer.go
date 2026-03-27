package jukebox

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const audioChunkSize = 8192

// Streamer fetches MP3 data from track URLs and broadcasts to connected audio channels.
type Streamer struct {
	mu     sync.RWMutex
	conns  map[io.WriteCloser]bool
	cancel context.CancelFunc
	client *http.Client
}

// NewStreamer creates a new audio streamer.
func NewStreamer() *Streamer {
	return &Streamer{
		conns:  make(map[io.WriteCloser]bool),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// AddConn registers a new audio channel connection.
func (s *Streamer) AddConn(conn io.WriteCloser) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[conn] = true
}

// RemoveConn removes an audio channel connection.
func (s *Streamer) RemoveConn(conn io.WriteCloser) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, conn)
}

// ConnCount returns the number of connected audio channels.
func (s *Streamer) ConnCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.conns)
}

// StreamTrack starts streaming a track to all connected clients.
func (s *Streamer) StreamTrack(track Track) {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	go s.stream(ctx, track)
}

// Stop cancels the current stream.
func (s *Streamer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

func (s *Streamer) stream(ctx context.Context, track Track) {
	if track.URL == "" {
		return
	}

	s.broadcastHeader(track)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, track.URL, nil)
	if err != nil {
		log.Printf("streamer: request error: %v", err)
		return
	}

	resp, err := s.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		log.Printf("streamer: fetch error: %v", err)
		return
	}
	defer resp.Body.Close()

	buf := make([]byte, audioChunkSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			s.broadcastBytes(buf[:n])
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("streamer: read error: %v", err)
			}
			return
		}
	}
}

func (s *Streamer) broadcastHeader(track Track) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for conn := range s.conns {
		if err := EncodeTrackHeader(conn, track); err != nil {
			log.Printf("streamer: header write error: %v", err)
		}
	}
}

func (s *Streamer) broadcastBytes(data []byte) {
	s.mu.RLock()
	conns := make([]io.WriteCloser, 0, len(s.conns))
	for conn := range s.conns {
		conns = append(conns, conn)
	}
	s.mu.RUnlock()

	var failed []io.WriteCloser
	for _, conn := range conns {
		if _, err := conn.Write(data); err != nil {
			failed = append(failed, conn)
		}
	}

	if len(failed) > 0 {
		s.mu.Lock()
		for _, conn := range failed {
			delete(s.conns, conn)
			conn.Close()
		}
		s.mu.Unlock()
	}
}
