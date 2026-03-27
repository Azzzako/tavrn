package jukebox

import (
	"bytes"
	"testing"
)

func TestProtocolEncodeHeader(t *testing.T) {
	track := Track{
		ID:       "123",
		Title:    "Test Song",
		Artist:   "Test Artist",
		Duration: 180,
		Source:   "jamendo",
	}
	var buf bytes.Buffer
	err := EncodeTrackHeader(&buf, track)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty buffer")
	}
}

func TestProtocolDecodeHeader(t *testing.T) {
	track := Track{
		ID:       "123",
		Title:    "Test Song",
		Artist:   "Test Artist",
		Duration: 180,
		Source:   "jamendo",
	}
	var buf bytes.Buffer
	EncodeTrackHeader(&buf, track)

	decoded, err := DecodeTrackHeader(&buf)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded.Title != "Test Song" {
		t.Errorf("expected 'Test Song', got '%s'", decoded.Title)
	}
	if decoded.Artist != "Test Artist" {
		t.Errorf("expected 'Test Artist', got '%s'", decoded.Artist)
	}
	if decoded.Duration != 180 {
		t.Errorf("expected duration 180, got %d", decoded.Duration)
	}
}

func TestProtocolRoundTrip(t *testing.T) {
	track := Track{
		ID:       "456",
		Title:    "Another Song",
		Artist:   "Another Artist",
		Duration: 240,
		Source:   "radio",
	}

	var buf bytes.Buffer
	EncodeTrackHeader(&buf, track)

	audioData := []byte("fake mp3 data here")
	buf.Write(audioData)

	decoded, err := DecodeTrackHeader(&buf)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded.ID != "456" {
		t.Errorf("expected ID '456', got '%s'", decoded.ID)
	}

	remaining := buf.Bytes()
	if string(remaining) != "fake mp3 data here" {
		t.Errorf("expected audio data remaining, got '%s'", string(remaining))
	}
}
