package robot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	cronC "github.com/robfig/cron/v3"
	"github.com/sashabaranov/go-openai"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/mcp-client-go/clients"
	mcpParam "github.com/yincongcyincong/mcp-client-go/clients/param"
)

const (
	maxConversationConfigActions = 20
	maxFileReadBytes             = 256 * 1024
	maxFileWriteBytes            = 1024 * 1024
	maxCommandOutputBytes        = 16 * 1024
	maxFileListEntries           = 500
)

var (
	configNameReg      = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,64}$`)
	sensitiveConfig    = regexp.MustCompile(`(?i)(token|secret|password|passphrase|credential|db_conf|encoding_aes|cert|key_file|ca_file|crt_file|_ak$|_sk$)`)
	configCandidateReg = regexp.MustCompile(`(?i)(mcp|skill|cron|config|configuration|setting|settings|schedule|scheduled|remind|reminder|model|token|switch|change|update|add|create|delete|reload|sync|enable|disable|daily|weekly|command|shell|terminal|file|directory|folder|read|write|edit|list|run|execute|move|mv|remove|rm|mkdir|touch|cat|view|save|rename|ls|git|加载|配置|设置|切换|修改|新增|添加|创建|删除|重载|同步|技能|定时|计划任务|提醒|每天|每周|开启|启用|关闭|禁用|模型|令牌|密钥|执行|运行|跑|命令|终端|文件|目录|文件夹|读取|写入|编辑|移动|移到|重命名|保存|查看)`)
)

func (r *RobotInfo) tryConversationConfig() bool {
	if r.Robot.getCommand() != "" || r.Robot.getPrompt() == "" ||
		!configCandidateReg.MatchString(r.Robot.getPrompt()) {
		return false
	}

	intent, err := r.detectConversationConfigIntent()
	if err != nil {
		logger.ErrorCtx(r.Ctx, "detect conversation config intent fail", "err", err)
		return false
	}
	if !intent.Config {
		return false
	}

	_, _, userId := r.GetChatIdAndMsgIdAndUserID()
	logger.InfoCtx(r.Ctx, "conversation config request", "userID", userId, "prompt", r.Robot.getPrompt())
	r.handleConversationConfig()
	return true
}

func (r *RobotInfo) detectConversationConfigIntent() (*param.ConversationConfigIntent, error) {
	llmClient := llm.NewLLM(
		llm.WithChatId(r.ChatIdValue()),
		llm.WithUserId(r.UserIdValue()),
		llm.WithMsgId(r.MsgIdValue()),
		llm.WithContent(i18n.GetMessage("conversation_config_intent_prompt", map[string]interface{}{
			"request": r.Robot.getPrompt(),
		})),
		llm.WithContext(r.Ctx),
	)
	llmClient.LLMClient.GetModel(llmClient)
	llmClient.LLMClient.GetMessage(openai.ChatMessageRoleUser, llmClient.Content)
	content, err := llmClient.LLMClient.SyncSend(r.Ctx, llmClient)
	if err != nil {
		return nil, err
	}
	r.cs.Token += llmClient.Cs.Token

	return param.ParseConversationConfigIntent(content)
}

func (r *RobotInfo) handleConversationConfig() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	if !r.isAdminUser(userId) {
		r.SendMsg(chatId, i18n.GetMessage("config_permission_denied", nil), msgId, "", nil)
		return
	}
	if strings.TrimSpace(r.Robot.getPrompt()) == "" {
		r.SendMsg(chatId, i18n.GetMessage("config_usage", nil), msgId, "", nil)
		return
	}

	plan, err := r.buildConversationConfigPlan()
	if err != nil {
		logger.ErrorCtx(r.Ctx, "build conversation config plan fail", "err", err)
		r.SendMsg(chatId, err.Error(), msgId, "", nil)
		return
	}

	results := make([]string, 0, len(plan.Actions))
	for _, action := range plan.Actions {
		result, err := r.executeConversationConfigAction(action, userId)
		if err != nil {
			logger.ErrorCtx(r.Ctx, "execute conversation config action fail", "type", action.Type, "operation", action.Operation, "err", err)
			results = append(results, fmt.Sprintf("❌ %s %s: %s", action.Type, action.Operation, err.Error()))
			continue
		}
		results = append(results, "✅ "+result)
	}
	r.SendMsg(chatId, strings.Join(results, "\n"), msgId, "", nil)
}

func (r *RobotInfo) isAdminUser(userId string) bool {
	if len(conf.BaseConfInfo.AdminUserIds) == 0 || conf.BaseConfInfo.AdminUserIds["0"] {
		return false
	}
	return conf.BaseConfInfo.AdminUserIds[userId]
}

func (r *RobotInfo) buildConversationConfigPlan() (*param.ConversationConfigPlan, error) {
	prompt := i18n.GetMessage("conversation_config_prompt", map[string]interface{}{
		"schema":  conversationConfigSchema(),
		"mcp":     conversationMCPSummary(),
		"skills":  conversationSkillSummary(),
		"builtin": conversationBuiltinSummary(),
		"request": r.Robot.getPrompt(),
	})

	llmClient := llm.NewLLM(
		llm.WithChatId(r.ChatIdValue()),
		llm.WithUserId(r.UserIdValue()),
		llm.WithMsgId(r.MsgIdValue()),
		llm.WithContent(prompt),
		llm.WithContext(r.Ctx),
		llm.WithCS(r.cs),
	)
	llmClient.LLMClient.GetModel(llmClient)
	llmClient.LLMClient.GetMessage(openai.ChatMessageRoleUser, prompt)
	content, err := llmClient.LLMClient.SyncSend(r.Ctx, llmClient)
	if err != nil {
		return nil, fmt.Errorf("get configuration plan fail: %w", err)
	}
	r.cs.Token = llmClient.Cs.Token

	plan, err := param.ParseConversationConfigPlan(content)
	if err != nil {
		return nil, err
	}
	if len(plan.Actions) > maxConversationConfigActions {
		return nil, fmt.Errorf("too many configuration actions")
	}
	return plan, nil
}

func (r *RobotInfo) executeConversationConfigAction(action *param.ConversationConfigAction, userId string) (string, error) {
	if action == nil {
		return "", fmt.Errorf("empty action")
	}
	if action.Type == "cron" {
		logger.InfoCtx(r.Ctx, "conversation cron action", "operation", action.Operation,
			"id", action.ID.Int64(), "name", action.Name, "cron", action.Cron,
			"prompt", action.Prompt, "command", action.Command)
	}
	switch action.Type {
	case "config":
		return r.executeConfigAction(action)
	case "mcp":
		return r.executeMCPAction(action)
	case "skill":
		return r.executeSkillAction(action)
	case "cron":
		return r.executeConversationCronAction(action, userId)
	case "command":
		return r.executeConversationCommandAction(action)
	case "file":
		return r.executeConversationFileAction(action)
	default:
		return "", fmt.Errorf("unsupported action type %q", action.Type)
	}
}

func conversationConfigTargets() map[string]interface{} {
	return map[string]interface{}{
		"base":     conf.BaseConfInfo,
		"audio":    conf.AudioConfInfo,
		"llm":      conf.LLMConfInfo,
		"photo":    conf.PhotoConfInfo,
		"rag":      conf.RagConfInfo,
		"video":    conf.VideoConfInfo,
		"register": conf.RegisterConfInfo,
		"tools":    conf.ToolsConfInfo,
	}
}

func conversationConfigSchema() string {
	types := make([]string, 0)
	for configType := range conversationConfigTargets() {
		types = append(types, configType)
	}
	sort.Strings(types)

	parts := make([]string, 0, len(types))
	for _, configType := range types {
		target := reflect.ValueOf(conversationConfigTargets()[configType])
		if target.Kind() == reflect.Ptr {
			target = target.Elem()
		}
		fields := make([]string, 0)
		for i := 0; i < target.NumField(); i++ {
			tag := target.Type().Field(i).Tag.Get("json")
			key := strings.Split(tag, ",")[0]
			if key == "" || key == "-" {
				continue
			}
			fields = append(fields, key)
		}
		parts = append(parts, configType+": "+strings.Join(fields, ", "))
	}
	return strings.Join(parts, "\n")
}

func (r *RobotInfo) executeConfigAction(action *param.ConversationConfigAction) (string, error) {
	targets := conversationConfigTargets()
	target, ok := targets[action.ConfigType]
	if !ok {
		return "", fmt.Errorf("unknown config type %q", action.ConfigType)
	}

	switch action.Operation {
	case "list":
		return "config fields:\n" + conversationConfigSchema(), nil
	case "get":
		return conversationConfigValue(target, action.Key)
	case "set":
		if err := validateConversationConfigAction(action); err != nil {
			return "", err
		}
		if err := setConversationConfigField(target, action.Key, action.Value); err != nil {
			return "", err
		}
		conf.SaveConf()
		if action.ConfigType == "tools" && action.Key == "skill_path" {
			conf.InitSkills()
		}
		return fmt.Sprintf("set %s.%s", action.ConfigType, action.Key), nil
	default:
		return "", fmt.Errorf("unsupported config operation %q", action.Operation)
	}
}

func validateConversationConfigAction(action *param.ConversationConfigAction) error {
	if action.Operation != "set" {
		return nil
	}
	if action.ConfigType == "base" && action.Key == "admin_user_ids" {
		adminIDs := make(map[string]bool)
		if err := json.Unmarshal(action.Value, &adminIDs); err != nil {
			return fmt.Errorf("decode admin_user_ids: %w", err)
		}
		if len(adminIDs) == 0 {
			return fmt.Errorf("admin user IDs cannot be empty")
		}
	}
	return nil
}

func setConversationConfigField(target interface{}, key string, value json.RawMessage) error {
	if isRuntimeImmutableConfigKey(key) {
		return fmt.Errorf("config key %q cannot be changed at runtime", key)
	}

	elem := reflect.ValueOf(target).Elem()
	typ := elem.Type()
	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldValue := elem.Field(i)
		if strings.Split(fieldType.Tag.Get("json"), ",")[0] != key || !fieldValue.CanSet() {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			newValue := reflect.New(fieldValue.Type().Elem())
			if err := json.Unmarshal(value, newValue.Interface()); err != nil {
				return fmt.Errorf("decode %s: %w", key, err)
			}
			fieldValue.Set(newValue)
		} else if err := json.Unmarshal(value, fieldValue.Addr().Interface()); err != nil {
			return fmt.Errorf("decode %s: %w", key, err)
		}
		return nil
	}
	return fmt.Errorf("unknown config key %q", key)
}

func isRuntimeImmutableConfigKey(key string) bool {
	switch key {
	case "bot_name", "http_host", "db_type", "db_conf", "allowed_commands", "file_root_path":
		return true
	default:
		return false
	}
}

func conversationConfigValue(target interface{}, key string) (string, error) {
	elem := reflect.ValueOf(target).Elem()
	typ := elem.Type()
	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		if strings.Split(fieldType.Tag.Get("json"), ",")[0] != key {
			continue
		}
		value := elem.Field(i)
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				return key + "=nil", nil
			}
			value = value.Elem()
		}
		if sensitiveConfig.MatchString(key) {
			return key + "=[redacted]", nil
		}
		data, err := json.Marshal(value.Interface())
		if err != nil {
			return "", err
		}
		return key + "=" + string(data), nil
	}
	return "", fmt.Errorf("unknown config key %q", key)
}

func conversationMCPSummary() string {
	config, err := readMCPConfig()
	if err != nil {
		return "MCP config unavailable: " + err.Error()
	}
	names := make([]string, 0, len(config.McpServers))
	for name := range config.McpServers {
		names = append(names, name)
	}
	sort.Strings(names)
	return "configured MCP servers: " + strings.Join(names, ", ")
}

func conversationSkillSummary() string {
	names := make([]string, 0)
	conf.SkillTools.Range(func(key, _ any) bool {
		names = append(names, fmt.Sprint(key))
		return true
	})
	sort.Strings(names)
	return "loaded skills: " + strings.Join(names, ", ")
}

func conversationBuiltinSummary() string {
	fileRoot := conversationFileRootOrDefault()
	return "allowed commands: " + conf.ToolsConfInfo.AllowedCommands +
		"; command timeout seconds: " + fmt.Sprint(conversationCommandTimeout()) +
		"; file operation root: " + fileRoot
}

func (r *RobotInfo) executeMCPAction(action *param.ConversationConfigAction) (string, error) {
	switch action.Operation {
	case "list":
		return conversationMCPList()
	case "set":
		return upsertConversationMCP(action)
	case "delete", "enable", "disable":
		return changeConversationMCP(action)
	case "sync":
		syncMCPAndSkills()
		return "synchronized MCP servers and skills", nil
	default:
		return "", fmt.Errorf("unsupported MCP operation %q", action.Operation)
	}
}

func conversationMCPList() (string, error) {
	config, err := readMCPConfig()
	if err != nil {
		return "", err
	}
	names := make([]string, 0, len(config.McpServers))
	for name := range config.McpServers {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		server := config.McpServers[name]
		parts = append(parts, fmt.Sprintf("%s: type=%s url=%s disabled=%t", name, server.Type, server.Url, server.Disabled))
	}
	return strings.Join(parts, "\n"), nil
}

func upsertConversationMCP(action *param.ConversationConfigAction) (string, error) {
	if !configNameReg.MatchString(action.Name) {
		return "", fmt.Errorf("invalid MCP name")
	}
	if len(action.Server) == 0 {
		return "", fmt.Errorf("MCP server config is required")
	}

	config, err := readMCPConfig()
	if err != nil {
		return "", err
	}
	server := config.McpServers[action.Name]
	if server == nil {
		server = new(mcpParam.MCPConfig)
	}
	if err := json.Unmarshal(action.Server, server); err != nil {
		return "", fmt.Errorf("decode MCP server: %w", err)
	}
	if server.Description == "" || (server.Url == "" && server.Command == "") {
		return "", fmt.Errorf("MCP description and either url or command are required")
	}
	config.McpServers[action.Name] = server
	if err := writeMCPConfig(config); err != nil {
		return "", err
	}
	syncMCPAndSkills()
	return "set and synchronized MCP server " + action.Name, nil
}

func changeConversationMCP(action *param.ConversationConfigAction) (string, error) {
	if !configNameReg.MatchString(action.Name) {
		return "", fmt.Errorf("invalid MCP name")
	}
	config, err := readMCPConfig()
	if err != nil {
		return "", err
	}
	server := config.McpServers[action.Name]
	if server == nil {
		return "", fmt.Errorf("MCP server %q does not exist", action.Name)
	}

	switch action.Operation {
	case "delete":
		delete(config.McpServers, action.Name)
		if err := clients.RemoveMCPClient(action.Name); err != nil {
			logger.Warn("remove MCP client fail", "name", action.Name, "err", err)
		}
		conf.TaskTools.Delete(action.Name)
	case "disable":
		server.Disabled = true
	case "enable":
		server.Disabled = false
	default:
		return "", fmt.Errorf("unsupported MCP operation %q", action.Operation)
	}
	if err := writeMCPConfig(config); err != nil {
		return "", err
	}
	syncMCPAndSkills()
	return action.Operation + "d MCP server " + action.Name, nil
}

func readMCPConfig() (*mcpParam.McpClientGoConfig, error) {
	data, err := os.ReadFile(*conf.ToolsConfInfo.McpConfPath)
	if err != nil {
		return nil, err
	}
	config := new(mcpParam.McpClientGoConfig)
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	if config.McpServers == nil {
		config.McpServers = make(map[string]*mcpParam.MCPConfig)
	}
	return config, nil
}

func writeMCPConfig(config *mcpParam.McpClientGoConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(*conf.ToolsConfInfo.McpConfPath, data, 0644)
}

func syncMCPAndSkills() {
	clients.ClearAllMCPClient()
	conf.TaskTools.Clear()
	conf.InitTools()
}

func (r *RobotInfo) executeSkillAction(action *param.ConversationConfigAction) (string, error) {
	switch action.Operation {
	case "list":
		return conversationSkillSummary(), nil
	case "reload":
		conf.InitSkills()
		return "reloaded skills", nil
	case "set_path":
		path := strings.TrimSpace(action.Path)
		if path == "" {
			return "", fmt.Errorf("skill path is required")
		}
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		info, err := os.Stat(absolutePath)
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("skill path is not a directory")
		}
		conf.ToolsConfInfo.SkillPath = &absolutePath
		conf.SaveConf()
		conf.InitSkills()
		return "set and reloaded skill path " + absolutePath, nil
	case "create":
		if err := conf.CreateSkill(action.Name, action.Description, action.Instructions); err != nil {
			return "", err
		}
		if _, ok := conf.SkillTools.Load(action.Name); !ok {
			return "", fmt.Errorf("skill name conflicts with an existing task agent")
		}
		return "created and loaded skill " + action.Name, nil
	default:
		return "", fmt.Errorf("unsupported skill operation %q", action.Operation)
	}
}

func (r *RobotInfo) executeConversationCommandAction(action *param.ConversationConfigAction) (string, error) {
	if action.Operation != "run" {
		return "", fmt.Errorf("unsupported command operation %q", action.Operation)
	}
	if len(action.Args) == 0 {
		return "", fmt.Errorf("command args are required")
	}

	executable, err := exec.LookPath(action.Args[0])
	if err != nil {
		return "", err
	}
	if !conversationCommandAllowed(executable) {
		return "", fmt.Errorf("command %q is not allowed", action.Args[0])
	}

	workingDir := conversationFileRootOrDefault()
	if action.Path != "" {
		workingDir, err = secureConversationFilePath(action.Path, true)
		if err != nil {
			return "", fmt.Errorf("invalid command working directory: %w", err)
		}
		info, err := os.Stat(workingDir)
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("command working directory is not a directory")
		}
	}

	timeout := conversationCommandTimeout()
	if action.Timeout > 0 {
		timeout = min(timeout, action.Timeout)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, executable, action.Args[1:]...)
	cmd.Dir = workingDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	output := formatCommandOutput(stdout.Bytes(), stderr.Bytes())
	result := "command completed"
	if ctx.Err() != nil {
		result = "command timed out"
	} else if runErr != nil {
		result = "command failed: " + runErr.Error()
	}
	if output != "" {
		result += "\n" + output
	}
	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	return result, runErr
}

func conversationCommandAllowed(executable string) bool {
	for _, allowedCommand := range strings.Split(conf.ToolsConfInfo.AllowedCommands, ",") {
		allowedCommand = strings.TrimSpace(allowedCommand)
		if allowedCommand == "" {
			continue
		}
		if allowedCommand == executable || allowedCommand == filepath.Base(executable) || allowedCommand == filepath.Clean(executable) {
			return true
		}
	}
	return false
}

func conversationCommandTimeout() int {
	if conf.ToolsConfInfo.CommandTimeoutSec <= 0 {
		return 60
	}
	return conf.ToolsConfInfo.CommandTimeoutSec
}

func formatCommandOutput(stdout, stderr []byte) string {
	output := bytes.TrimSpace(bytes.Join([][]byte{stdout, stderr}, []byte("\n--- stderr ---\n")))
	if len(output) > maxCommandOutputBytes {
		prefix := output[:maxCommandOutputBytes/2]
		suffix := output[len(output)-maxCommandOutputBytes/2:]
		output = bytes.Join([][]byte{prefix, []byte("\n... output truncated ...\n"), suffix}, nil)
	}
	return string(output)
}

func (r *RobotInfo) executeConversationFileAction(action *param.ConversationConfigAction) (string, error) {
	if action.Path == "" {
		return "", fmt.Errorf("file path is required")
	}

	switch action.Operation {
	case "list":
		return listConversationFile(action.Path)
	case "read":
		return readConversationFile(action.Path)
	case "write":
		return writeConversationFile(action.Path, action.Content)
	case "edit":
		return editConversationFile(action.Path, action.OldText, action.NewText)
	case "mkdir":
		targetPath, err := secureConversationFilePath(action.Path, false)
		if err != nil {
			return "", err
		}
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return "", err
		}
		return "created directory " + action.Path, nil
	case "delete":
		targetPath, err := secureConversationFilePath(action.Path, true)
		if err != nil {
			return "", err
		}
		rootPath, err := conversationFileRoot()
		if err != nil {
			return "", err
		}
		if targetPath == rootPath {
			return "", fmt.Errorf("cannot delete the file operation root")
		}
		if err := os.Remove(targetPath); err != nil {
			return "", err
		}
		return "deleted " + action.Path, nil
	case "move":
		return moveConversationFile(action.Path, action.Target)
	default:
		return "", fmt.Errorf("unsupported file operation %q", action.Operation)
	}
}

func conversationFileRootOrDefault() string {
	if conf.ToolsConfInfo.FileRootPath != nil && *conf.ToolsConfInfo.FileRootPath != "" {
		return *conf.ToolsConfInfo.FileRootPath
	}
	return conf.GetAbsPath("data/agent_files")
}

func conversationFileRoot() (string, error) {
	root, err := filepath.Abs(conversationFileRootOrDefault())
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(root)
}

func secureConversationFilePath(path string, mustExist bool) (string, error) {
	root, err := conversationFileRoot()
	if err != nil {
		return "", err
	}

	absolutePath := filepath.Clean(path)
	if !filepath.IsAbs(absolutePath) {
		absolutePath = filepath.Join(root, absolutePath)
	}
	if !conversationPathInsideRoot(root, absolutePath) {
		return "", fmt.Errorf("path must stay inside the file operation root")
	}

	info, err := os.Lstat(absolutePath)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			absolutePath, err = filepath.EvalSymlinks(absolutePath)
			if err != nil {
				return "", err
			}
		}
	} else if mustExist || !os.IsNotExist(err) {
		return "", err
	} else {
		parent := filepath.Dir(absolutePath)
		if err := os.MkdirAll(parent, 0755); err != nil {
			return "", err
		}
		parent, err = filepath.EvalSymlinks(parent)
		if err != nil {
			return "", err
		}
		absolutePath = filepath.Join(parent, filepath.Base(absolutePath))
	}

	if !conversationPathInsideRoot(root, absolutePath) {
		return "", fmt.Errorf("resolved path must stay inside the file operation root")
	}
	return absolutePath, nil
}

func conversationPathInsideRoot(root, path string) bool {
	relativePath, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator))
}

func listConversationFile(path string) (string, error) {
	targetPath, err := secureConversationFilePath(path, true)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return "", err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		names = append(names, name)
		if len(names) == maxFileListEntries {
			names = append(names, "... output truncated ...")
			break
		}
	}
	if len(names) == 0 {
		return "directory is empty", nil
	}
	return strings.Join(names, "\n"), nil
}

func readConversationFile(path string) (string, error) {
	targetPath, err := secureConversationFilePath(path, true)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(targetPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory")
	}
	if info.Size() > maxFileReadBytes {
		return "", fmt.Errorf("file exceeds %d bytes", maxFileReadBytes)
	}
	content, err := os.ReadFile(targetPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeConversationFile(path, content string) (string, error) {
	if len(content) > maxFileWriteBytes {
		return "", fmt.Errorf("file content exceeds %d bytes", maxFileWriteBytes)
	}
	targetPath, err := secureConversationFilePath(path, false)
	if err != nil {
		return "", err
	}
	if err := atomicWriteConversationFile(targetPath, []byte(content)); err != nil {
		return "", err
	}
	return "wrote " + path, nil
}

func editConversationFile(path, oldText, newText string) (string, error) {
	if oldText == "" || oldText == newText {
		return "", fmt.Errorf("file edit requires a non-empty old_text different from new_text")
	}
	targetPath, err := secureConversationFilePath(path, true)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(targetPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory")
	}
	if info.Size() > maxFileWriteBytes {
		return "", fmt.Errorf("file exceeds %d bytes", maxFileWriteBytes)
	}
	content, err := os.ReadFile(targetPath)
	if err != nil {
		return "", err
	}
	count := strings.Count(string(content), oldText)
	if count == 0 {
		return "", fmt.Errorf("old_text was not found")
	}
	updatedContent := strings.ReplaceAll(string(content), oldText, newText)
	if len(updatedContent) > maxFileWriteBytes {
		return "", fmt.Errorf("edited file exceeds %d bytes", maxFileWriteBytes)
	}
	if err := atomicWriteConversationFile(targetPath, []byte(updatedContent)); err != nil {
		return "", err
	}
	return fmt.Sprintf("replaced %d occurrence(s) in %s", count, path), nil
}

func moveConversationFile(path, target string) (string, error) {
	if target == "" {
		return "", fmt.Errorf("file target is required")
	}
	rootPath, err := conversationFileRoot()
	if err != nil {
		return "", err
	}
	sourcePath, err := secureConversationFilePath(path, true)
	if err != nil {
		return "", err
	}
	targetPath, err := secureConversationFilePath(target, false)
	if err != nil {
		return "", err
	}
	if sourcePath == rootPath || targetPath == rootPath {
		return "", fmt.Errorf("cannot move the file operation root")
	}
	if sourcePath == targetPath {
		return "", fmt.Errorf("source and target are the same")
	}
	if _, err := os.Lstat(targetPath); err == nil {
		return "", fmt.Errorf("target already exists")
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := os.Rename(sourcePath, targetPath); err != nil {
		return "", err
	}
	return "moved " + path + " to " + target, nil
}

func atomicWriteConversationFile(path string, content []byte) error {
	tempFile, err := os.CreateTemp(filepath.Dir(path), ".musebot-write-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(content); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Chmod(0644); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func (r *RobotInfo) executeConversationCronAction(action *param.ConversationConfigAction, userId string) (string, error) {
	switch action.Operation {
	case "create":
		if action.Cron == "" || action.Prompt == "" {
			return "", fmt.Errorf("cron expression and prompt are required")
		}
		if err := r.InsertCron(action.Cron, action.Prompt); err != nil {
			return "", err
		}
		return "created cron task", nil
	case "list":
		crons, err := db.GetCronsByPage(1, 20, "", userId)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(crons)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "get":
		return r.getConversationCron(action.ID.Int64())
	case "update":
		return r.updateConversationCron(action)
	case "enable", "disable":
		return r.setConversationCronStatus(action.ID.Int64(), action.Operation == "enable")
	case "delete":
		cronID := action.ID.Int64()
		cronInfo, err := r.getConversationCronInfo(cronID)
		if err != nil {
			return "", err
		}
		if err := db.DeleteCronByID(cronID); err != nil {
			return "", err
		}
		removeConversationCronSchedule(cronInfo)
		return fmt.Sprintf("deleted cron task %d", cronID), nil
	case "clear":
		crons, err := db.GetCronsByPage(1, 1000, "", userId)
		if err != nil {
			return "", err
		}
		if err := db.DeleteCronByCreateBy(userId, ""); err != nil {
			return "", err
		}
		for _, cronInfo := range crons {
			removeConversationCronSchedule(&cronInfo)
		}
		return "cleared your cron tasks", nil
	default:
		return "", fmt.Errorf("unsupported cron operation %q", action.Operation)
	}
}

func (r *RobotInfo) getConversationCron(id int64) (string, error) {
	cronInfo, err := db.GetCronByID(id)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(cronInfo)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *RobotInfo) getConversationCronInfo(id int64) (*db.Cron, error) {
	if id == 0 {
		return nil, fmt.Errorf("cron id is required")
	}
	return db.GetCronByID(id)
}

func (r *RobotInfo) updateConversationCron(action *param.ConversationConfigAction) (string, error) {
	cronInfo, err := r.getConversationCronInfo(action.ID.Int64())
	if err != nil {
		return "", err
	}
	if action.Cron != "" {
		cronInfo.CronSpec = action.Cron
	}
	if action.Prompt != "" {
		cronInfo.Prompt = action.Prompt
	}
	if action.Name != "" {
		cronInfo.CronName = action.Name
	}
	if action.Command != "" {
		cronInfo.Command = action.Command
	}
	if err := db.UpdateCron(cronInfo.ID, cronInfo.CronName, cronInfo.CronSpec, cronInfo.TargetID,
		cronInfo.GroupID, cronInfo.Command, cronInfo.Prompt, cronInfo.Type); err != nil {
		return "", err
	}
	removeConversationCronSchedule(cronInfo)
	if cronInfo.Status == 1 {
		updated, err := db.GetCronByID(cronInfo.ID)
		if err != nil {
			return "", err
		}
		if err := AddCron(updated); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("updated cron task %d", action.ID), nil
}

func (r *RobotInfo) setConversationCronStatus(id int64, enabled bool) (string, error) {
	cronInfo, err := r.getConversationCronInfo(id)
	if err != nil {
		return "", err
	}
	status := 0
	if enabled {
		status = 1
	}
	if cronInfo.Status == status {
		return fmt.Sprintf("cron task %d already has status %d", id, status), nil
	}
	if err := db.UpdateCronStatus(id, status); err != nil {
		return "", err
	}
	removeConversationCronSchedule(cronInfo)
	if enabled {
		updated, err := db.GetCronByID(id)
		if err != nil {
			return "", err
		}
		if err := AddCron(updated); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("cron task %d status set to %d", id, status), nil
}

func removeConversationCronSchedule(cronInfo *db.Cron) {
	if Cron != nil && cronInfo.Status == 1 && cronInfo.CronJobId > 0 {
		Cron.Remove(cronC.EntryID(cronInfo.CronJobId))
	}
}

func (r *RobotInfo) ChatIdValue() string {
	chatId, _, _ := r.GetChatIdAndMsgIdAndUserID()
	return chatId
}

func (r *RobotInfo) UserIdValue() string {
	_, _, userId := r.GetChatIdAndMsgIdAndUserID()
	return userId
}

func (r *RobotInfo) MsgIdValue() string {
	_, msgId, _ := r.GetChatIdAndMsgIdAndUserID()
	return msgId
}
