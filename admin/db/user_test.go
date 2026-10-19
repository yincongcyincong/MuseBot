package db

import (
	"testing"
	"time"

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
	if ok, isLegacy := VerifyPassword("123456", u.Password); !ok || isLegacy {
		t.Errorf("expected bcrypt password to verify, ok=%v isLegacy=%v", ok, isLegacy)
	}
	if ok, _ := VerifyPassword("wrongpass", u.Password); ok {
		t.Error("expected wrong password to fail")
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
	if ok, isLegacy := VerifyPassword("newpass", u3.Password); !ok || isLegacy {
		t.Errorf("expected updated bcrypt password to verify, ok=%v isLegacy=%v", ok, isLegacy)
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
	if users[0].Password != "" {
		t.Error("expected ListUsers to not expose the password hash")
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

// legacy unsalted MD5 hashes must keep verifying (transparent bcrypt upgrade)
func TestVerifyPasswordLegacyMD5(t *testing.T) {
	ok, isLegacy := VerifyPassword("admin", utils.MD5("admin"))
	if !ok || !isLegacy {
		t.Errorf("expected legacy MD5 hash to verify, ok=%v isLegacy=%v", ok, isLegacy)
	}

	ok, _ = VerifyPassword("wrong", utils.MD5("admin"))
	if ok {
		t.Error("expected wrong password to fail against legacy hash")
	}
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

// accounts still carrying the public admin/admin hash must be reset on startup
func TestEnsureAdminUserResetsDefaultPassword(t *testing.T) {
	DeleteAllUserData()
	defer DeleteAllUserData()

	now := time.Now().Unix()
	_, err := DB.Exec(`INSERT INTO admin_users (username, password, create_time, update_time) VALUES ('admin', ?, ?, ?)`,
		DefaultAdminMD5Password, now, now)
	if err != nil {
		t.Fatalf("insert legacy admin failed: %v", err)
	}

	if err := ensureAdminUser(); err != nil {
		t.Fatalf("ensureAdminUser failed: %v", err)
	}

	u, err := GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if u.Password == DefaultAdminMD5Password {
		t.Fatal("expected the default admin/admin hash to be reset")
	}
	// the default password must no longer verify
	if ok, _ := VerifyPassword("admin", u.Password); ok {
		t.Error("expected admin/admin to stop working after the reset")
	}
}
