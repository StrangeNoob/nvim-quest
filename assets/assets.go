// Package assets embeds the game's lesson content.
package assets

import "embed"

//go:embed lessons/*.json
var Lessons embed.FS
