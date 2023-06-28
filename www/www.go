package www

import (
	"embed"
	"io/fs"
	"log"
)

//go:generate ../scripts/package_react.sh
//go:embed dist
var efs embed.FS

func FS() fs.FS {
	fsys, err := fs.Sub(efs, "dist")
	if err != nil {
		log.Fatal(err)
	}
	return fsys
}
