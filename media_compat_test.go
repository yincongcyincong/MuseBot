package main

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/robot"
	"github.com/yincongcyincong/MuseBot/utils"
)

func newMediaTestRobot(t *testing.T) *robot.RobotInfo {
	t.Helper()
	rec := httptest.NewRecorder()
	web := robot.NewWeb("", 0, "media_compat_user", "", "", "", nil, rec, rec)
	return web.Robot
}

func TestGetLastImageContentCompat(t *testing.T) {
	r := newMediaTestRobot(t)
	ctx := context.Background()

	// legacy row: base64 data uri still decodes
	legacyBytes := []byte("legacy-image-bytes")
	legacyURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(legacyBytes)
	_, err := db.InsertRecordInfo(ctx, &db.Record{
		UserId:     "media_compat_user",
		Answer:     legacyURI,
		RecordType: param.ImageRecordType,
	})
	assert.NoError(t, err)

	got, err := r.GetLastImageContent()
	assert.NoError(t, err)
	assert.Equal(t, legacyBytes, got)

	// new row: local media path wins as the latest record
	newBytes := []byte("file-image-bytes")
	mediaPath, err := utils.SaveMediaFile("web_media_compat_user", "", newBytes, "png")
	assert.NoError(t, err)
	_, err = db.InsertRecordInfo(ctx, &db.Record{
		UserId:     "media_compat_user",
		Answer:     mediaPath,
		RecordType: param.ImageRecordType,
	})
	assert.NoError(t, err)

	got, err = r.GetLastImageContent()
	assert.NoError(t, err)
	assert.Equal(t, newBytes, got)
}
