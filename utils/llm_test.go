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

func TestGetAvailTxtType_IncludesSynthorai(t *testing.T) {
	conf.BaseConfInfo.SynthoraiToken = "sk-synthorai-test"
	defer func() { conf.BaseConfInfo.SynthoraiToken = "" }()
	got := GetAvailTxtType()
	assert.Contains(t, got, param.Synthorai)
}

func TestGetTxtModel_SynthoraiDefault(t *testing.T) {
	assert.Equal(t, param.SynthoraiDeepSeekV4Flash, GetTxtModel(param.Synthorai))
}

func TestGetTxtType_SynthoraiFallback(t *testing.T) {
	conf.BaseConfInfo.Type = param.Synthorai
	assert.Equal(t, param.Synthorai, GetTxtType(nil))
}
