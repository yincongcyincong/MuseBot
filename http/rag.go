package http

import (
	"io"
	"net/http"
	"os"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type RagFile struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

func CreateRagFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ragFile := &RagFile{}
	err := utils.HandleJsonBody(r, ragFile)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	path := *conf.RagConfInfo.KnowledgePath + "/" + ragFile.FileName
	err = os.WriteFile(path, []byte(ragFile.Content), 0644)
	if err != nil {
		logger.ErrorCtx(ctx, "write file error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	utils.Success(ctx, w, r, nil)
}

func GetRagFileContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	name := r.FormValue("file_name")
	
	path := *conf.RagConfInfo.KnowledgePath + "/" + name
	file, err := os.Open(path)
	if err != nil {
		logger.ErrorCtx(ctx, "open file error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	defer file.Close()
	
	content, err := io.ReadAll(file)
	if err != nil {
		logger.ErrorCtx(ctx, "write file error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	result := map[string]interface{}{
		"content": string(content),
	}
	
	utils.Success(ctx, w, r, result)
}

func DeleteRagFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	err = os.Remove(*conf.RagConfInfo.KnowledgePath + "/" + r.FormValue("file_name"))
	if err != nil {
		logger.ErrorCtx(ctx, "delete file error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	if conf.RagConfInfo.Store == nil {
		err = db.DeleteRagFileByFileName(r.FormValue("name"))
		if err != nil {
			logger.ErrorCtx(ctx, "delete dir error", "err", err)
			utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
			return
		}
	}
	
	utils.Success(ctx, w, r, nil)
}

func GetRagFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	page := utils.ParseInt(r.FormValue("page"))
	pageSize := utils.ParseInt(r.FormValue("page_size"))
	name := r.FormValue("name")
	
	ragFiles, err := db.GetRagFilesByPage(page, pageSize, name)
	if err != nil {
		logger.ErrorCtx(ctx, "get user error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	total, err := db.GetRagFilesCount(name)
	if err != nil {
		logger.ErrorCtx(ctx, "get user count error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBWriteFail, err)
		return
	}
	
	result := map[string]interface{}{
		"list":  ragFiles,
		"total": total,
	}
	utils.Success(ctx, w, r, result)
}
