package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestGetAvailTxtType_IncludesOrcaRouter(t *testing.T) {
	conf.BaseConfInfo.OrcaRouterToken = "sk-orca-test"
	defer func() { conf.BaseConfInfo.OrcaRouterToken = "" }()
	got := GetAvailTxtType()
	assert.Contains(t, got, param.OrcaRouter)
}

func TestGetAvailImgType_IncludesOrcaRouter(t *testing.T) {
	conf.BaseConfInfo.OrcaRouterToken = "sk-orca-test"
	defer func() { conf.BaseConfInfo.OrcaRouterToken = "" }()
	got := GetAvailImgType()
	assert.Contains(t, got, param.OrcaRouter)
}

func TestGetTxtModel_OrcaRouterDefault(t *testing.T) {
	assert.Equal(t, param.OrcaRouterAuto, GetTxtModel(param.OrcaRouter))
}

func TestGetTxtType_OrcaRouterFallback(t *testing.T) {
	conf.BaseConfInfo.Type = param.OrcaRouter
	assert.Equal(t, param.OrcaRouter, GetTxtType(nil))
}
