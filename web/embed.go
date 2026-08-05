package web

import (
	"embed"
	"io/fs"
)

//go:embed "all:dist"
var files embed.FS

func WebUI() (fs.FS, error) {
	return fs.Sub(files, "dist")
}
