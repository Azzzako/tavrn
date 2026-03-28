// internal/jukebox/genre.go
package jukebox

// Genre represents a music genre in the jukebox.
type Genre int

const (
	GenreLofi       Genre = iota
	GenreJazz
	GenreElectronic
	GenreCantina
)

var genreLabels = [...]string{"Lofi", "Jazz", "Electronic", "Cantina"}

func (g Genre) String() string {
	if int(g) < len(genreLabels) {
		return genreLabels[g]
	}
	return "Unknown"
}

// AllGenres returns all available genres in display order.
func AllGenres() []Genre {
	return []Genre{GenreLofi, GenreJazz, GenreElectronic, GenreCantina}
}
