package poll

import (
	"sync"
	"time"
)

// Poll represents a room poll with up to 4 options.
type Poll struct {
	ID          int
	Room        string
	Creator     string // fingerprint
	CreatorNick string
	Title       string
	Options     []string       // 2-4 options
	Votes       map[string]int // fingerprint → option index
	CreatedAt   time.Time
	Closed      bool
}

// VoteCount returns the number of votes for each option.
func (p *Poll) VoteCount() []int {
	counts := make([]int, len(p.Options))
	for _, idx := range p.Votes {
		if idx >= 0 && idx < len(counts) {
			counts[idx]++
		}
	}
	return counts
}

// TotalVotes returns the total number of votes cast.
func (p *Poll) TotalVotes() int {
	return len(p.Votes)
}

// Store holds all polls in memory. Thread-safe.
type Store struct {
	mu     sync.RWMutex
	polls  []Poll
	nextID int
}

func NewStore() *Store {
	return &Store{nextID: 1}
}

func (s *Store) Create(room, fingerprint, nickname, title string, options []string) *Poll {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := Poll{
		ID:          s.nextID,
		Room:        room,
		Creator:     fingerprint,
		CreatorNick: nickname,
		Title:       title,
		Options:     options,
		Votes:       make(map[string]int),
		CreatedAt:   time.Now(),
	}
	s.nextID++
	s.polls = append(s.polls, p)
	return &s.polls[len(s.polls)-1]
}

func (s *Store) Vote(pollID int, fingerprint string, optionIndex int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.polls {
		if s.polls[i].ID == pollID && !s.polls[i].Closed {
			if optionIndex >= 0 && optionIndex < len(s.polls[i].Options) {
				s.polls[i].Votes[fingerprint] = optionIndex
				return true
			}
		}
	}
	return false
}

// Close closes a poll. Only the creator can close it.
func (s *Store) Close(pollID int, fingerprint string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.polls {
		if s.polls[i].ID == pollID && s.polls[i].Creator == fingerprint && !s.polls[i].Closed {
			s.polls[i].Closed = true
			return true
		}
	}
	return false
}

// RoomPolls returns all polls (active + closed) for a room.
func (s *Store) RoomPolls(room string) []Poll {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Poll
	for _, p := range s.polls {
		if p.Room == room {
			result = append(result, p)
		}
	}
	return result
}

// ActiveRoomPolls returns only active (not closed) polls for a room.
func (s *Store) ActiveRoomPolls(room string) []Poll {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Poll
	for _, p := range s.polls {
		if p.Room == room && !p.Closed {
			result = append(result, p)
		}
	}
	return result
}

// LatestByCreator returns the most recent active poll by a fingerprint in a room.
func (s *Store) LatestByCreator(room, fingerprint string) *Poll {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.polls) - 1; i >= 0; i-- {
		if s.polls[i].Room == room && s.polls[i].Creator == fingerprint && !s.polls[i].Closed {
			return &s.polls[i]
		}
	}
	return nil
}

// Get returns a poll by ID.
func (s *Store) Get(pollID int) *Poll {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.polls {
		if s.polls[i].ID == pollID {
			return &s.polls[i]
		}
	}
	return nil
}

// Clear removes all polls (called on purge).
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.polls = nil
}
