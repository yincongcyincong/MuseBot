package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"regexp"
	"strings"
	"time"

	godeepseek "github.com/cohesion-org/deepseek-go"
	"github.com/sashabaranov/go-openai"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"google.golang.org/genai"
)

const (
	OneMsgLen       = 3896
	FirstSendLen    = 30
	NonFirstSendLen = 500
	MostLoop        = 30
)

var (
	ToolsJsonErr = errors.New("tools json error")
)

type LLM struct {
	MessageChan chan *param.MsgInfo
	HTTPMsgChan chan string
	Content     string
	RagContext  string
	Images      [][]byte

	Model string
	Cs    *param.ContextState

	ChatId           string
	UserId           string
	MsgId            string
	PerMsgLen        int
	ContentParameter map[string]string

	LLMClient LLMClient

	Ctx context.Context

	DeepseekTools []godeepseek.Tool
	VolTools      []*model.Tool
	OpenAITools   []openai.Tool
	GeminiTools   []*genai.Tool

	WholeContent string // whole answer from llm
	LoopNum      int

	ToolApprovals bool
	searchSession *utils.WebSearchSession

	// providerOverride/modelOverride pin a specific provider+model for this
	// request: utility routing (cheap internal calls) and failover rungs.
	// Empty = resolve from the user config as usual.
	providerOverride string
	modelOverride    string

	// ConfigExecutor runs MuseBot configuration actions (config/mcp/skill/
	// cron/command/file) from inside the /task workflow. It is injected by the
	// robot layer to avoid an llm -> robot import cycle; approve lets the
	// executor ask the user before expanding the runtime boundary.
	ConfigExecutor func(action *param.ConversationConfigAction, approve TaskApprovalFunc) (string, error)
}

// TaskApprovalFunc asks the triggering user to confirm an operation; it
// returns true when the user approved within the timeout window.
type TaskApprovalFunc func(operation, target, detail string) (bool, error)

type LLMClient interface {
	Send(ctx context.Context, l *LLM) error

	GetMessage(role, msg string)

	GetImageMessage(image [][]byte, msg string)

	GetAudioMessage(audio []byte, msg string)

	AppendMessages(client LLMClient)

	SyncSend(ctx context.Context, l *LLM) (string, error)

	GetModel(l *LLM)
}

func (l *LLM) CallLLM() error {

	totalContent := l.GetContent(l.getPrompt())
	l.InsertCharacter(l.Ctx)
	l.GetMessages(l.UserId, totalContent)
	l.LLMClient.GetModel(l)

	logger.InfoCtx(l.Ctx, "msg receive", "userID", l.UserId, "prompt", totalContent, "type",
		utils.GetTxtType(db.GetCtxUserInfo(l.Ctx).LLMConfigRaw), "model", l.Model)

	metrics.APIRequestCount.WithLabelValues(l.Model).Inc()

	err := l.sendWithFailover()
	if err != nil {
		logger.ErrorCtx(l.Ctx, "Error calling LLM API", "err", err)
		return err
	}

	err = l.InsertOrUpdate()
	if err != nil {
		logger.ErrorCtx(l.Ctx, "insert or update record", "err", err)
		return err
	}

	return nil
}

func (l *LLM) GetContent(content string) string {
	return content
}

func (l *LLM) getPrompt() string {
	if l.RagContext == "" {
		return l.Content
	}

	prompt := i18n.GetMessage("rag_prompt", map[string]interface{}{
		"question": l.Content,
		"context":  l.RagContext,
	})
	if prompt == "" {
		return l.Content
	}
	return prompt
}

func (l *LLM) InsertCharacter(ctx context.Context) {
	if conf.BaseConfInfo.Character != "" {
		if l.ContentParameter != nil {
			tmpl, err := template.New("character").Parse(conf.BaseConfInfo.Character)
			if err != nil {
				logger.ErrorCtx(ctx, "parse template fail", "err", err)
				return
			}

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, l.ContentParameter)
			if err != nil {
				logger.ErrorCtx(ctx, "exec template fail", "err", err)
				return
			}

			logger.InfoCtx(ctx, "character", "character", buf.String())
			l.LLMClient.GetMessage(openai.ChatMessageRoleSystem, buf.String())
		}
	}
}

func NewLLM(opts ...Option) *LLM {
	l := new(LLM)
	l.searchSession = new(utils.WebSearchSession)
	l.Cs = new(param.ContextState)
	for _, opt := range opts {
		opt(l)
	}

	l.LLMClient = NewChatClient(l.Ctx)

	return l
}

func (l *LLM) DirectSendMsg(content string, ignoreLen bool) {
	if !ignoreLen && len([]byte(content)) > l.PerMsgLen {
		content = string([]byte(content)[:l.PerMsgLen])
	}

	if l.MessageChan != nil {
		l.MessageChan <- &param.MsgInfo{
			Content:  content,
			Finished: true,
		}
	}

	if l.HTTPMsgChan != nil {
		l.HTTPMsgChan <- content
	}
}

func (l *LLM) SendMsg(msgInfoContent *param.MsgInfo, content string) *param.MsgInfo {
	if l.MessageChan != nil {
		if l.PerMsgLen == 0 {
			l.PerMsgLen = OneMsgLen
		}

		// exceed max one message length
		if len([]byte(msgInfoContent.Content)) > l.PerMsgLen {
			msgInfoContent.Finished = true
			l.MessageChan <- msgInfoContent
			msgInfoContent = &param.MsgInfo{
				SendLen: NonFirstSendLen,
			}
		}

		msgInfoContent.Content += content
		l.WholeContent += content
		if len(msgInfoContent.Content) > msgInfoContent.SendLen {
			l.MessageChan <- msgInfoContent
			msgInfoContent.SendLen += NonFirstSendLen
		}

		return msgInfoContent
	} else {
		l.WholeContent += content
		// HTTPMsgChan may be nil when no consumer is attached (e.g. the task
		// planner); sending on a nil channel would block forever
		if l.HTTPMsgChan != nil {
			l.HTTPMsgChan <- content
		}
		return nil
	}
}

func (l *LLM) OverLoop() bool {
	if l.LoopNum >= MostLoop {
		return true
	}
	l.LoopNum++
	return false
}

func (l *LLM) InsertOrUpdate() error {
	if l.Cs.RecordID == 0 {
		db.InsertMsgRecord(l.Ctx, l.UserId, &db.AQ{
			Question:   l.Content,
			Answer:     l.WholeContent,
			Token:      l.Cs.Token,
			CreateTime: time.Now().Unix(),
		}, true)
		return nil
	}

	db.InsertMsgRecord(l.Ctx, l.UserId, &db.AQ{
		Question:   l.Content,
		Answer:     l.WholeContent,
		CreateTime: time.Now().Unix(),
	}, false)
	err := db.UpdateRecordInfo(&db.Record{
		ID:     l.Cs.RecordID,
		Answer: l.WholeContent,
		Token:  l.Cs.Token,
		UserId: l.UserId,
		Mode:   utils.GetTxtType(db.GetCtxUserInfo(l.Ctx).LLMConfigRaw),
	})
	if err != nil {
		logger.ErrorCtx(l.Ctx, "update record fail", "err", err)
		return err
	}

	return nil
}

func (l *LLM) GetMessages(userId string, prompt string) {
	l.appendHistory(userId)
	l.appendLatestImage(userId)

	if len(l.Images) > 0 {
		l.LLMClient.GetImageMessage(l.Images, prompt)
	} else {
		l.LLMClient.GetMessage(openai.ChatMessageRoleUser, prompt)
	}

}

// appendHistory appends the user's recent chat history (and any compression
// summary) to the conversation without appending a new user message, so
// follow-up flows can resolve references to earlier conversations.
func (l *LLM) appendHistory(userId string) {
	msgRecords := db.GetMsgRecord(userId)
	if msgRecords == nil || !l.Cs.UseRecord {
		return
	}

	if conf.BaseConfInfo.EnableAutoCompress {
		db.MaybeCompress(l.Ctx, userId, msgRecords,
			conf.BaseConfInfo.CompressThreshold, conf.BaseConfInfo.CompressKeepPairs,
			CompressHistory)
	}

	if msgRecords.Summary != "" {
		l.LLMClient.GetMessage(openai.ChatMessageRoleSystem,
			i18n.GetMessage("compress_summary_context", map[string]interface{}{
				"summary": msgRecords.Summary,
			}))
	}

	if memories := db.GetMemories(userId); len(memories) > 0 {
		lines := make([]string, 0, len(memories))
		for _, m := range memories {
			lines = append(lines, "- "+m.Content)
		}
		l.LLMClient.GetMessage(openai.ChatMessageRoleSystem,
			i18n.GetMessage("memory_context", map[string]interface{}{
				"memories": strings.Join(lines, "\n"),
			}))
	}

	aqs := db.FilterByMaxContextFromLatest(msgRecords.AQs, param.DefaultContextToken)
	for i, record := range aqs {
		if record.Question != "" && record.Answer != "" && record.CreateTime > time.Now().Unix()-int64(conf.BaseConfInfo.ContextExpireTime) {
			logger.InfoCtx(l.Ctx, "context content", "dialog", i, "question:", record.Question, "answer:", record.Answer)
			l.LLMClient.GetMessage(openai.ChatMessageRoleUser, record.Question)
			l.LLMClient.GetMessage(openai.ChatMessageRoleAssistant, record.Answer)
		}
	}
}

// appendLatestImage keeps the most recent picture of the conversation (usually
// a photo the bot just generated, else the user's last upload) visible to the
// current chat turn, so follow-up prompts like "make it brighter" resolve
// against it. Skipped when the current turn already carries its own images.
func (l *LLM) appendLatestImage(userId string) {
	if len(l.Images) > 0 || !l.Cs.UseRecord {
		return
	}
	record, err := db.GetLastImageRecord(userId)
	if err != nil {
		logger.WarnCtx(l.Ctx, "get last image record fail", "err", err)
		return
	}
	if record == nil || record.Answer == "" {
		return
	}
	if record.CreateTime > 0 && record.CreateTime <= time.Now().Unix()-int64(conf.BaseConfInfo.ContextExpireTime) {
		return
	}
	content, err := utils.DecodeStoredImage(record.Answer)
	if err != nil {
		logger.WarnCtx(l.Ctx, "decode last image record fail", "err", err)
		return
	}
	if len(content) == 0 {
		return
	}
	l.LLMClient.GetImageMessage([][]byte{content}, i18n.GetMessage("latest_image_context", nil))
}

// CompressHistory summarizes oldSummary + oldAQs into a single condensed summary.
// It is injected into db.MaybeCompress to avoid a db -> llm import cycle. The call
// runs synchronously and must not itself load history (UseRecord = false) to avoid
// recursion.
func CompressHistory(ctx context.Context, userId, oldSummary string, oldAQs []*db.AQ) (string, error) {
	var sb strings.Builder
	if oldSummary != "" {
		sb.WriteString(oldSummary)
		sb.WriteString("\n\n")
	}
	for _, aq := range oldAQs {
		if aq.Question != "" {
			sb.WriteString("Q: " + aq.Question + "\n")
		}
		if aq.Answer != "" {
			sb.WriteString("A: " + aq.Answer + "\n")
		}
	}

	prompt := i18n.GetMessage("compress_history_prompt", map[string]interface{}{
		"history": sb.String(),
	})

	l := NewLLM(WithUserId(userId), WithContent(prompt), WithContext(ctx), WithUtilityModel())
	l.Cs.UseRecord = false
	l.LLMClient.GetModel(l)
	l.LLMClient.GetMessage(openai.ChatMessageRoleUser, prompt)

	metrics.APIRequestCount.WithLabelValues(l.Model).Inc()
	summary, err := l.LLMClient.SyncSend(ctx, l)
	if err != nil {
		return "", err
	}

	if l.Cs.Token > 0 {
		if addErr := db.AddToken(userId, l.Cs.Token); addErr != nil {
			logger.ErrorCtx(ctx, "compress add token fail", "err", addErr)
		}
	}

	return strings.TrimSpace(summary), nil
}

type Option func(p *LLM)

func WithModel(model string) Option {
	return func(p *LLM) {
		p.Model = model
	}
}

func WithContent(content string) Option {
	return func(p *LLM) {
		p.Content = content
	}
}

// WithUtilityModel routes this request through the configured utility
// provider/model (cheap internal calls: history compression, memory,
// smart mode). No-op when utility routing is not configured.
func WithUtilityModel() Option {
	return func(p *LLM) {
		if conf.BaseConfInfo.UtilityType != "" && conf.BaseConfInfo.UtilityModel != "" {
			p.providerOverride = conf.BaseConfInfo.UtilityType
			p.modelOverride = conf.BaseConfInfo.UtilityModel
		}
	}
}

func WithRagContext(ragContext string) Option {
	return func(p *LLM) {
		p.RagContext = ragContext
	}
}

func WithPerMsgLen(perMsgLen int) Option {
	return func(p *LLM) {
		p.PerMsgLen = perMsgLen
	}
}

func WithMessageChan(messageChan chan *param.MsgInfo) Option {
	return func(p *LLM) {
		p.MessageChan = messageChan
	}
}

func WithHTTPMsgChan(messageChan chan string) Option {
	return func(p *LLM) {
		p.HTTPMsgChan = messageChan
	}
}

func WithChatId(chatId string) Option {
	return func(p *LLM) {
		p.ChatId = chatId
	}
}

func WithUserId(userId string) Option {
	return func(p *LLM) {
		p.UserId = userId
	}
}

func WithMsgId(msgId string) Option {
	return func(p *LLM) {
		p.MsgId = msgId
	}
}

func WithCS(cs *param.ContextState) Option {
	return func(p *LLM) {
		p.Cs = cs
	}
}

func WithImages(images [][]byte) Option {
	return func(p *LLM) {
		p.Images = images
	}
}

func WithTaskTools(taskTool *conf.AgentInfo) Option {
	return func(p *LLM) {
		if taskTool == nil {
			p.DeepseekTools = nil
			p.VolTools = nil
			p.OpenAITools = nil
			p.GeminiTools = nil
			return
		}
		p.DeepseekTools = taskTool.DeepseekTool
		p.VolTools = taskTool.VolTool
		p.OpenAITools = taskTool.OpenAITools
		p.GeminiTools = taskTool.GeminiTools
	}
}

func WithToolApprovals(enabled bool) Option {
	return func(p *LLM) {
		p.ToolApprovals = enabled
	}
}

func WithConfigExecutor(executor func(action *param.ConversationConfigAction, approve TaskApprovalFunc) (string, error)) Option {
	return func(p *LLM) {
		p.ConfigExecutor = executor
	}
}

func WithContext(ctx context.Context) Option {
	return func(p *LLM) {
		p.Ctx = ctx
	}
}

func WithContentParameter(contentParameter map[string]string) Option {
	return func(p *LLM) {
		p.ContentParameter = contentParameter
	}
}

func (l *LLM) ExecMcpReq(ctx context.Context, funcName string, property map[string]interface{}) (string, error) {
	if funcName == conf.BuiltinWebSearchTool {
		query, _ := property["query"].(string)
		switch strings.ToLower(strings.TrimSpace(conf.ToolsConfInfo.WebSearchProvider)) {
		case "", "default":
			return utils.WebSearch(ctx, query)
		case "parallel":
			if l.searchSession == nil {
				return "", errors.New("web search session is not initialized")
			}
			return utils.ParallelWebSearch(ctx, query, l.searchSession.ID())
		case "disabled":
			return "", errors.New("web search is disabled")
		default:
			return "", errors.New("unknown web search provider")
		}
	}
	if funcName == conf.BuiltinConfigTool {
		return l.executeTaskConfigTool(ctx, property)
	}
	if l.ToolApprovals && funcName == conf.BuiltinShellTool {
		return l.executeTaskShellTool(ctx, property)
	}

	mc, err := clients.GetMCPClientByToolName(funcName)
	if err != nil {
		logger.ErrorCtx(ctx, "get mcp fail", "err", err, "function", funcName, "argument", property)
		return "", err
	}

	metrics.MCPRequestCount.WithLabelValues(mc.Conf.Name, funcName).Inc()
	startTime := time.Now()

	var toolsData string
	toolsData, err = retryRequest(ctx, func() (string, error) {
		return mc.ExecTools(ctx, funcName, property)
	})

	if err != nil {
		logger.ErrorCtx(ctx, "get mcp fail", "err", err, "function", funcName, "argument", property)
		return "", err
	}

	metrics.MCPRequestDuration.WithLabelValues(mc.Conf.Name, funcName).Observe(time.Since(startTime).Seconds())

	logger.InfoCtx(ctx, "get mcp", "function", funcName, "argument", property, "res", toolsData)

	// 无论 SendMcpRes 是否开启，都检测图片类型并单独发送
	jsonRegex := regexp.MustCompile(`(\{\s*"type"\s*:[\s\S]*?\})`)
	imageSent := false
	for _, match := range jsonRegex.FindAllString(toolsData, -1) {
		if len(match) < 1 {
			continue
		}
		mcpResp := new(param.MCPResp)
		if err := json.Unmarshal([]byte(match), mcpResp); err == nil && mcpResp.Type == "image" {
			markdown := "![image](data:" + mcpResp.MimeType + ";base64," + mcpResp.Data + ")"
			l.DirectSendMsg(markdown, true)
			imageSent = true
			// 图片 base64 不传给 LLM，替换为占位符
			if !conf.BaseConfInfo.SendMcpMediaToLLM {
				toolsData = strings.Replace(toolsData, match, "[MCP 返回的 Base64 图像已直接发送给用户]", -1)
			}
			break
		}
	}

	if conf.BaseConfInfo.SendMcpRes {
		sendContent := toolsData
		if !imageSent {
			// 非图片类型，直接发送原始内容
			l.DirectSendMsg(i18n.GetMessage("send_mcp_info", map[string]interface{}{
				"function_name": funcName,
				"request_args":  property,
				"response":      sendContent,
			}), false)
		} else {
			// 图片已单独发送，sendContent 已是替换后的内容，发送文字部分
			l.DirectSendMsg(i18n.GetMessage("send_mcp_info", map[string]interface{}{
				"function_name": funcName,
				"request_args":  property,
				"response":      sendContent,
			}), false)
		}
	}

	return toolsData, nil
}
