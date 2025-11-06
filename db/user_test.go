package db

import (
	"testing"
	
	_ "github.com/mattn/go-sqlite3"
	"github.com/yincongcyincong/MuseBot/conf"
)

func TestInsertAndGetUser(t *testing.T) {
	conf.BaseConfInfo.TokenPerUser = new(int)
	*conf.BaseConfInfo.TokenPerUser = 100
	
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
