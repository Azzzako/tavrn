package session

import (
	"tavrn.sh/internal/ratelimit"
)

type MsgType int

const (
	MsgChat MsgType = iota
	MsgSystem
	MsgCanvasDelta
	MsgUserJoined
	MsgUserLeft
	MsgPurge
	MsgTyping
	MsgNoteCreate
	MsgNoteMove
	MsgNoteDelete
	MsgJukeboxUpdate
	MsgBanner
)

// NoteData carries sticky note info through the hub.
type NoteData struct {
	ID    int
	X, Y  int
	Text  string
	Nick  string
	Color int
}

// Msg is a message sent from the hub to a session.
type Msg struct {
	Type        MsgType
	Nickname    string
	Fingerprint string
	ColorIndex  int
	Text        string
	Room        string
	Note        *NoteData
}

// Session represents a connected user.
type Session struct {
	Fingerprint string
	Nickname    string
	Flair       bool
	Room        string
	ColorIndex  int
	Send        chan Msg
	ChatLimiter *ratelimit.Limiter
}

// New creates a session with a buffered send channel (256 messages).
func New(fingerprint, nickname string, colorIndex int, flair bool) *Session {
	return &Session{
		Fingerprint: fingerprint,
		Nickname:    nickname,
		Flair:       flair,
		Room:        "lounge",
		ColorIndex:  colorIndex,
		Send:        make(chan Msg, 256),
		ChatLimiter: ratelimit.NewChat(),
	}
}
