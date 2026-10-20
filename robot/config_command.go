package robot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	cronC "github.com/robfig/cron/v3"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/mcp-client-go/clients"
	mcpParam "github.com/yincongcyincong/mcp-client-go/clients/param"
)

const (
	maxConversationConfigActions = 20
	maxCommandOutputBytes        = 16 * 1024
)

var (
	configNameReg   = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,64}$`)
	sensitiveConfig = regexp.MustCompile(`(?i)(token|secret|password|passphrase|credential|db_conf|encoding_aes|cert|key_file|ca_file|crt_file|_ak$|_sk$)`)
)

// isAdminUser reports whether the user may run MuseBot configuration actions.
func (r *RobotInfo) isAdminUser(userId string) bool {
	if len(conf.BaseConfInfo.AdminUserIds) == 0 || conf.BaseConfInfo.AdminUserIds["0"] {
		return false
	}
	return conf.BaseConfInfo.AdminUserIds[userId]
}

// executeConversationConfigAction runs a single configuration action. It is
// exposed to the /task workflow through the bot_config tool; the admin gate
// is enforced here so every entry path is protected. approve (when non-nil)
// lets boundary expansions ask the triggering user first.
func (r *RobotInfo) executeConversationConfigAction(action *param.ConversationConfigAction, userId string, approve llm.TaskApprovalFunc) (string, error) {
	if action == nil {
		return "", fmt.Errorf("empty action")
	}
	if !r.isAdminUser(userId) {
		return "", fmt.Errorf("configuration actions require an administrator user")
	}
	if action.Operation == "" {
		return "", fmt.Errorf("action type %q is missing the operation field", action.Type)
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
		return r.executeConversationCommandAction(action, approve)
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

		targetKind := fieldValue.Kind()
		if targetKind == reflect.Ptr {
			targetKind = fieldValue.Type().Elem().Kind()
		}
		value = coerceConfigValue(value, targetKind)
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

// coerceConfigValue repairs scalar type mismatches that models produce
// frequently: booleans and numbers quoted as strings ("false", "10"), and
// string fields receiving unquoted scalars (5, true). Values that do not
// match the target kind are returned unchanged so the decoder reports the
// original error.
func coerceConfigValue(value json.RawMessage, kind reflect.Kind) json.RawMessage {
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 {
		return value
	}

	switch kind {
	case reflect.Bool:
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return value
		}
		switch strings.ToLower(strings.TrimSpace(text)) {
		case "true", "1", "yes", "on":
			return json.RawMessage("true")
		case "false", "0", "no", "off":
			return json.RawMessage("false")
		}
		return value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return value
		}
		text = strings.TrimSpace(text)
		if _, err := strconv.ParseFloat(text, 64); err != nil {
			return value
		}
		return json.RawMessage(text)
	case reflect.String:
		if trimmed[0] == '"' {
			return value // already a JSON string
		}
		var raw interface{}
		if err := json.Unmarshal(trimmed, &raw); err != nil {
			return value
		}
		if data, err := json.Marshal(fmt.Sprintf("%v", raw)); err == nil {
			return data
		}
		return value
	}
	return value
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
	if server.Url == "" && server.Command == "" {
		return "", fmt.Errorf("MCP url or command is required")
	}
	config.McpServers[action.Name] = server
	server.Description = ""
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
	for _, server := range config.McpServers {
		if server != nil {
			server.Description = ""
		}
	}
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
	case "search":
		candidates, err := conf.SearchRemoteSkills(r.Ctx, action.Query)
		if err != nil {
			return "", err
		}
		var output strings.Builder
		output.WriteString(fmt.Sprintf("found %d remote skills:", len(candidates)))
		for index, candidate := range candidates {
			description := candidate.Description
			if description == "" {
				description = "(no description)"
			}
			output.WriteString(fmt.Sprintf("\n%d. %s - %s\n   %s", index+1, candidate.Name, description, candidate.Source))
		}
		return output.String(), nil
	case "install":
		if strings.TrimSpace(action.Path) != "" {
			return conf.InstallRemoteSkill(r.Ctx, strings.TrimSpace(action.Path))
		}
		return conf.SearchAndInstallRemoteSkill(r.Ctx, action.Query)
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

func (r *RobotInfo) executeConversationCommandAction(action *param.ConversationConfigAction, approve llm.TaskApprovalFunc) (string, error) {
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
	// an unapproved command triggers an interactive approval: the user can
	// add it to the allowlist and run it right away
	if !conf.IsCommandAllowed(executable) {
		if approve == nil {
			return "", fmt.Errorf("command %q is not allowed; add it to the startup allowlist (ALLOWED_COMMANDS or -allowed_commands) to permit it", action.Args[0])
		}
		detail := fmt.Sprintf("command %q is not in the allowlist; approving adds it to the allowlist and runs it now: argv %s",
			action.Args[0], strings.Join(action.Args, " "))
		approved, approvalErr := approve("allow_command", action.Args[0], detail)
		if approvalErr != nil {
			return "", approvalErr
		}
		if !approved {
			return "", fmt.Errorf("command %q was not approved by the user", action.Args[0])
		}
		conf.AddAllowedCommand(executable)
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

// errConversationPathOutsideRoots marks paths outside every allowed root so
// callers can offer an interactive approval to add the path as a new root.
var errConversationPathOutsideRoots = errors.New("path must stay inside the allowed file operation roots")

// conversationFileRoots returns the evaluated absolute sandbox roots: the
// configured file root plus the roots approved at runtime.
func conversationFileRoots() ([]string, error) {
	configured := conf.AllowedPathRoots()
	if len(configured) == 0 {
		configured = []string{conversationFileRootOrDefault()}
	}

	roots := make([]string, 0, len(configured))
	for index, root := range configured {
		absoluteRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		// the configured root is created on demand; runtime roots must exist
		if index == 0 {
			if err := os.MkdirAll(absoluteRoot, 0755); err != nil {
				return nil, err
			}
		}
		evaluated, err := filepath.EvalSymlinks(absoluteRoot)
		if err != nil {
			continue
		}
		roots = append(roots, evaluated)
	}
	if len(roots) == 0 {
		return nil, errors.New("no usable file operation root")
	}
	return roots, nil
}

func secureConversationFilePath(path string, mustExist bool) (string, error) {
	roots, err := conversationFileRoots()
	if err != nil {
		return "", err
	}

	absolutePath := filepath.Clean(path)
	if !filepath.IsAbs(absolutePath) {
		absolutePath = filepath.Join(roots[0], absolutePath)
	}
	// resolve symlinked ancestors (e.g. /var -> /private/var) before the
	// prefix match; symlinks inside a root are still rejected below
	evaluatedPath := utils.EvalExistingPathPrefix(absolutePath)
	root := roots[0]
	inside := false
	for _, candidate := range roots {
		if conversationPathInsideRoot(candidate, absolutePath) || conversationPathInsideRoot(candidate, evaluatedPath) {
			root = candidate
			inside = true
			break
		}
	}
	if !inside {
		return "", errConversationPathOutsideRoots
	}
	if evaluatedPath != absolutePath {
		absolutePath = evaluatedPath
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
		return "", errConversationPathOutsideRoots
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

func (r *RobotInfo) executeConversationCronAction(action *param.ConversationConfigAction, userId string) (string, error) {
	switch action.Operation {
	case "create":
		if action.Cron == "" || action.Prompt == "" {
			return "", fmt.Errorf("cron expression and prompt are required")
		}
		cronID, err := r.InsertCron(action.Cron, action.Prompt)
		if err != nil {
			return "", err
		}
		// the id is what later delete/update calls need, so the model must see it
		return fmt.Sprintf("created cron task %d (cron: %s, prompt: %s)", cronID, action.Cron, action.Prompt), nil
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
