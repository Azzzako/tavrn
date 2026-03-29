package catalogs

import "embed"

//go:embed catalog.json *.txt
var FS embed.FS
