package jukebox

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// TrackHeader is the metadata sent before MP3 bytes for each track.
type TrackHeader struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Duration int    `json:"duration"`
	Source   string `json:"source"`
}

// EncodeTrackHeader writes a track metadata header to w.
// Format: [4 bytes big-endian length][JSON bytes]
func EncodeTrackHeader(w io.Writer, track Track) error {
	header := TrackHeader{
		ID:       track.ID,
		Title:    track.Title,
		Artist:   track.Artist,
		Duration: track.Duration,
		Source:   track.Source,
	}
	data, err := json.Marshal(header)
	if err != nil {
		return fmt.Errorf("marshal track header: %w", err)
	}

	length := uint32(len(data))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return fmt.Errorf("write header length: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write header data: %w", err)
	}

	return nil
}

// DecodeTrackHeader reads a track metadata header from r.
func DecodeTrackHeader(r io.Reader) (TrackHeader, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return TrackHeader{}, fmt.Errorf("read header length: %w", err)
	}

	if length > 1<<20 {
		return TrackHeader{}, fmt.Errorf("header too large: %d bytes", length)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return TrackHeader{}, fmt.Errorf("read header data: %w", err)
	}

	var header TrackHeader
	if err := json.Unmarshal(data, &header); err != nil {
		return TrackHeader{}, fmt.Errorf("unmarshal track header: %w", err)
	}

	return header, nil
}
