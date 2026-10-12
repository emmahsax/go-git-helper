// Package plugins embeds the git-* wrapper scripts so that `git-helper setup`
// can install them without any network access.
package plugins

import "embed"

//go:embed git-*
var FS embed.FS
