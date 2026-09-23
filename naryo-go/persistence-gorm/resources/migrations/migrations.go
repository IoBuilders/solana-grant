// Package migrations embeds this directory's Atlas migration files (and their checksum
// manifest) so persistencegormsetup.Setup can apply them at server startup via atlasexec,
// without needing this directory to exist separately on disk in the runtime image.
package migrations

import "embed"

//go:embed *.sql atlas.sum
var FS embed.FS
