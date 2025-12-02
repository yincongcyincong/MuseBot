package llm

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"time"
	
	godeepseek "github.com/cohesion-org/deepseek-go"
	"github.com/revrost/go-openrouter"
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
	MostLoop        = 15
)

var (
	ToolsJsonErr = errors.New("tools json error")
)

type LLM struct {
	MessageChan chan *param.MsgInfo
	HTTPMsgChan chan string
	Content     string
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
	
	DeepseekTools   []godeepseek.Tool
	VolTools        []*model.Tool
	OpenAITools     []openai.Tool
	GeminiTools     []*genai.Tool
	OpenRouterTools []openrouter.Tool
	
	WholeContent string // whole answer from llm
	LoopNum      int
}

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
	
	totalContent := l.GetContent(l.Content)
	l.GetMessages(l.UserId, totalContent)
	l.InsertCharacter(l.Ctx)
	l.LLMClient.GetModel(l)
	
	logger.InfoCtx(l.Ctx, "msg receive", "userID", l.UserId, "prompt", totalContent, "type",
		utils.GetTxtType(db.GetCtxUserInfo(l.Ctx).LLMConfigRaw), "model", l.Model)
	
	metrics.APIRequestCount.WithLabelValues(l.Model).Inc()
	
	var err error
	if conf.BaseConfInfo.IsStreaming {
		err = l.LLMClient.Send(l.Ctx, l)
		if err != nil {
			logger.ErrorCtx(l.Ctx, "Error calling LLM API", "err", err)
			return err
		}
	} else {
		content, err := l.LLMClient.SyncSend(l.Ctx, l)
		if err != nil {
			logger.ErrorCtx(l.Ctx, "Error calling LLM API", "err", err)
			return err
		}
		
		l.MessageChan <- &param.MsgInfo{
			Content: content,
		}
		l.WholeContent = content
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
	l.Cs = new(param.ContextState)
	for _, opt := range opts {
		opt(l)
	}
	
	switch utils.GetTxtType(db.GetCtxUserInfo(l.Ctx).LLMConfigRaw) {
	case param.Ollama:
		l.LLMClient = &OllamaReq{
			ToolCall:           []godeepseek.ToolCall{},
			ToolMessage:        []godeepseek.ChatCompletionMessage{},
			CurrentToolMessage: []godeepseek.ChatCompletionMessage{},
		}
	default:
		l.LLMClient = &OpenAIReq{
			ToolCall:           []openai.ToolCall{},
			ToolMessage:        []openai.ChatCompletionMessage{},
			CurrentToolMessage: []openai.ChatCompletionMessage{},
		}
	}
	
	return l
}

func (l *LLM) DirectSendMsg(content string) {
	if len([]byte(content)) > l.PerMsgLen {
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
		l.HTTPMsgChan <- content
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
	msgRecords := db.GetMsgRecord(userId)
	if msgRecords != nil && l.Cs.UseRecord {
		aqs := db.FilterByMaxContextFromLatest(msgRecords.AQs, param.DefaultContextToken)
		for i, record := range aqs {
			if record.Question != "" && record.Answer != "" && record.CreateTime > time.Now().Unix()-int64(conf.BaseConfInfo.ContextExpireTime) {
				logger.InfoCtx(l.Ctx, "context content", "dialog", i, "question:", record.Question, "answer:", record.Answer)
				l.LLMClient.GetMessage(openai.ChatMessageRoleUser, record.Question)
				l.LLMClient.GetMessage(openai.ChatMessageRoleAssistant, record.Answer)
			}
		}
	}
	
	if len(l.Images) > 0 {
		l.LLMClient.GetImageMessage(l.Images, prompt)
	} else {
		l.LLMClient.GetMessage(openai.ChatMessageRoleUser, prompt)
	}
	
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
			p.OpenRouterTools = nil
			return
		}
		p.DeepseekTools = taskTool.DeepseekTool
		p.VolTools = taskTool.VolTool
		p.OpenAITools = taskTool.OpenAITools
		p.GeminiTools = taskTool.GeminiTools
		p.OpenRouterTools = taskTool.OpenRouterTools
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
	mc, err := clients.GetMCPClientByToolName(funcName)
	if err != nil {
		logger.ErrorCtx(ctx, "get mcp fail", "err", err, "function", funcName, "argument", property)
		return "", err
	}
	
	metrics.MCPRequestCount.WithLabelValues(mc.Conf.Name, funcName).Inc()
	startTime := time.Now()
	
	var toolsData string
	for i := 0; i < conf.BaseConfInfo.LLMRetryTimes; i++ {
		toolsData, err = mc.ExecTools(ctx, funcName, property)
		if err != nil {
			logger.ErrorCtx(ctx, "get mcp fail", "err", err, "function", funcName, "argument", property)
			continue
		}
		break
	}
	
	if err != nil {
		logger.ErrorCtx(ctx, "get mcp fail", "err", err, "function", funcName, "argument", property)
		return "", err
	}
	
	metrics.MCPRequestDuration.WithLabelValues(mc.Conf.Name, funcName).Observe(time.Since(startTime).Seconds())
	
	logger.InfoCtx(ctx, "get mcp", "function", funcName, "argument", property, "res", toolsData)
	
	if conf.BaseConfInfo.SendMcpRes {
		l.DirectSendMsg(i18n.GetMessage("send_mcp_info", map[string]interface{}{
			"function_name": funcName,
			"request_args":  property,
			"response":      toolsData,
		}))
	}
	
	return toolsData, nil
}
