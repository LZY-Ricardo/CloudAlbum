package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/dist
var webDist embed.FS

func WebFS() http.FileSystem {
	sub, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}
