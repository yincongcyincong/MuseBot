package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/param"
	"google.golang.org/genai"
)

func TestGenerateGeminiText_EmptyAudio(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_info", &db.User{
		LLMConfig:    `{"type":"gemini"}`,
		LLMConfigRaw: &param.LLMConfig{TxtType: param.Gemini},
	})
	text, _, err := GenerateGeminiText(ctx, []byte{})
	assert.Error(t, err)
	assert.Empty(t, text)
}

func TestGenerateGeminiImage_EmptyPrompt(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_info", &db.User{
		LLMConfig:    `{"type":"gemini"}`,
		LLMConfigRaw: &param.LLMConfig{TxtType: param.Gemini},
	})
	image, _, err := GenerateGeminiImg(ctx, "", nil)
	assert.Error(t, err)
	assert.Nil(t, image)
}

func TestGenerateGeminiVideo_InvalidPrompt(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_info", &db.User{
		LLMConfig:    `{"type":"gemini"}`,
		LLMConfigRaw: &param.LLMConfig{TxtType: param.Gemini},
	})
	video, _, err := GenerateGeminiVideo(ctx, "", nil)
	assert.Error(t, err)
	assert.Nil(t, video)
}

func TestGeminiVideoBytes(t *testing.T) {
	cases := []struct {
		name      string
		operation *genai.GenerateVideosOperation
		want      string
		wantErr   string
	}{
		{name: "nil operation", wantErr: "video operation is empty"},
		{name: "failed operation", operation: &genai.GenerateVideosOperation{Error: map[string]any{"message": "quota exceeded"}}, wantErr: "quota exceeded"},
		{name: "nil response", operation: &genai.GenerateVideosOperation{Done: true}, wantErr: "video is empty"},
		{name: "empty videos", operation: &genai.GenerateVideosOperation{Response: &genai.GenerateVideosResponse{}}, wantErr: "video is empty"},
		{name: "nil video", operation: &genai.GenerateVideosOperation{Response: &genai.GenerateVideosResponse{GeneratedVideos: []*genai.GeneratedVideo{{}}}}, wantErr: "video is empty"},
		{name: "no bytes or URI", operation: &genai.GenerateVideosOperation{Response: &genai.GenerateVideosResponse{GeneratedVideos: []*genai.GeneratedVideo{{Video: &genai.Video{}}}}}, wantErr: "video is empty"},
		{name: "embedded bytes", operation: &genai.GenerateVideosOperation{Response: &genai.GenerateVideosResponse{GeneratedVideos: []*genai.GeneratedVideo{{Video: &genai.Video{VideoBytes: []byte("mp4")}}}}}, want: "mp4"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := geminiVideoBytes(context.Background(), tc.operation, func(context.Context, *genai.Video) ([]byte, error) {
				t.Fatal("download should not be called")
				return nil, nil
			})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v, want substring %q", err, tc.wantErr)
				}
				return
			}
			if err != nil || string(data) != tc.want {
				t.Fatalf("data = %q, error = %v; want %q", data, err, tc.want)
			}
		})
	}
}

func TestGeminiVideoBytesDownloadsURI(t *testing.T) {
	operation := &genai.GenerateVideosOperation{Response: &genai.GenerateVideosResponse{
		GeneratedVideos: []*genai.GeneratedVideo{{Video: &genai.Video{URI: "files/generated-video"}}},
	}}
	called := false
	data, err := geminiVideoBytes(context.Background(), operation, func(_ context.Context, video *genai.Video) ([]byte, error) {
		called = true
		if video.URI != "files/generated-video" {
			t.Fatalf("unexpected URI %q", video.URI)
		}
		return []byte("mp4"), nil
	})
	if err != nil || string(data) != "mp4" || !called {
		t.Fatalf("data = %q, error = %v, download called = %v", data, err, called)
	}

	wantErr := errors.New("download failed")
	_, err = geminiVideoBytes(context.Background(), operation, func(context.Context, *genai.Video) ([]byte, error) {
		return nil, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapped download error", err)
	}
}

func TestWaitForGeminiVideo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := waitForGeminiVideo(ctx, &genai.GenerateVideosOperation{}, time.Hour, func(context.Context, *genai.GenerateVideosOperation) (*genai.GenerateVideosOperation, error) {
		t.Fatal("poll should not be called after cancellation")
		return nil, nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context cancellation", err)
	}

	done := &genai.GenerateVideosOperation{Done: true}
	result, err := waitForGeminiVideo(context.Background(), &genai.GenerateVideosOperation{}, time.Millisecond, func(context.Context, *genai.GenerateVideosOperation) (*genai.GenerateVideosOperation, error) {
		return done, nil
	})
	if err != nil || result != done {
		t.Fatalf("result = %v, error = %v; want completed operation", result, err)
	}

	_, err = waitForGeminiVideo(context.Background(), &genai.GenerateVideosOperation{}, time.Millisecond, func(context.Context, *genai.GenerateVideosOperation) (*genai.GenerateVideosOperation, error) {
		return nil, nil
	})
	if err == nil || err.Error() != "video operation is empty" {
		t.Fatalf("error = %v, want empty operation error", err)
	}
}
