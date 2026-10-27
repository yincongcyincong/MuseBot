package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/param"
)

func ctxWithUser(llmConf *param.LLMConfig) context.Context {
	return context.WithValue(context.Background(), "user_info", &db.User{LLMConfigRaw: llmConf})
}

func TestBuiltinRecognizersRegistered(t *testing.T) {
	for _, name := range []string{param.Vol, param.OpenAi, param.Gemini, param.Aliyun} {
		assert.NotNil(t, recognizers[name], "missing recognizer for %s", name)
	}
}

func TestBuiltinChatClientRegistered(t *testing.T) {
	assert.NotNil(t, chatClients[param.Ollama], "missing chat client for ollama")
}

func TestRecognizeAudioDispatchesByMediaType(t *testing.T) {
	RegisterRecognizer("test-rec", func(ctx context.Context, audioContent []byte) (string, int, error) {
		return "ok", 7, nil
	})

	previous := conf.BaseConfInfo.MediaType
	conf.BaseConfInfo.MediaType = "test-rec"
	t.Cleanup(func() { conf.BaseConfInfo.MediaType = previous })

	// "test-rec" is not in GetAvailRecType, so GetRecType falls back to MediaType
	text, token, err := RecognizeAudio(ctxWithUser(&param.LLMConfig{RecType: "test-rec"}), []byte("audio"))
	assert.NoError(t, err)
	assert.Equal(t, "ok", text)
	assert.Equal(t, 7, token)
}

func TestRecognizeAudioUnknownType(t *testing.T) {
	previous := conf.BaseConfInfo.MediaType
	conf.BaseConfInfo.MediaType = "no-such-recognizer"
	t.Cleanup(func() { conf.BaseConfInfo.MediaType = previous })

	text, token, err := RecognizeAudio(ctxWithUser(nil), []byte("audio"))
	assert.Error(t, err)
	assert.Empty(t, text)
	assert.Equal(t, param.AudioTokenUsage, token)
}

func TestNewChatClientFallsBackToOpenAI(t *testing.T) {
	// zero out token fields so GetAvailTxtType derives nothing from them
	tokens := map[string]*string{
		"deepseek":     &conf.BaseConfInfo.DeepseekToken,
		"gemini":       &conf.BaseConfInfo.GeminiToken,
		"openai":       &conf.BaseConfInfo.OpenAIToken,
		"aliyun":       &conf.BaseConfInfo.AliyunToken,
		"vol":          &conf.BaseConfInfo.VolToken,
		"chatanywhere": &conf.BaseConfInfo.ChatAnyWhereToken,
	}
	previous := make(map[string]string, len(tokens))
	for name, p := range tokens {
		previous[name] = *p
		*p = ""
	}
	t.Cleanup(func() {
		for name, p := range tokens {
			*p = previous[name]
		}
	})

	previousType := conf.BaseConfInfo.Type
	t.Cleanup(func() { conf.BaseConfInfo.Type = previousType })

	conf.BaseConfInfo.Type = param.Ollama
	assert.IsType(t, &OllamaReq{}, NewChatClient(ctxWithUser(nil)))

	conf.BaseConfInfo.Type = "unknown-type"
	assert.IsType(t, &OpenAIReq{}, NewChatClient(ctxWithUser(nil)))
}
