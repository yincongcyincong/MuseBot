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
	mode := "default"
	
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
	if user.UserId != userId || user.Mode != mode || user.AvailToken != 100 {
		t.Errorf("unexpected user data: %+v", user)
	}
	
	users, err := GetUsers()
	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("user not found")
	}
	
	err = UpdateUserMode(user.UserId, "mode")
	if err != nil {
		t.Fatalf("UpdateUserMode failed: %v", err)
	}
	
	err = UpdateUserUpdateTime(user.UserId, 111)
	if err != nil {
		t.Fatalf("UpdateUserUpdateTime failed: %v", err)
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
	if user.UserId != userId || user.Mode != "mode" || user.Token != 1000 || user.AvailToken != 1100 {
		t.Errorf("unexpected user data: %+v", user)
	}
	
}
