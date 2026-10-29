package conf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitConfig_WithEnv(t *testing.T) {

	os.Setenv("DB_TYPE", "mysql")
	os.Setenv("DB_CONF", "root:pass@tcp(localhost:3306)/muse")
	os.Setenv("SESSION_KEY", "env_session_key_0123456789abcdef")
	os.Setenv("ADMIN_PORT", "8080")
	os.Setenv("CHECK_BOT_SEC", "99")

	os.Setenv("REGISTER_TYPE", "etcd")
	os.Setenv("ETCD_URLS", "http://127.0.0.1:2379,http://127.0.0.2:2379")
	os.Setenv("ETCD_USERNAME", "admin")
	os.Setenv("ETCD_PASSWORD", "123456")

	InitConfig()

	if BaseConfInfo.DBType != "mysql" {
		t.Errorf("expected DBType = mysql, got %s", BaseConfInfo.DBType)
	}

	if BaseConfInfo.DBConf != "root:pass@tcp(localhost:3306)/muse" {
		t.Errorf("expected DBConf = mysql conf, got %s", BaseConfInfo.DBConf)
	}

	if BaseConfInfo.SessionKey != "env_session_key_0123456789abcdef" {
		t.Errorf("expected SessionKey = env_session_key_0123456789abcdef, got %s", BaseConfInfo.SessionKey)
	}

	if BaseConfInfo.AdminPort != "8080" {
		t.Errorf("expected AdminPort = 8080, got %s", BaseConfInfo.AdminPort)
	}

	if BaseConfInfo.CheckBotSec != 99 {
		t.Errorf("expected CheckBotSec = 99, got %d", BaseConfInfo.CheckBotSec)
	}

	if RegisterConfInfo.Type != "etcd" {
		t.Errorf("expected Type = etcd, got %s", RegisterConfInfo.Type)
	}

	expectedURLs := []string{"http://127.0.0.1:2379", "http://127.0.0.2:2379"}
	if len(RegisterConfInfo.EtcdURLs) != len(expectedURLs) {
		t.Fatalf("expected %d EtcdURLs, got %d", len(expectedURLs), len(RegisterConfInfo.EtcdURLs))
	}
	for i, u := range expectedURLs {
		if RegisterConfInfo.EtcdURLs[i] != u {
			t.Errorf("expected EtcdURLs[%d] = %s, got %s", i, u, RegisterConfInfo.EtcdURLs[i])
		}
	}

	if RegisterConfInfo.EtcdUsername != "admin" {
		t.Errorf("expected EtcdUsername = admin, got %s", RegisterConfInfo.EtcdUsername)
	}

	if RegisterConfInfo.EtcdPassword != "123456" {
		t.Errorf("expected EtcdPassword = 123456, got %s", RegisterConfInfo.EtcdPassword)
	}

	os.Clearenv()
}

func TestEnsureSessionKeyAt(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "admin_session_key")

	BaseConfInfo.SessionKey = ""
	source := ensureSessionKeyAt(keyPath)
	if BaseConfInfo.SessionKey == "" || len(BaseConfInfo.SessionKey) < 32 {
		t.Fatalf("expected a generated session key, got %q", BaseConfInfo.SessionKey)
	}
	if !strings.Contains(source, "generated") {
		t.Errorf("expected generated source, got %q", source)
	}
	firstKey := BaseConfInfo.SessionKey

	// restart: the persisted key is reloaded unchanged
	BaseConfInfo.SessionKey = ""
	source = ensureSessionKeyAt(keyPath)
	if BaseConfInfo.SessionKey != firstKey {
		t.Errorf("expected persisted key %q to be reloaded, got %q", firstKey, BaseConfInfo.SessionKey)
	}
	if !strings.Contains(source, "loaded") {
		t.Errorf("expected loaded source, got %q", source)
	}

	// the legacy hard-coded default must be replaced
	BaseConfInfo.SessionKey = legacySessionKeyDefault
	legacyPath := filepath.Join(t.TempDir(), "admin_session_key")
	ensureSessionKeyAt(legacyPath)
	if BaseConfInfo.SessionKey == legacySessionKeyDefault {
		t.Error("expected the legacy default session key to be replaced")
	}
}

func TestValidateSessionKey(t *testing.T) {
	// the legacy public default is rejected outright
	if err := validateSessionKey(legacySessionKeyDefault); err == nil {
		t.Error("expected the legacy default session key to be rejected")
	}

	// short keys are forgeable material and must be rejected
	for _, key := range []string{"", "abc", "0123456789abcdef0123456789abcde"} {
		if err := validateSessionKey(key); err == nil {
			t.Errorf("expected short session key %q to be rejected", key)
		}
	}

	// 32 or more characters pass
	if err := validateSessionKey("0123456789abcdef0123456789abcdef"); err != nil {
		t.Errorf("expected a 32-char session key to pass, got %v", err)
	}
}
