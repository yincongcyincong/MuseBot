package rag

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yincongcyincong/MuseBot/conf"
)

func TestDocExt(t *testing.T) {
	cases := map[string]string{
		"notes.txt":   ".txt",
		"PAGE.HTML":   ".html",
		"archive.zip": ".zip",
		"Makefile":    "",
		"a.b.PDF":     ".pdf",
	}
	for name, want := range cases {
		if got := docExt(name); got != want {
			t.Errorf("docExt(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestFindDirEntry(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "kb.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write temp file fail: %v", err)
	}

	entry, err := findDirEntry(dir, "kb.txt")
	if err != nil {
		t.Fatalf("findDirEntry fail: %v", err)
	}
	if entry.Name() != "kb.txt" {
		t.Errorf("expected kb.txt, got %q", entry.Name())
	}

	if _, err := findDirEntry(dir, "missing.txt"); err == nil {
		t.Error("expected error for missing file")
	}
	if _, err := findDirEntry(filepath.Join(dir, "nope"), "kb.txt"); err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestBuildContextWithoutStore(t *testing.T) {
	// no vector store configured: empty context, never touches a db
	orig := conf.RagConfInfo.Store
	defer func() { conf.RagConfInfo.Store = orig }()
	conf.RagConfInfo.Store = nil

	if got := BuildContext(context.Background(), "anything"); got != "" {
		t.Errorf("expected empty context without store, got %q", got)
	}
}
