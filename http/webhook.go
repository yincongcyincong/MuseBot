package http

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/robot"
	"github.com/yincongcyincong/MuseBot/utils"
)

const (
	// webhookPrefix is the public trigger path: POST /webhook/{token}
	webhookPrefix = "/webhook/"
	// webhookTokenLen is the hex length of the 32-byte random token
	webhookTokenLen = 64
	// webhookPromptMaxBytes caps the rendered prompt so a huge payload cannot
	// be poured straight into the LLM context
	webhookPromptMaxBytes = 8192
)

// webhookSources lists the payload adapters; P1 ships the generic JSON one.
// Adding github/alertmanager later means adding a normalizer here plus the
// source name validation below.
var webhookSources = map[string]bool{
	"generic": true,
}

type WebhookRequest struct {
	ID             int64  `json:"id"`
	WebhookName    string `json:"webhook_name"`
	Source         string `json:"source"`
	Type           string `json:"type"`
	TargetID       string `json:"target_id"`
	GroupID        string `json:"group_id"`
	Command        string `json:"command"`
	PromptTemplate string `json:"prompt_template"`
	CreateBy       string `json:"create_by"`
}

// generateWebhookToken returns a fresh 32-byte hex token plus its SHA-256
// hex hash; only the hash is persisted, the token is shown once at creation.
func generateWebhookToken() (token, tokenHash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate webhook token fail: %w", err)
	}
	token = hex.EncodeToString(buf)
	return token, sha256Hex(token), nil
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func isHexString(s string) bool {
	if s == "" {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// renderWebhookPrompt runs the stored text/template over the parsed JSON
// payload and truncates the result to a sane prompt size.
func renderWebhookPrompt(tmpl string, payload map[string]any) (string, error) {
	t, err := template.New("webhook").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse template fail: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, payload); err != nil {
		return "", fmt.Errorf("render template fail: %w", err)
	}
	rendered := strings.TrimSpace(buf.String())
	// rune-safe truncation: cut then back off to a valid boundary
	if len(rendered) > webhookPromptMaxBytes {
		rendered = string([]rune(rendered)[:utf8.RuneCountInString(rendered[:webhookPromptMaxBytes])])
	}
	return rendered, nil
}

// writeWebhookStatus answers the trigger endpoint with a real HTTP status
// code (senders such as GitHub only retry on non-2xx), bypassing
// utils.Success/Failure which always write 200.
func writeWebhookStatus(w http.ResponseWriter, statusCode int, code int, message string) {
	resp := map[string]any{"code": code, "message": message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}

// TriggerWebhook receives POST /webhook/{token}: look the webhook up by the
// token hash, render its prompt template over the JSON payload, acknowledge
// immediately, and execute asynchronously.
func TriggerWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		writeWebhookStatus(w, http.StatusMethodNotAllowed, param.CodeParamError, "method not allowed")
		return
	}
	if !conf.BaseConfInfo.EnableWebhook {
		writeWebhookStatus(w, http.StatusNotFound, param.CodeParamError, "not found")
		return
	}

	token := strings.TrimPrefix(r.URL.Path, webhookPrefix)
	// a bad token is indistinguishable from a missing webhook: plain 404
	if len(token) != webhookTokenLen || !isHexString(token) {
		writeWebhookStatus(w, http.StatusNotFound, param.CodeParamError, "not found")
		return
	}

	webhookInfo, err := db.GetWebhookByTokenHash(sha256Hex(token))
	if err != nil || webhookInfo == nil {
		writeWebhookStatus(w, http.StatusNotFound, param.CodeParamError, "not found")
		return
	}

	limitBytes := int64(conf.BaseConfInfo.WebhookPayloadMB) << 20
	if limitBytes <= 0 {
		limitBytes = 1 << 20
	}
	if r.ContentLength > limitBytes {
		writeWebhookStatus(w, http.StatusRequestEntityTooLarge, param.CodeParamError, "payload too large")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, limitBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeWebhookStatus(w, http.StatusRequestEntityTooLarge, param.CodeParamError, "payload too large")
		return
	}

	payload := make(map[string]any)
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			logger.ErrorCtx(ctx, "webhook payload is not json", "err", err, "webhook", webhookInfo.WebhookName)
			writeWebhookStatus(w, http.StatusBadRequest, param.CodeParamError, "invalid json payload")
			return
		}
	}

	prompt, err := renderWebhookPrompt(webhookInfo.PromptTemplate, payload)
	if err != nil {
		logger.ErrorCtx(ctx, "render webhook prompt fail", "err", err, "webhook", webhookInfo.WebhookName)
		writeWebhookStatus(w, http.StatusInternalServerError, param.CodeServerFail, param.MsgServerFail)
		return
	}

	go robot.ExecWebhook(ctx, webhookInfo, prompt)

	writeWebhookStatus(w, http.StatusAccepted, param.CodeSuccess, "accepted")
}

func validateWebhookRequest(req *WebhookRequest, forCreate bool) error {
	if forCreate && req.WebhookName == "" {
		return fmt.Errorf("webhook_name is required")
	}
	if req.ID == 0 && !forCreate {
		return fmt.Errorf("id is required")
	}
	if req.Type == "" {
		return fmt.Errorf("type is required")
	}
	if req.TargetID == "" && req.GroupID == "" {
		return fmt.Errorf("target_id or group_id is required")
	}
	if req.PromptTemplate == "" {
		return fmt.Errorf("prompt_template is required")
	}
	if _, err := template.New("webhook").Parse(req.PromptTemplate); err != nil {
		return fmt.Errorf("invalid prompt_template: %w", err)
	}
	source := req.Source
	if source == "" {
		source = "generic"
	}
	if !webhookSources[source] {
		return fmt.Errorf("unsupported source %q", source)
	}
	return nil
}

func CreateWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &WebhookRequest{}
	err := utils.HandleJsonBody(r, req)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}

	if err := validateWebhookRequest(req, true); err != nil {
		utils.Failure(ctx, w, r, param.CodeParamError, err.Error(), nil)
		return
	}
	source := req.Source
	if source == "" {
		source = "generic"
	}

	token, tokenHash, err := generateWebhookToken()
	if err != nil {
		logger.ErrorCtx(ctx, "generate webhook token error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}

	id, err := db.InsertWebhook(req.WebhookName, tokenHash, source, req.Type,
		req.TargetID, req.GroupID, req.Command, req.PromptTemplate, req.CreateBy)
	if err != nil {
		logger.ErrorCtx(ctx, "insert webhook error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}

	// the token is returned exactly once; the database keeps only its hash
	utils.Success(ctx, w, r, map[string]string{
		"id":       fmt.Sprintf("%d", id),
		"url_path": webhookPrefix + token,
		"token":    token,
	})
}

func UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &WebhookRequest{}
	err := utils.HandleJsonBody(r, req)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}

	if err := validateWebhookRequest(req, false); err != nil {
		utils.Failure(ctx, w, r, param.CodeParamError, err.Error(), nil)
		return
	}

	_, err = db.GetWebhookByID(req.ID)
	if err != nil {
		logger.ErrorCtx(ctx, "get webhook error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}

	source := req.Source
	if source == "" {
		source = "generic"
	}
	err = db.UpdateWebhook(req.ID, req.WebhookName, source, req.Type,
		req.TargetID, req.GroupID, req.Command, req.PromptTemplate)
	if err != nil {
		logger.ErrorCtx(ctx, "update webhook error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}

	utils.Success(ctx, w, r, nil)
}

func UpdateWebhookStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type StatusRequest struct {
		ID     int64 `json:"id"`
		Status int   `json:"status"` // 0:disable 1:enable
	}
	req := &StatusRequest{}
	err := utils.HandleJsonBody(r, req)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}

	if req.ID == 0 || (req.Status != 0 && req.Status != 1) {
		utils.Failure(ctx, w, r, param.CodeParamError, "Invalid ID or Status", nil)
		return
	}

	err = db.UpdateWebhookStatus(req.ID, req.Status)
	if err != nil {
		logger.ErrorCtx(ctx, "update webhook status error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}

	utils.Success(ctx, w, r, nil)
}

func DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}

	id := utils.ParseInt(r.FormValue("id"))
	if id == 0 {
		utils.Failure(ctx, w, r, param.CodeParamError, "ID is required for delete", nil)
		return
	}

	err = db.DeleteWebhookByID(int64(id))
	if err != nil {
		logger.ErrorCtx(ctx, "delete webhook error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}

	utils.Success(ctx, w, r, nil)
}

func GetWebhooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	page := utils.ParseInt(r.FormValue("page"))
	pageSize := utils.ParseInt(r.FormValue("page_size"))
	name := r.FormValue("name")

	webhooks, err := db.GetWebhooksByPage(page, pageSize, name)
	if err != nil {
		logger.ErrorCtx(ctx, "get webhooks error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}

	total, err := db.GetWebhooksCount(name)
	if err != nil {
		logger.ErrorCtx(ctx, "get webhooks count error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}

	utils.Success(ctx, w, r, map[string]interface{}{
		"list":  webhooks,
		"total": total,
	})
}
