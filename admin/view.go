package main

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"time"
)

//go:embed adminui/*
var staticFiles embed.FS

func View() http.HandlerFunc {
	distFS, _ := fs.Sub(staticFiles, "adminui")
	
	staticHandler := http.FileServer(http.FS(distFS))
	
	return func(w http.ResponseWriter, r *http.Request) {
		if fileExists(distFS, r.URL.Path[1:]) {
			staticHandler.ServeHTTP(w, r)
			return
		}
		
		fileBytes, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		
		reader := bytes.NewReader(fileBytes)
		http.ServeContent(w, r, "index.html", time.Now(), reader)
	}
}

func fileExists(fsys fs.FS, path string) bool {
	f, err := fsys.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		return false
	}
	return true
}
