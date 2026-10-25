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

// an empty table gets the initial admin account with the default admin/admin
// password; the UI prompts for a change after login
func TestEnsureAdminUserCreatesDefaultAdmin(t *testing.T) {
	DeleteAllUserData()
	defer DeleteAllUserData()

	if err := ensureAdminUser(); err != nil {
		t.Fatalf("ensureAdminUser failed: %v", err)
	}

	u, err := GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if ok, isLegacy := VerifyPassword(DefaultAdminPassword, u.Password); !ok || isLegacy {
		t.Errorf("expected default admin password to verify as bcrypt, ok=%v isLegacy=%v", ok, isLegacy)
	}
	if !HasDefaultPassword(u) {
		t.Error("expected HasDefaultPassword true for the initial account")
	}

	// a second run must not create a duplicate
	if err := ensureAdminUser(); err != nil {
		t.Fatalf("second ensureAdminUser failed: %v", err)
	}
	_, total, err := ListUsers(0, 10, "admin")
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected exactly 1 admin account, got %d", total)
	}
}

func TestHasDefaultPassword(t *testing.T) {
	DeleteAllUserData()
	defer DeleteAllUserData()

	// legacy unsalted MD5 admin/admin hashes still count as the default
	if _, err := DB.Exec(`INSERT INTO admin_users (username, password, create_time, update_time) VALUES ('legacy', ?, 0, 0)`,
		utils.MD5(DefaultAdminPassword)); err != nil {
		t.Fatalf("insert legacy admin failed: %v", err)
	}
	u, err := GetUserByUsername("legacy")
	if err != nil {
		t.Fatalf("get legacy admin failed: %v", err)
	}
	if !HasDefaultPassword(u) {
		t.Error("expected legacy MD5 default hash to count as default password")
	}

	// changing the password clears the flag
	if err := UpdateUserPassword(u.ID, "changed"); err != nil {
		t.Fatalf("update password failed: %v", err)
	}
	u, err = GetUserByUsername("legacy")
	if err != nil {
		t.Fatalf("get user after update failed: %v", err)
	}
	if HasDefaultPassword(u) {
		t.Error("expected HasDefaultPassword false after password change")
	}
}
