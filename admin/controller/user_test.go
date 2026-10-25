package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/yincongcyincong/MuseBot/admin/db"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

func postJSON(url, body string) *http.Request {
	r := testRequest(url)
	r.Method = http.MethodPost
	r.Body = io.NopCloser(strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func getQuery(url string) *http.Request {
	return testRequest(url)
}

func decodeResponse(w *httptest.ResponseRecorder) *utils.Response {
	res := new(utils.Response)
	if err := json.NewDecoder(w.Body).Decode(res); err != nil {
		return nil
	}
	return res
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func TestUserHandlerLifecycle(t *testing.T) {
	name := "handler_user"
	defer func() {
		if u, err := db.GetUserByUsername(name); err == nil {
			_ = db.DeleteUser(u.ID)
		}
	}()

	// create
	w := httptest.NewRecorder()
	CreateUser(w, postJSON("/user/create", `{"username":"`+name+`","password":"pass123"}`))
	if res := decodeResponse(w); res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("create user fail: %+v", res)
	}

	// get by id
	u, err := db.GetUserByUsername(name)
	if err != nil {
		t.Fatalf("seed user not found: %v", err)
	}
	w = httptest.NewRecorder()
	GetUser(w, getQuery("/user/get?id="+itoa(u.ID)))
	res := decodeResponse(w)
	if res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("get user fail: %+v", res)
	}
	data, _ := res.Data.(map[string]interface{})
	if data == nil || data["username"] != name {
		t.Errorf("expected username %q in response, got %+v", name, res.Data)
	}

	// invalid id is rejected
	w = httptest.NewRecorder()
	GetUser(w, getQuery("/user/get?id=abc"))
	if res := decodeResponse(w); res == nil || res.Code == param.CodeSuccess {
		t.Errorf("expected invalid id to fail, got %+v", res)
	}

	// list with username filter
	w = httptest.NewRecorder()
	ListUsers(w, getQuery("/user/list?username="+name))
	res = decodeResponse(w)
	if res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("list users fail: %+v", res)
	}
	listData, _ := res.Data.(map[string]interface{})
	if listData == nil || listData["total"].(float64) < 1 {
		t.Errorf("expected at least one user, got %+v", res.Data)
	}

	// update password and verify the hash really changed
	w = httptest.NewRecorder()
	UpdateUserPassword(w, postJSON("/user/update", `{"id":`+itoa(u.ID)+`,"password":"newpass"}`))
	if res := decodeResponse(w); res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("update password fail: %+v", res)
	}
	fresh, err := db.GetUserByUsername(name)
	if err != nil {
		t.Fatalf("get user after update fail: %v", err)
	}
	if ok, _ := db.VerifyPassword("newpass", fresh.Password); !ok {
		t.Error("expected updated password to verify")
	}

	// delete
	w = httptest.NewRecorder()
	DeleteUser(w, getQuery("/user/delete?id="+itoa(u.ID)))
	if res := decodeResponse(w); res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("delete user fail: %+v", res)
	}
	if _, err := db.GetUserByID(u.ID); err == nil {
		t.Error("expected deleted user to be gone")
	}

	// deleting a missing id is idempotent: success without error
	w = httptest.NewRecorder()
	DeleteUser(w, getQuery("/user/delete?id=999999"))
	if res := decodeResponse(w); res == nil || res.Code != param.CodeSuccess {
		t.Errorf("expected idempotent delete of missing user, got %+v", res)
	}
}

func TestParsePaginationParams(t *testing.T) {
	cases := []struct {
		query    string
		page     int
		pageSize int
	}{
		{"", 1, 10},
		{"?page=0&page_size=-5", 1, 10},
		{"?page=3&page_size=20", 3, 20},
		{"?page=2&page_size=500", 2, 10}, // over cap falls back to 10
		{"?page=abc&page_size=xyz", 1, 10},
	}
	for _, c := range cases {
		r := getQuery("/user/list" + c.query)
		page, pageSize := parsePaginationParams(r)
		if page != c.page || pageSize != c.pageSize {
			t.Errorf("query %q: expected (%d,%d), got (%d,%d)", c.query, c.page, c.pageSize, page, pageSize)
		}
	}
}
