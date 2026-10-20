package db

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestInsertAndGetUser(t *testing.T) {
	conf.BaseConfInfo.TokenPerUser = 100

	userId := "123456789"
	mode := `{"txt_type":"gemini","txt_model":"gemini-2.0-flash","img_type":"gemini","img_model":"gemini-2.0-flash-preview-image-generation","video_type":"gemini","video_model":"veo-2.0-generate-001"}`

	// 插入用户
	id, err := InsertUser(userId, mode)
	if err != nil {
		t.Fatalf("InsertUser failed: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non-zero ID")
	}

	// 获取用户
	user, err := GetUserByID(userId)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if user == nil {
		t.Fatalf("user not found")
	}
	if user.UserId != userId || user.LLMConfig != mode || user.AvailToken != 100 {
		t.Errorf("unexpected user data: %+v", user)
	}

	users, err := GetUsers()
	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("user not found")
	}

	err = UpdateUserLLMConfig(user.UserId, `{"txt_type":"gemini","txt_model":"gemini-2.0-flash","img_type":"gemini","img_model":"gemini-2.0-flash-preview-image-generation","video_type":"gemini","video_model":"veo-2.0-generate-001"}`)
	if err != nil {
		t.Fatalf("UpdateUserMode failed: %v", err)
	}

	err = AddAvailToken(user.UserId, 1000)
	if err != nil {
		t.Fatalf("UpdateUserUpdateTime failed: %v", err)
	}

	err = AddToken(user.UserId, 1000)
	if err != nil {
		t.Fatalf("AddToken failed: %v", err)
	}

	user, err = GetUserByID(userId)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if user.UserId != userId || user.LLMConfig != `{"txt_type":"gemini","txt_model":"gemini-2.0-flash","img_type":"gemini","img_model":"gemini-2.0-flash-preview-image-generation","video_type":"gemini","video_model":"veo-2.0-generate-001"}` || user.Token != 1000 || user.AvailToken != 1100 {
		t.Errorf("unexpected user data: %+v", user)
	}

}

func TestSyncUserTxtConfig(t *testing.T) {
	conf.BaseConfInfo.DeepseekToken = "sk-test"
	defer func() { conf.BaseConfInfo.DeepseekToken = "" }()

	// non-json config rows must not abort the sync
	_, err := InsertUser("sync-user-bad", "default")
	assert.NoError(t, err)

	// stored type no longer has a token: falls back to the configured type
	// and its default model
	_, err = InsertUser("sync-user-1", `{"txt_type":"openai","txt_model":"deepseek-chat"}`)
	assert.NoError(t, err)

	// stored type still has a token: untouched, even with a model natively
	// belonging to another provider (relay setups)
	_, err = InsertUser("sync-user-2", `{"txt_type":"deepseek","txt_model":"gpt-3.5-turbo"}`)
	assert.NoError(t, err)

	SyncUserTxtConfig()

	user1, err := GetUserByID("sync-user-1")
	assert.NoError(t, err)
	assert.Equal(t, param.DeepSeek, user1.LLMConfigRaw.TxtType)
	assert.Equal(t, "deepseek-chat", user1.LLMConfigRaw.TxtModel)

	user2, err := GetUserByID("sync-user-2")
	assert.NoError(t, err)
	assert.Equal(t, param.DeepSeek, user2.LLMConfigRaw.TxtType)
	assert.Equal(t, "gpt-3.5-turbo", user2.LLMConfigRaw.TxtModel)
}
