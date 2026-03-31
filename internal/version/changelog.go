package version

// Entry holds a single version's changelog.
type Entry struct {
	Version string
	Changes []string
}

// Changelog lists versions newest-first.
var Changelog = []Entry{
	{
		Version: "0.2",
		Changes: []string{
			"polls — /poll, /vote, /endpoll",
			"tankard clicker in the sidebar (F6)",
			"drink count survives the weekly purge",
			"@mentions with F4 popup + room badges",
			"all-time visitor count in top bar",
			"ban/unban by nickname from admin CLI",
		},
	},
	{
		Version: "0.1",
		Changes: []string{
			"SSH tavern with 4 rooms",
			"gallery — sticky notes you can drag around",
			"co-op sudoku in #games",
			"animated splash with floating sparks",
			"weekly purge every Sunday 23:59 UTC",
			"server banner, nick colors, flair",
		},
	},
}
