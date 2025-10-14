package main

import (
	"bytes"
	"context"
	"embed"
	"io/fs"
	"net/http"
	"time"
	
	"github.com/google/uuid"
	"github.com/yincongcyincong/MuseBot/logger"
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

func WithRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logID := uuid.New().String()
		
		isSSE := r.Header.Get("Accept") == "text/event-stream"
		
		if !isSSE {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
			defer cancel()
		}
		
		// 通用的 context 值
		ctx = context.WithValue(ctx, "log_id", logID)
		ctx = context.WithValue(ctx, "start_time", time.Now())
		r = r.WithContext(ctx)
		
		logger.InfoCtx(ctx, "request start", "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
