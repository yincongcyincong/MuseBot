package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/param"
)

// fakeProvider is a stand-in OpenAI-compatible endpoint recording requests.
type fakeProvider struct {
	server    *httptest.Server
	hits      int
	status    int
	lastModel string
}

func newFakeProvider(t *testing.T, status int) *fakeProvider {
	t.Helper()
	fp := &fakeProvider{status: status}
	fp.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp.hits++
		var req openai.ChatCompletionRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		fp.lastModel = req.Model
		w.WriteHeader(fp.status)
		if fp.status == http.StatusOK {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{{"message": map[string]string{"content": "ok"}}},
				"usage":   map[string]int{"total_tokens": 1},
			})
		}
	}))
	t.Cleanup(fp.server.Close)
	return fp
}

// routeProvider points a provider type at a fake server for the test.
func routeProvider(t *testing.T, provider string, fp *fakeProvider) {
	t.Helper()
	prev, hadPrev := openaiEndpoints[provider]
	openaiEndpoints[provider] = struct {
		token   func() string
		baseURL string
	}{token: func() string { return "test-token" }, baseURL: fp.server.URL}
	t.Cleanup(func() {
		if hadPrev {
			openaiEndpoints[provider] = prev
		} else {
			delete(openaiEndpoints, provider)
		}
	})
}

// patchConf saves and restores the conf fields these tests touch.
func patchConf(t *testing.T, failover bool, chain string, retryTimes int) {
	t.Helper()
	prevFailover, prevChain := conf.BaseConfInfo.EnableFailover, conf.BaseConfInfo.FailoverProviders
	prevRetry, prevInterval := conf.BaseConfInfo.LLMRetryTimes, conf.BaseConfInfo.LLMRetryInterval
	prevStream, prevCustom := conf.BaseConfInfo.IsStreaming, conf.BaseConfInfo.CustomUrl
	conf.BaseConfInfo.EnableFailover = failover
	conf.BaseConfInfo.FailoverProviders = chain
	conf.BaseConfInfo.LLMRetryTimes = retryTimes
	conf.BaseConfInfo.LLMRetryInterval = 0
	conf.BaseConfInfo.IsStreaming = false
	conf.BaseConfInfo.CustomUrl = ""
	t.Cleanup(func() {
		conf.BaseConfInfo.EnableFailover = prevFailover
		conf.BaseConfInfo.FailoverProviders = prevChain
		conf.BaseConfInfo.LLMRetryTimes = prevRetry
		conf.BaseConfInfo.LLMRetryInterval = prevInterval
		conf.BaseConfInfo.IsStreaming = prevStream
		conf.BaseConfInfo.CustomUrl = prevCustom
	})
}

// newChatLLM builds an LLM whose user config resolves to the openai type.
func newChatLLM(t *testing.T) *LLM {
	t.Helper()
	ctx := context.WithValue(context.Background(), "user_info",
		&db.User{LLMConfigRaw: &param.LLMConfig{TxtType: param.OpenAi, TxtModel: "primary-m"}})
	l := NewLLM(WithUserId("u1"), WithContext(ctx))
	l.MessageChan = make(chan *param.MsgInfo, 1)
	l.LLMClient.GetMessage(openai.ChatMessageRoleUser, "hi")
	l.LLMClient.GetModel(l)
	return l
}

func TestFailoverOnServerError(t *testing.T) {
	patchConf(t, true, param.Gemini, 1)
	primary := newFakeProvider(t, http.StatusInternalServerError)
	backup := newFakeProvider(t, http.StatusOK)
	routeProvider(t, param.OpenAi, primary)
	routeProvider(t, param.Gemini, backup)

	l := newChatLLM(t)
	assert.NoError(t, l.sendWithFailover())

	assert.Equal(t, 1, primary.hits)
	assert.Equal(t, 1, backup.hits)
	assert.Equal(t, "ok", l.WholeContent)
	// the fallback rung runs on the provider's default model
	assert.Equal(t, param.ModelGemini25Flash, backup.lastModel)
	assert.Equal(t, param.Gemini, l.providerOverride)
}

func TestFailoverSkipsRetriesOnAuthError(t *testing.T) {
	// 401 cannot be retried into success: the outer loop must stop early.
	// With LLMRetryTimes=2 the primary sees 2 hits (inner retry only), not 4.
	patchConf(t, true, param.Gemini, 2)
	primary := newFakeProvider(t, http.StatusUnauthorized)
	backup := newFakeProvider(t, http.StatusOK)
	routeProvider(t, param.OpenAi, primary)
	routeProvider(t, param.Gemini, backup)

	l := newChatLLM(t)
	assert.NoError(t, l.sendWithFailover())

	assert.Equal(t, 2, primary.hits)
	assert.Equal(t, 1, backup.hits)
}

func TestFailoverDisabledByConfig(t *testing.T) {
	patchConf(t, false, param.Gemini, 1)
	primary := newFakeProvider(t, http.StatusInternalServerError)
	backup := newFakeProvider(t, http.StatusOK)
	routeProvider(t, param.OpenAi, primary)
	routeProvider(t, param.Gemini, backup)

	l := newChatLLM(t)
	assert.Error(t, l.sendWithFailover())
	assert.Zero(t, backup.hits)
}

func TestFailoverAbortsOnCtxCancel(t *testing.T) {
	patchConf(t, true, param.Gemini, 1)
	primary := newFakeProvider(t, http.StatusInternalServerError)
	backup := newFakeProvider(t, http.StatusOK)
	routeProvider(t, param.OpenAi, primary)
	routeProvider(t, param.Gemini, backup)

	l := newChatLLM(t)
	ctx, cancel := context.WithCancel(l.Ctx)
	l.Ctx = ctx
	cancel()

	assert.ErrorIs(t, l.sendWithFailover(), context.Canceled)
	assert.Zero(t, backup.hits)
}

func TestFailoverNoChainAvailable(t *testing.T) {
	// explicit chain is empty and no other provider has a token
	patchConf(t, true, "", 1)
	primary := newFakeProvider(t, http.StatusInternalServerError)
	routeProvider(t, param.OpenAi, primary)

	prevTokens := map[string]*string{
		"deepseek":     &conf.BaseConfInfo.DeepseekToken,
		"gemini":       &conf.BaseConfInfo.GeminiToken,
		"openai":       &conf.BaseConfInfo.OpenAIToken,
		"aliyun":       &conf.BaseConfInfo.AliyunToken,
		"vol":          &conf.BaseConfInfo.VolToken,
		"chatanywhere": &conf.BaseConfInfo.ChatAnyWhereToken,
	}
	prevVals := map[string]string{}
	for name, p := range prevTokens {
		prevVals[name] = *p
		*p = ""
	}
	t.Cleanup(func() {
		for name, p := range prevTokens {
			*p = prevVals[name]
		}
	})

	l := newChatLLM(t)
	assert.Error(t, l.sendWithFailover())
}

func TestCanFailoverGuards(t *testing.T) {
	patchConf(t, true, "", 1)
	l := newChatLLM(t)
	assert.True(t, l.canFailover())

	conf.BaseConfInfo.EnableFailover = false
	assert.False(t, l.canFailover())
	conf.BaseConfInfo.EnableFailover = true

	// utility requests pin their provider, no failover
	l.providerOverride = param.Gemini
	assert.False(t, l.canFailover())
	l.providerOverride = ""

	// custom protocol clients cannot migrate messages mid-turn
	l.LLMClient = &OllamaReq{}
	assert.False(t, l.canFailover())
	l.LLMClient = &OpenAIReq{}

	// mid-stream output was already delivered, restarting would duplicate it
	conf.BaseConfInfo.IsStreaming = true
	l.WholeContent = "partial"
	assert.False(t, l.canFailover())
}

func TestUtilityModelRouting(t *testing.T) {
	patchConf(t, false, "", 1)
	primary := newFakeProvider(t, http.StatusOK)
	utility := newFakeProvider(t, http.StatusOK)
	routeProvider(t, param.OpenAi, primary)
	routeProvider(t, param.Gemini, utility)

	prevType, prevModel := conf.BaseConfInfo.UtilityType, conf.BaseConfInfo.UtilityModel
	conf.BaseConfInfo.UtilityType = param.Gemini
	conf.BaseConfInfo.UtilityModel = "cheap-m"
	t.Cleanup(func() {
		conf.BaseConfInfo.UtilityType = prevType
		conf.BaseConfInfo.UtilityModel = prevModel
	})

	l := newChatLLM(t)
	l.providerOverride, l.modelOverride = "", ""
	utilityLLM := NewLLM(WithUtilityModel(), WithUserId(l.UserId), WithContext(l.Ctx))
	assert.Equal(t, param.Gemini, utilityLLM.providerOverride)
	utilityLLM.LLMClient.GetModel(utilityLLM)
	assert.Equal(t, "cheap-m", utilityLLM.Model)

	content, err := utilityLLM.LLMClient.SyncSend(context.Background(), utilityLLM)
	assert.NoError(t, err)
	assert.Equal(t, "ok", content)
	assert.Equal(t, 1, utility.hits)
	assert.Equal(t, "cheap-m", utility.lastModel)
	assert.Zero(t, primary.hits)
}

func TestUtilityModelNoopWhenUnconfigured(t *testing.T) {
	prevType, prevModel := conf.BaseConfInfo.UtilityType, conf.BaseConfInfo.UtilityModel
	conf.BaseConfInfo.UtilityType = ""
	conf.BaseConfInfo.UtilityModel = ""
	t.Cleanup(func() {
		conf.BaseConfInfo.UtilityType = prevType
		conf.BaseConfInfo.UtilityModel = prevModel
	})

	ctx := context.WithValue(context.Background(), "user_info",
		&db.User{LLMConfigRaw: &param.LLMConfig{TxtType: param.OpenAi}})
	l := NewLLM(WithUtilityModel(), WithContext(ctx))
	assert.Empty(t, l.providerOverride)
	assert.Empty(t, l.modelOverride)
}
