package db

import (
	"testing"
	
	"github.com/yincongcyincong/MuseBot/utils"
)

func TestUserCRUD_SQLite(t *testing.T) {
	DeleteAllUserData()
	
	// 1️⃣ CreateUser
	err := CreateUser("alice", "123456")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	
	// 2️⃣ GetUserByUsername
	u, err := GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if u.Username != "alice" {
		t.Errorf("expected username 'alice', got %s", u.Username)
	}
	
	// 3️⃣ GetUserByID
	u2, err := GetUserByID(u.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if u2.Username != "alice" {
		t.Errorf("expected username 'alice', got %s", u2.Username)
	}
	
	// 4️⃣ UpdateUserPassword
	err = UpdateUserPassword(u.ID, "newpass")
	if err != nil {
		t.Fatalf("UpdateUserPassword failed: %v", err)
	}
	
	// 验证密码是否真的变了
	u3, err := GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("GetUserByUsername failed after update: %v", err)
	}
	if u3.Password != utils.MD5("newpass") {
		t.Errorf("expected password %s, got %s", utils.MD5("newpass"), u3.Password)
	}
	
	// 5️⃣ ListUsers
	users, total, err := ListUsers(0, 10, "")
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1 user, got %d", total)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user in result, got %d", len(users))
	}
	
	// 6️⃣ DeleteUser
	err = DeleteUser(u.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
	
	_, err = GetUserByID(u.ID)
	if err == nil {
		t.Errorf("expected error after DeleteUser, got nil")
	}
	
	DeleteAllUserData()
}

func TestListUsers_Filter_SQLite(t *testing.T) {
	
	CreateUser("bob", "111")
	CreateUser("bobby", "222")
	CreateUser("alice", "333")
	
	users, total, err := ListUsers(0, 10, "bob")
	if err != nil {
		t.Fatalf("ListUsers with filter failed: %v", err)
	}
	
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	
	found := false
	for _, u := range users {
		if u.Username == "bobby" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected to find user 'bobby'")
	}
	
	DeleteAllUserData()
}
