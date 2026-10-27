package llm

import (
	"context"
	"fmt"

	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

// Recognizer turns audio bytes into text and reports its token cost.
type Recognizer func(ctx context.Context, audioContent []byte) (string, int, error)

var recognizers = map[string]Recognizer{}

// RegisterRecognizer wires a recognizer under a provider type (param.Vol,
// param.OpenAi, ...). Providers self-register from their own file's init,
// so adding one means adding a file instead of touching the dispatch sites.
func RegisterRecognizer(name string, r Recognizer) {
	if name == "" {
		panic("llm: RegisterRecognizer called with empty name")
	}
	if r == nil {
		panic("llm: RegisterRecognizer called with nil recognizer: " + name)
	}
	recognizers[name] = r
}

// RecognizeAudio dispatches to the recognizer selected by the caller's rec type.
func RecognizeAudio(ctx context.Context, audioContent []byte) (string, int, error) {
	t := utils.GetRecType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	r, ok := recognizers[t]
	if !ok {
		logger.WarnCtx(ctx, "no recognizer registered for type", "type", t)
		return "", param.AudioTokenUsage, fmt.Errorf("no recognizer registered for type %q", t)
	}
	return r(ctx, audioContent)
}

// chatClients maps a txt provider type to its chat protocol client;
// types without an entry fall back to the OpenAI-compatible client.
var chatClients = map[string]func() LLMClient{}

// RegisterChatClient wires a chat protocol client under a provider type.
func RegisterChatClient(name string, newClient func() LLMClient) {
	if name == "" {
		panic("llm: RegisterChatClient called with empty name")
	}
	if newClient == nil {
		panic("llm: RegisterChatClient called with nil constructor: " + name)
	}
	chatClients[name] = newClient
}

// NewChatClient returns the chat client for the caller's txt type,
// defaulting to the OpenAI-compatible one.
func NewChatClient(ctx context.Context) LLMClient {
	if newClient, ok := chatClients[utils.GetTxtType(db.GetCtxUserInfo(ctx).LLMConfigRaw)]; ok {
		return newClient()
	}
	return &OpenAIReq{}
}
