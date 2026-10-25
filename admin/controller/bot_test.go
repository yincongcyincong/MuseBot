package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yincongcyincong/MuseBot/admin/db"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

// fakeBotBackend stands in for a running MuseBot instance so handlers that
// proxy to the bot (command/get probing) can be tested offline.
func fakeBotBackend(t *testing.T, command string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/command/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"code":0,"message":"success","data":%q}`, command)))
	})
	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func cleanupBot(t *testing.T, name string) {
	t.Helper()
	t.Cleanup(func() {
		// db.ListBots filters by address, so match bots by name here
		bots, _, err := db.ListBots(0, 1000, "")
		if err != nil {
			return
		}
		for _, b := range bots {
			if b.Name == name {
				_ = db.SoftDeleteBot(b.ID)
			}
		}
	})
}

// findBotByName scans unfiltered listings; db.ListBots can only filter by
// address substring, which is rarely what tests want.
func findBotByName(t *testing.T, name string) *db.Bot {
	t.Helper()
	bots, _, err := db.ListBots(0, 1000, "")
	if err != nil {
		t.Fatalf("list bots fail: %v", err)
	}
	for _, b := range bots {
		if b.Name == name {
			return b
		}
	}
	return nil
}

func TestCreateBotValidation(t *testing.T) {
	// missing bot_name / http_host in the command is rejected before any probe
	w := httptest.NewRecorder()
	CreateBot(w, postJSON("/bot/create", `{"command":"./MuseBot -http_host=:36060"}`))
	if res := decodeResponse(w); res == nil || res.Code == param.CodeSuccess {
		t.Errorf("expected command without bot_name to fail, got %+v", res)
	}

	w = httptest.NewRecorder()
	CreateBot(w, postJSON("/bot/create", `{"command":"./MuseBot -bot_name=x"}`))
	if res := decodeResponse(w); res == nil || res.Code == param.CodeSuccess {
		t.Errorf("expected command without http_host to fail, got %+v", res)
	}

	// malformed json body
	w = httptest.NewRecorder()
	CreateBot(w, postJSON("/bot/create", `{not json`))
	if res := decodeResponse(w); res == nil || res.Code == param.CodeSuccess {
		t.Errorf("expected malformed body to fail, got %+v", res)
	}
}

func TestCreateBotWithBackendProbe(t *testing.T) {
	backendCommand := "./MuseBot -bot_name=probebot -http_host=127.0.0.1:1"
	srv := fakeBotBackend(t, backendCommand)
	cleanupBot(t, "probebot")

	// http_host points at the fake backend so the create-time probe succeeds
	// and the command is refreshed from the running bot
	host := srv.URL[len("http://"):] // 127.0.0.1:port
	w := httptest.NewRecorder()
	CreateBot(w, postJSON("/bot/create",
		`{"command":"./MuseBot -bot_name=probebot -http_host=`+host+`"}`))
	if res := decodeResponse(w); res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("create bot fail: %+v", res)
	}

	bot := findBotByName(t, "probebot")
	if bot == nil {
		t.Fatal("created bot not found")
	}
	if bot.Address != utils.NormalizeHTTP(host) {
		t.Errorf("expected address %q, got %q", utils.NormalizeHTTP(host), bot.Address)
	}
	if bot.Command != backendCommand {
		t.Errorf("expected probed command %q, got %q", backendCommand, bot.Command)
	}

	// GetBot handler returns the stored bot
	w = httptest.NewRecorder()
	GetBot(w, getQuery("/bot/get?id="+itoa(bot.ID)))
	res := decodeResponse(w)
	if res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("get bot fail: %+v", res)
	}
}

func TestCreateBotOfflineBackend(t *testing.T) {
	cleanupBot(t, "offlinebot")

	// port 1 refuses connections: probe fails but creation still succeeds
	w := httptest.NewRecorder()
	CreateBot(w, postJSON("/bot/create",
		`{"command":"./MuseBot -bot_name=offlinebot -http_host=127.0.0.1:1"}`))
	if res := decodeResponse(w); res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("create bot with unreachable backend fail: %+v", res)
	}
	bot := findBotByName(t, "offlinebot")
	if bot == nil {
		t.Fatal("expected offline bot to be created")
	}
	if bot.Command == "" {
		t.Error("expected original command to be stored when probe fails")
	}
}

func TestUpdateBotAddressAndSoftDelete(t *testing.T) {
	cleanupBot(t, "movebot")

	// seed directly to keep the update test independent of create
	if err := db.CreateBot("http://127.0.0.1:1", "movebot", "", "", "",
		"./MuseBot -bot_name=movebot -http_host=127.0.0.1:1"); err != nil {
		t.Fatalf("seed bot fail: %v", err)
	}
	seeded := findBotByName(t, "movebot")
	if seeded == nil {
		t.Fatal("seed bot missing")
	}
	id := seeded.ID

	// command is non-empty so the new-address probe is skipped
	w := httptest.NewRecorder()
	UpdateBotAddress(w, postJSON("/bot/update", `{"id":`+itoa(id)+
		`,"command":"./MuseBot -bot_name=movebot -http_host=127.0.0.1:2"}`))
	if res := decodeResponse(w); res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("update bot fail: %+v", res)
	}
	fresh, err := db.GetBotByID(itoa(id))
	if err != nil {
		t.Fatalf("get bot after update fail: %v", err)
	}
	if fresh.Address != "http://127.0.0.1:2" {
		t.Errorf("expected updated address, got %q", fresh.Address)
	}

	// updating a missing bot fails
	w = httptest.NewRecorder()
	UpdateBotAddress(w, postJSON("/bot/update",
		`{"id":999999,"command":"./MuseBot -bot_name=x -http_host=127.0.0.1:3"}`))
	if res := decodeResponse(w); res == nil || res.Code == param.CodeSuccess {
		t.Errorf("expected update of missing bot to fail, got %+v", res)
	}

	// soft delete hides the bot from listings
	w = httptest.NewRecorder()
	SoftDeleteBot(w, getQuery("/bot/delete?id="+itoa(id)))
	if res := decodeResponse(w); res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("soft delete fail: %+v", res)
	}
	_, total, err := db.ListBots(0, 10, "movebot")
	if err != nil {
		t.Fatalf("list after delete fail: %v", err)
	}
	if total != 0 {
		t.Errorf("expected deleted bot hidden from list, total=%d", total)
	}
}

func TestGetBotCommandViaBackend(t *testing.T) {
	command := "./MuseBot -bot_name=cmdbot -http_host=127.0.0.1:9"
	srv := fakeBotBackend(t, command)
	cleanupBot(t, "cmdbot")

	host := srv.URL[len("http://"):]
	if err := db.CreateBot(utils.NormalizeHTTP(host), "cmdbot", "", "", "", command); err != nil {
		t.Fatalf("seed bot fail: %v", err)
	}
	seeded := findBotByName(t, "cmdbot")
	if seeded == nil {
		t.Fatal("seed bot missing")
	}

	w := httptest.NewRecorder()
	GetBotCommand(w, getQuery("/bot/command/get?id="+itoa(seeded.ID)))
	res := decodeResponse(w)
	if res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("get bot command fail: %+v", res)
	}
	if res.Data != command {
		t.Errorf("expected command %q from backend, got %v", command, res.Data)
	}

	fresh, err := db.GetBotByID(itoa(seeded.ID))
	if err != nil {
		t.Fatalf("reload bot fail: %v", err)
	}
	if fresh.Command != command {
		t.Errorf("expected command persisted, got %q", fresh.Command)
	}
}

func TestListBotsPagination(t *testing.T) {
	cleanupBot(t, "pagebota")
	cleanupBot(t, "pagebotb")
	// addresses share a distinctive substring the filter can match on
	_ = db.CreateBot("http://127.0.0.1:23001", "pagebota", "", "", "", "./MuseBot -bot_name=pagebota -http_host=127.0.0.1:23001")
	_ = db.CreateBot("http://127.0.0.1:23002", "pagebotb", "", "", "", "./MuseBot -bot_name=pagebotb -http_host=127.0.0.1:23002")

	w := httptest.NewRecorder()
	ListBots(w, getQuery("/bot/list?page=1&page_size=1&address=127.0.0.1:230"))
	res := decodeResponse(w)
	if res == nil || res.Code != param.CodeSuccess {
		t.Fatalf("list bots fail: %+v", res)
	}
	data, _ := res.Data.(map[string]interface{})
	if data == nil {
		t.Fatalf("unexpected list payload: %+v", res.Data)
	}
	if data["total"].(float64) != 2 {
		t.Errorf("expected 2 matching bots, got %v", data["total"])
	}
	if list, ok := data["list"].([]interface{}); !ok || len(list) != 1 {
		t.Errorf("expected page_size=1 to return 1 bot, got %v", data["list"])
	}
	if data["is_register"] != false {
		t.Errorf("expected is_register false in db mode, got %v", data["is_register"])
	}
}
