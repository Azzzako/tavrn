package identity

import (
	"crypto/sha256"
)

// DefaultNickname returns a tavern-themed name from the fingerprint.
func DefaultNickname(fingerprint string) string {
	hash := sha256.Sum256([]byte(fingerprint))
	idx := int(hash[0])

	adjectives := []string{
		"dusty", "quiet", "wandering", "sleepy", "hooded",
		"scarred", "weary", "shadowy", "lone", "grizzled",
		"nimble", "silent", "copper", "ashen", "veiled",
		"stout", "swift", "rusty", "hollow", "mossy",
	}
	nouns := []string{
		"pilgrim", "drifter", "bard", "ranger", "rogue",
		"sage", "merchant", "smith", "herald", "scout",
		"monk", "keeper", "hunter", "scribe", "warden",
		"brewer", "hermit", "jester", "rider", "ghost",
	}

	adj := adjectives[idx%len(adjectives)]
	noun := nouns[int(hash[1])%len(nouns)]
	return adj + "_" + noun
}

// ColorIndex returns 0-11 for mapping a fingerprint to one of 12 cantina colors.
func ColorIndex(fingerprint string) int {
	hash := sha256.Sum256([]byte(fingerprint))
	return int(hash[0]) % 12
}

// HasFlair returns true if the user has connected 3+ times this week.
func HasFlair(visitCount int) bool {
	return visitCount >= 3
}

// IsOwner returns true if the nickname is the tavern owner.
func IsOwner(nickname string) bool {
	return nickname == "neur0map"
}

// OwnerDisplayName returns the special display name for the owner.
func OwnerDisplayName() string {
	return "★ neur0map"
}
