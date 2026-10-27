package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"testing"
)

// a 32x32 lossless WebP image, the same format Telegram uses for static stickers
const testStickerWebP = "UklGRhwAAABXRUJQVlA4TA8AAAAvH8AHAAcQ/Y/+ByKi/wEA"

func TestWebPToPNG(t *testing.T) {
	raw, err := base64.StdEncoding.DecodeString(testStickerWebP)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	converted, err := WebPToPNG(raw)
	if err != nil {
		t.Fatalf("WebPToPNG() error = %v", err)
	}

	img, err := png.Decode(bytes.NewReader(converted))
	if err != nil {
		t.Fatalf("result is not a valid png: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 32 || b.Dy() != 32 {
		t.Fatalf("converted image is %dx%d, want 32x32", b.Dx(), b.Dy())
	}
}

func TestWebPToPNGRejectsEmpty(t *testing.T) {
	if _, err := WebPToPNG(nil); err == nil {
		t.Fatal("expected an error for empty content")
	}
}

func TestWebPToPNGRejectsNonWebP(t *testing.T) {
	if _, err := WebPToPNG([]byte("this is not a webp file")); err == nil {
		t.Fatal("expected an error for non-webp content")
	}
}

func TestWebPToPNGSize(t *testing.T) {
	raw, err := base64.StdEncoding.DecodeString(testStickerWebP)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	converted, err := WebPToPNG(raw)
	if err != nil {
		t.Fatalf("WebPToPNG() error = %v", err)
	}
	if len(converted) == 0 {
		t.Fatal(fmt.Sprintf("converted png is empty (input %d bytes)", len(raw)))
	}
}
