package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/yincongcyincong/MuseBot/admin/conf"
	"github.com/yincongcyincong/MuseBot/admin/db"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestMain(m *testing.M) {
	conf.InitConfig()
	db.InitTable()
	InitSessionStore()

	code := m.Run()
	os.Exit(code)
}

// testRequest builds a request with the context values the real
// WithRequestContext middleware injects (utils.Failure reads start_time).
func testRequest(url string) *http.Request {
	r := httptest.NewRequest("GET", url, nil)
	ctx := context.WithValue(r.Context(), "log_id", "test")
	ctx = context.WithValue(ctx, "start_time", time.Now())
	return r.WithContext(ctx)
}

// signedRequest returns a request carrying a validly-signed session cookie
// with the given user_id, as a leaked-key attacker would forge. An optional
// username mirrors what real login sessions store.
func signedRequest(t *testing.T, userID int, username ...string) *http.Request {
	t.Helper()
	r := testRequest("/user/get")
	w := httptest.NewRecorder()
	session, err := sessionStore.New(r, sessionName)
	if err != nil {
		t.Fatalf("new session fail: %v", err)
	}
	session.Values["user_id"] = userID
	if len(username) > 0 {
		session.Values["username"] = username[0]
	}
	if err := session.Save(r, w); err != nil {
		t.Fatalf("save session fail: %v", err)
	}
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == sessionName {
			r.AddCookie(cookie)
			return r
		}
	}
	t.Fatal("session cookie not set")
	return nil
}

func TestRequireLogin(t *testing.T) {
	called := false
	handler := RequireLogin(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	// no cookie at all is rejected
	called = false
	handler(httptest.NewRecorder(), testRequest("/user/get"))
	if called {
		t.Error("request without session cookie must be rejected")
	}

	// a validly-signed cookie for an existing user passes
	if err := db.CreateUser("require_login_user", "password123"); err != nil {
		t.Fatalf("create user fail: %v", err)
	}
	u, err := db.GetUserByUsername("require_login_user")
	if err != nil {
		t.Fatalf("get user fail: %v", err)
	}
	defer func() { _ = db.DeleteUser(u.ID) }()

	called = false
	handler(httptest.NewRecorder(), signedRequest(t, u.ID))
	if !called {
		t.Error("valid session for existing user must pass")
	}

	// a validly-signed cookie for a deleted user is rejected: the session key
	// alone must not grant access to arbitrary user ids
	if err := db.DeleteUser(u.ID); err != nil {
		t.Fatalf("delete user fail: %v", err)
	}
	called = false
	handler(httptest.NewRecorder(), signedRequest(t, u.ID))
	if called {
		t.Error("signed session for deleted user must be rejected")
	}

	// so must a signed session pointing at a user id that never existed
	called = false
	handler(httptest.NewRecorder(), signedRequest(t, 999999))
	if called {
		t.Error("signed session for unknown user id must be rejected")
	}
}

func TestSessionCookieFlags(t *testing.T) {
	r := testRequest("/user/get")
	w := httptest.NewRecorder()
	session, err := sessionStore.New(r, sessionName)
	if err != nil {
		t.Fatalf("new session fail: %v", err)
	}
	session.Values["user_id"] = 1
	if err := session.Save(r, w); err != nil {
		t.Fatalf("save session fail: %v", err)
	}

	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == sessionName {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("session cookie not set")
	}
	if !cookie.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}
	if conf.BaseConfInfo.SessionSecure && !cookie.Secure {
		t.Error("session cookie must be Secure when session_secure is enabled")
	}
	if cookie.Path != "/" {
		t.Errorf("session cookie path must be /, got %s", cookie.Path)
	}
}

// GetCurrentUserHandler flags accounts still on the default admin/admin
// password so the UI can prompt for a change.
func TestCurrentUserFlagsDefaultPassword(t *testing.T) {
	if err := db.CreateUser("default_pwd_user", "admin"); err != nil {
		t.Fatalf("create user fail: %v", err)
	}
	u, err := db.GetUserByUsername("default_pwd_user")
	if err != nil {
		t.Fatalf("get user fail: %v", err)
	}
	defer func() { _ = db.DeleteUser(u.ID) }()

	w := httptest.NewRecorder()
	GetCurrentUserHandler(w, signedRequest(t, u.ID, "default_pwd_user"))
	res := decodeResponse(w)
	if res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("current user fail: %+v", res)
	}
	data, _ := res.Data.(map[string]interface{})
	if data == nil {
		t.Fatalf("unexpected payload: %+v", res.Data)
	}
	if data["is_default_password"] != true {
		t.Errorf("expected is_default_password true for the default password, got %v", data["is_default_password"])
	}

	// changing the password clears the flag
	if err := db.UpdateUserPassword(u.ID, "new-password"); err != nil {
		t.Fatalf("update password fail: %v", err)
	}
	w = httptest.NewRecorder()
	GetCurrentUserHandler(w, signedRequest(t, u.ID, "default_pwd_user"))
	res = decodeResponse(w)
	if res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("current user after update fail: %+v", res)
	}
	data, _ = res.Data.(map[string]interface{})
	if data["is_default_password"] != false {
		t.Errorf("expected is_default_password false after password change, got %v", data["is_default_password"])
	}
}
