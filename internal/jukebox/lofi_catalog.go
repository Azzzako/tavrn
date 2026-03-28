package jukebox

import (
	_ "embed"
	"strings"
)

//go:embed chillhop_catalog.txt
var chillhopCatalog string

// parseLofiCatalog parses the embedded chillhop catalog.
// Format: one track per line, "ID!Title".
func parseLofiCatalog() []lofiTrack {
	var tracks []lofiTrack
	for _, line := range strings.Split(chillhopCatalog, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "!", 2)
		if len(parts) != 2 {
			continue
		}
		tracks = append(tracks, lofiTrack{
			id:    parts[0],
			title: parts[1],
		})
	}
	return tracks
}
