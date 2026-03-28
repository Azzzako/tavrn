// internal/jukebox/genre_test.go
package jukebox

import "testing"

func TestGenreString(t *testing.T) {
	tests := []struct {
		g    Genre
		want string
	}{
		{GenreLofi, "Lofi"},
		{GenreJazz, "Jazz"},
		{GenreElectronic, "Electronic"},
		{GenreCantina, "Cantina"},
	}
	for _, tt := range tests {
		if got := tt.g.String(); got != tt.want {
			t.Errorf("Genre(%d).String() = %q, want %q", tt.g, got, tt.want)
		}
	}
}

func TestAllGenres(t *testing.T) {
	genres := AllGenres()
	if len(genres) != 4 {
		t.Errorf("expected 4 genres, got %d", len(genres))
	}
	if genres[0] != GenreLofi {
		t.Errorf("first genre should be Lofi, got %s", genres[0])
	}
}
