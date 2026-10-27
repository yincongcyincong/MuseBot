package utils

import (
	"bytes"
	"fmt"
	"image/png"

	"golang.org/x/image/webp"
)

// WebPToPNG decodes a WebP image and re-encodes it as PNG.
//
// Vision models generally accept image/png and image/jpeg over data URLs but
// not image/webp, and Telegram stickers are WebP. Converting keeps the rest of
// the pipeline (which already handles photos) unchanged. Animated WebP decodes
// to its first frame.
func WebPToPNG(content []byte) ([]byte, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("empty webp content")
	}

	img, err := webp.Decode(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
