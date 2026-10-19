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

func TestGetAvailTxtType_IncludesCheaperInference(t *testing.T) {
	conf.BaseConfInfo.CheaperInferenceToken = "ci_live_test"
	defer func() { conf.BaseConfInfo.CheaperInferenceToken = "" }()
	got := GetAvailTxtType()
	assert.Contains(t, got, param.CheaperInference)
}

func TestGetAvailImgType_IncludesCheaperInference(t *testing.T) {
	conf.BaseConfInfo.CheaperInferenceToken = "ci_live_test"
	defer func() { conf.BaseConfInfo.CheaperInferenceToken = "" }()
	got := GetAvailImgType()
	assert.Contains(t, got, param.CheaperInference)
}

func TestGetTxtModel_CheaperInferenceDefault(t *testing.T) {
	assert.Equal(t, param.CheaperInferenceGPT54Mini, GetTxtModel(param.CheaperInference))
}

func TestGetTxtType_CheaperInferenceFallback(t *testing.T) {
	conf.BaseConfInfo.Type = param.CheaperInference
	assert.Equal(t, param.CheaperInference, GetTxtType(nil))
}
