package poll

import "testing"

func TestCreateAndVote(t *testing.T) {
	s := NewStore()
	p := s.Create("lounge", "fp1", "alice", "Best room?", []string{"lounge", "gallery"})
	if p.ID != 1 {
		t.Errorf("ID = %d, want 1", p.ID)
	}
	if !s.Vote(p.ID, "fp2", 0) {
		t.Error("Vote should succeed")
	}
	if !s.Vote(p.ID, "fp3", 1) {
		t.Error("Vote should succeed")
	}
	counts := s.Get(p.ID).VoteCount()
	if counts[0] != 1 || counts[1] != 1 {
		t.Errorf("counts = %v, want [1 1]", counts)
	}
}

func TestChangeVote(t *testing.T) {
	s := NewStore()
	p := s.Create("lounge", "fp1", "alice", "Best room?", []string{"lounge", "gallery"})
	s.Vote(p.ID, "fp2", 0)
	s.Vote(p.ID, "fp2", 1) // change vote
	counts := s.Get(p.ID).VoteCount()
	if counts[0] != 0 || counts[1] != 1 {
		t.Errorf("counts = %v, want [0 1]", counts)
	}
}

func TestCloseOnlyByCreator(t *testing.T) {
	s := NewStore()
	p := s.Create("lounge", "fp1", "alice", "Best room?", []string{"lounge", "gallery"})
	if s.Close(p.ID, "fp2") {
		t.Error("non-creator should not close")
	}
	if !s.Close(p.ID, "fp1") {
		t.Error("creator should close")
	}
	if !s.Get(p.ID).Closed {
		t.Error("poll should be closed")
	}
}

func TestVoteOnClosedPoll(t *testing.T) {
	s := NewStore()
	p := s.Create("lounge", "fp1", "alice", "Best room?", []string{"lounge", "gallery"})
	s.Close(p.ID, "fp1")
	if s.Vote(p.ID, "fp2", 0) {
		t.Error("should not vote on closed poll")
	}
}

func TestActiveRoomPolls(t *testing.T) {
	s := NewStore()
	s.Create("lounge", "fp1", "alice", "Poll 1", []string{"a", "b"})
	p2 := s.Create("lounge", "fp1", "alice", "Poll 2", []string{"c", "d"})
	s.Create("gallery", "fp2", "bob", "Poll 3", []string{"e", "f"})
	s.Close(p2.ID, "fp1")

	active := s.ActiveRoomPolls("lounge")
	if len(active) != 1 {
		t.Errorf("active polls = %d, want 1", len(active))
	}
}

func TestClear(t *testing.T) {
	s := NewStore()
	s.Create("lounge", "fp1", "alice", "Poll 1", []string{"a", "b"})
	s.Clear()
	if len(s.RoomPolls("lounge")) != 0 {
		t.Error("polls should be cleared")
	}
}

func TestLatestByCreator(t *testing.T) {
	s := NewStore()
	s.Create("lounge", "fp1", "alice", "Poll 1", []string{"a", "b"})
	s.Create("lounge", "fp1", "alice", "Poll 2", []string{"c", "d"})
	latest := s.LatestByCreator("lounge", "fp1")
	if latest == nil || latest.Title != "Poll 2" {
		t.Error("should return most recent poll")
	}
}
