package main

import (
	"embed"
	"fmt"
	"net/http"
)

//go:embed adminui/*
var staticFiles embed.FS

func main() {
	fs := http.FileServer(http.FS(staticFiles))
	http.Handle("/", fs)
	http.HandleFunc("/hello", helloHandler)
	
	err := http.ListenAndServe(":18080", nil)
	if err != nil {
		panic(err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}
