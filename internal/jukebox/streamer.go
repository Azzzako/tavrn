package jukebox

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"sync"
)

const audioChunkSize = 8192

// Streamer fetches MP3 data from track URLs and broadcasts to connected audio channels.
// It buffers the current track so late-joining clients receive the full audio.
type Streamer struct {
	mu           sync.RWMutex
	conns        map[io.WriteCloser]bool
	cancel       context.CancelFunc
	currentTrack *Track
	audioBuffer  *bytes.Buffer // buffered MP3 data for current track
	bufferReady  bool          // true when download is complete
	client       *http.Client
}

// NewStreamer creates a new audio streamer.
func NewStreamer() *Streamer {
	return &Streamer{
		conns:       make(map[io.WriteCloser]bool),
		audioBuffer: &bytes.Buffer{},
		client:      &http.Client{},
	}
}

// AddConn registers a new audio channel connection.
// Sends the current track header + any buffered audio immediately.
func (s *Streamer) AddConn(conn io.WriteCloser) {
	s.mu.Lock()
	track := s.currentTrack
	var audioData []byte
	if s.audioBuffer.Len() > 0 {
		audioData = make([]byte, s.audioBuffer.Len())
		copy(audioData, s.audioBuffer.Bytes())
	}
	s.conns[conn] = true
	s.mu.Unlock()

	if track != nil {
		// Send header
		if err := EncodeTrackHeader(conn, *track); err != nil {
			log.Printf("streamer: header write to new conn: %v", err)
			return
		}
		// Send buffered audio so the client catches up
		if len(audioData) > 0 {
			if _, err := conn.Write(audioData); err != nil {
				log.Printf("streamer: buffer write to new conn: %v", err)
			}
		}
	}
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
	s.currentTrack = &track
	s.audioBuffer = &bytes.Buffer{}
	s.bufferReady = false

	// Send header to all currently connected clients
	for conn := range s.conns {
		if err := EncodeTrackHeader(conn, track); err != nil {
			log.Printf("streamer: header write error: %v", err)
		}
	}
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
	s.currentTrack = nil
	s.audioBuffer = &bytes.Buffer{}
}

func (s *Streamer) stream(ctx context.Context, track Track) {
	if track.URL == "" {
		return
	}

	// Fetch MP3 data (header already sent by StreamTrack)
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

	log.Printf("streamer: downloading %s (%s)", track.Title, track.URL)

	// Read and broadcast chunks, also buffer for late joiners
	buf := make([]byte, audioChunkSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]

			// Buffer it
			s.mu.Lock()
			s.audioBuffer.Write(chunk)
			s.mu.Unlock()

			// Broadcast to live clients
			s.broadcastBytes(chunk)
		}
		if err != nil {
			if err == io.EOF {
				log.Printf("streamer: download complete (%d bytes buffered)", s.audioBuffer.Len())
			} else {
				log.Printf("streamer: read error: %v", err)
			}

			s.mu.Lock()
			s.bufferReady = true
			s.mu.Unlock()
			return
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
