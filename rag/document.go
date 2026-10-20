package rag

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/langchaingo/documentloaders"
)

// MaxDocTextLen caps the extracted text injected into the LLM prompt.
const MaxDocTextLen = 50 * 1024

const docTruncateNotice = "\n\n...(document truncated)...\n\n"

func docExt(fileName string) string {
	return strings.ToLower(filepath.Ext(fileName))
}

// ParseDocument extracts the plain text of a chat document for LLM context.
// It mirrors the loader choice of the rag knowledge base but skips chunking:
// the whole text is returned, truncated to MaxDocTextLen. Unknown extensions
// (e.g. .bat/.diff/.log) are treated as plain text when the content decodes
// as text; binary content is rejected.
func ParseDocument(fileName string, r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read document fail: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var loader documentloaders.Loader
	switch docExt(fileName) {
	case ".txt":
		loader = documentloaders.NewText(bytes.NewReader(data))
	case ".pdf":
		loader = documentloaders.NewPDF(bytes.NewReader(data), int64(len(data)))
	case ".csv":
		loader = documentloaders.NewCSV(bytes.NewReader(data))
	case ".html", ".htm":
		loader = documentloaders.NewHTML(bytes.NewReader(data))
	default:
		if !utils.LooksLikeText(data) {
			return "", fmt.Errorf("unsupported document type: %s", fileName)
		}
		loader = documentloaders.NewText(bytes.NewReader(data))
	}

	docs, err := loader.Load(ctx)
	if err != nil {
		return "", fmt.Errorf("load document fail: %w", err)
	}

	parts := make([]string, 0, len(docs))
	for _, doc := range docs {
		if doc.PageContent != "" {
			parts = append(parts, doc.PageContent)
		}
	}

	return TruncateDocText(strings.Join(parts, "\n\n")), nil
}

// TruncateDocText keeps the head and tail of an over-long document so both
// the opening and the conclusion stay visible to the LLM.
func TruncateDocText(s string) string {
	if len(s) <= MaxDocTextLen {
		return s
	}

	headLen := MaxDocTextLen * 4 / 5
	tailLen := MaxDocTextLen - headLen - len(docTruncateNotice)
	return s[:headLen] + docTruncateNotice + s[len(s)-tailLen:]
}
