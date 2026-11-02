package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestMain(m *testing.M) {
	conf.InitConf()
	db.InitTable()
	os.Exit(m.Run())
}

// testWebhookRequest builds a request carrying the context values the real
// WithRequestContext middleware injects (utils.Success/Failure read them).
func testWebhookRequest(method, url string, body []byte) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, url, bytes.NewReader(body))
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), "log_id", "webhook-test")
	ctx = context.WithValue(ctx, "start_time", time.Now())
	return r.WithContext(ctx)
}

func decodeWebhookResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	res := make(map[string]any)
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode response fail: %v, body: %s", err, w.Body.String())
	}
	return res
}

func TestGenerateWebhookToken(t *testing.T) {
	token, tokenHash, err := generateWebhookToken()
	if err != nil {
		t.Fatalf("generate token fail: %v", err)
	}
	if len(token) != webhookTokenLen || !isHexString(token) {
		t.Fatalf("expected %d hex chars, got %q", webhookTokenLen, token)
	}
	if tokenHash != sha256Hex(token) {
		t.Fatal("token hash must equal sha256 of the token")
	}

	// two generations never collide
	token2, _, _ := generateWebhookToken()
	if token == token2 {
		t.Fatal("tokens must be random per call")
	}
}

func TestRenderWebhookPrompt(t *testing.T) {
	payload := map[string]any{
		"msg":    "deploy ok",
		"nested": map[string]any{"env": "prod"},
	}

	out, err := renderWebhookPrompt("event: {{.msg}} env={{.nested.env}}", payload)
	if err != nil || out != "event: deploy ok env=prod" {
		t.Fatalf("render = %q, err = %v", out, err)
	}

	// missing keys render as <no value>, they do not fail
	out, err = renderWebhookPrompt("{{.missing}}", payload)
	if err != nil || out != "<no value>" {
		t.Fatalf("missing key render = %q, err = %v", out, err)
	}

	// oversized values are truncated to the prompt cap
	big := map[string]any{"msg": strings.Repeat("x", webhookPromptMaxBytes*3)}
	out, err = renderWebhookPrompt("{{.msg}}", big)
	if err != nil {
		t.Fatalf("long render fail: %v", err)
	}
	if len(out) > webhookPromptMaxBytes+8 { // rune boundary may add a few bytes
		t.Fatalf("expected truncation to ~%d bytes, got %d", webhookPromptMaxBytes, len(out))
	}

	// invalid templates are rejected
	if _, err := renderWebhookPrompt("{{.unclosed", payload); err == nil {
		t.Fatal("expected invalid template to fail")
	}
}

// insertTestWebhook creates a webhook with a known token and returns its id.
func insertTestWebhook(t *testing.T, name, token, tmpl string, status int) int64 {
	t.Helper()
	id, err := db.InsertWebhook(name, sha256Hex(token), "generic", param.Wechat,
		"wx_target", "", "", tmpl, "tester")
	if err != nil {
		t.Fatalf("insert webhook fail: %v", err)
	}
	if status == 0 {
		if err := db.UpdateWebhookStatus(id, 0); err != nil {
			t.Fatalf("disable webhook fail: %v", err)
		}
	}
	return id
}

func TestTriggerWebhook(t *testing.T) {
	token := strings.Repeat("ab", 32) // valid 64-char hex
	id := insertTestWebhook(t, "trigger_test", token, "event: {{.msg}}", 1)
	defer func() { _ = db.DeleteWebhookByID(id) }()

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/", TriggerWebhook)

	post := func(url string, body []byte, contentType string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		r.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}

	// valid trigger is accepted
	payload, _ := json.Marshal(map[string]string{"msg": "build passed"})
	w := post("/webhook/"+token, payload, "application/json")
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d, body: %s", w.Code, w.Body.String())
	}

	// wrong token length/format is a plain 404
	if w := post("/webhook/short", payload, "application/json"); w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for short token, got %d", w.Code)
	}
	if w := post("/webhook/"+strings.Repeat("zz", 32), payload, "application/json"); w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-hex token, got %d", w.Code)
	}

	// unknown token is a plain 404, indistinguishable from a missing webhook
	if w := post("/webhook/"+strings.Repeat("cd", 32), payload, "application/json"); w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown token, got %d", w.Code)
	}

	// non-JSON payload is rejected so templates never see garbage
	if w := post("/webhook/"+token, []byte("not json"), "text/plain"); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-json payload, got %d", w.Code)
	}

	// oversized payload is rejected before parsing
	huge := bytes.Repeat([]byte(`{"pad":"`), 1)
	huge = append(huge, bytes.Repeat([]byte("a"), 2<<20)...)
	huge = append(huge, []byte(`"}`)...)
	if w := post("/webhook/"+token, huge, "application/json"); w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversized payload, got %d", w.Code)
	}

	// non-POST methods are rejected
	r := httptest.NewRequest(http.MethodGet, "/webhook/"+token, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET, got %d", w.Code)
	}

	// a disabled webhook is invisible to the trigger endpoint
	if err := db.UpdateWebhookStatus(id, 0); err != nil {
		t.Fatalf("disable webhook fail: %v", err)
	}
	if w := post("/webhook/"+token, payload, "application/json"); w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for disabled webhook, got %d", w.Code)
	}
	_ = db.UpdateWebhookStatus(id, 1)
}

func TestTriggerWebhookDisabled(t *testing.T) {
	if !conf.BaseConfInfo.EnableWebhook {
		t.Skip("webhook already disabled")
	}
	conf.BaseConfInfo.EnableWebhook = false
	defer func() { conf.BaseConfInfo.EnableWebhook = true }()

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/", TriggerWebhook)
	r := httptest.NewRequest(http.MethodPost, "/webhook/"+strings.Repeat("ab", 32), nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when webhooks are disabled, got %d", w.Code)
	}
}

func TestWebhookCRUDHandlers(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/create", CreateWebhook)
	mux.HandleFunc("/webhook/update", UpdateWebhook)
	mux.HandleFunc("/webhook/update_status", UpdateWebhookStatus)
	mux.HandleFunc("/webhook/delete", DeleteWebhook)
	mux.HandleFunc("/webhook/list", GetWebhooks)

	// invalid template rejected at create
	badReq := testWebhookRequest(http.MethodPost, "/webhook/create",
		[]byte(`{"webhook_name":"bad","type":"wechat","target_id":"t","prompt_template":"{{.broken"}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, badReq)
	res := decodeWebhookResponse(t, w)
	if res["code"].(float64) == float64(param.CodeSuccess) {
		t.Fatalf("expected failure for invalid template, got %+v", res)
	}

	// valid create returns the token exactly once
	req := testWebhookRequest(http.MethodPost, "/webhook/create",
		[]byte(`{"webhook_name":"crud_test","type":"wechat","target_id":"wx_1","prompt_template":"hi {{.msg}}"}`))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	res = decodeWebhookResponse(t, w)
	if res["code"].(float64) != float64(param.CodeSuccess) {
		t.Fatalf("create failed: %+v", res)
	}
	data := res["data"].(map[string]any)
	token, _ := data["token"].(string)
	urlPath, _ := data["url_path"].(string)
	webhookID, _ := data["id"].(string)
	if len(token) != webhookTokenLen || !strings.HasSuffix(urlPath, token) || webhookID == "" {
		t.Fatalf("unexpected create payload: %+v", data)
	}
	cleanup := func() {
		r := testWebhookRequest(http.MethodGet, "/webhook/delete?id="+webhookID, nil)
		mux.ServeHTTP(httptest.NewRecorder(), r)
	}
	defer cleanup()

	// list shows the webhook but never the token or its hash
	r := testWebhookRequest(http.MethodGet, "/webhook/list?page=1&page_size=10&name=crud_test", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	body := w.Body.String()
	if !strings.Contains(body, "crud_test") {
		t.Fatalf("expected webhook in list, body: %s", body)
	}
	if strings.Contains(body, token) || strings.Contains(body, sha256Hex(token)) {
		t.Fatal("list response leaked the token or its hash")
	}

	// update keeps the token immutable by design; template revalidated
	r = testWebhookRequest(http.MethodPost, "/webhook/update",
		[]byte(fmt.Sprintf(`{"id":%s,"webhook_name":"crud_test","type":"wechat","target_id":"wx_2","prompt_template":"ok {{.msg}}"}`, webhookID)))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	res = decodeWebhookResponse(t, w)
	if res["code"].(float64) != float64(param.CodeSuccess) {
		t.Fatalf("update failed: %+v", res)
	}
	wh, err := db.GetWebhookByTokenHash(sha256Hex(token))
	if err != nil || wh.TargetID != "wx_2" || wh.PromptTemplate != "ok {{.msg}}" {
		t.Fatalf("update not persisted: %+v err=%v", wh, err)
	}

	// disable makes the trigger 404
	r = testWebhookRequest(http.MethodPost, "/webhook/update_status", []byte(`{"id":`+webhookID+`,"status":0}`))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	res = decodeWebhookResponse(t, w)
	if res["code"].(float64) != float64(param.CodeSuccess) {
		t.Fatalf("update status failed: %+v", res)
	}
	if _, err := db.GetWebhookByTokenHash(sha256Hex(token)); err == nil {
		t.Fatal("disabled webhook must not be reachable by token")
	}

	// delete removes it from the list
	r = testWebhookRequest(http.MethodGet, "/webhook/delete?id="+webhookID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	res = decodeWebhookResponse(t, w)
	if res["code"].(float64) != float64(param.CodeSuccess) {
		t.Fatalf("delete failed: %+v", res)
	}
	r = testWebhookRequest(http.MethodGet, "/webhook/list?page=1&page_size=100&name=crud_test", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if strings.Contains(w.Body.String(), "crud_test") {
		t.Fatal("deleted webhook still listed")
	}
}
