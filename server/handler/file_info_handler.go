package handler

import (
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/server/common"
	"autograph-backend-controller/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetFileInfo(ctx *gin.Context) {
	handler := getFileInfoHandler{
		ctx: ctx,
	}

	if err := handler.checkParam(); err != nil {
		logging.Default().WithError(err).Errorf("parse req error: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, common.MakeUnknownErrorResp())
		return
	}

	resp, err := handler.produce()
	if err != nil {
		logging.Default().WithError(err).Errorf("produce error: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, common.MakeUnknownErrorResp())
		return
	}

	ctx.JSON(http.StatusOK, common.MakeSuccessResp(resp))
}

type getFileInfoHandler struct {
	ctx *gin.Context

	// params
	id uint
}

type getFileInfoResp struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (h *getFileInfoHandler) checkParam() error {
	id := h.ctx.Query("id")

	if len(id) == 0 {
		return utils.WrapError(common.ErrRequestParamEmpty, "query 'id' is empty")
	}

	idInteger, err := strconv.Atoi(id)
	if err != nil {
		return utils.WrapErrorf(err, "atoi(%#v) fail", id)
	}

	if idInteger < 0 {
		return utils.WrapErrorf(common.ErrRequestParamInvalid, "id(%d) cannot be negative", idInteger)
	}

	h.id = uint(idInteger)

	return nil
}

func (h *getFileInfoHandler) produce() (*getFileInfoResp, error) {
	var file metadata.File
	err := metadata.DatabaseRaw().First(&file, h.id).Error
	if err != nil {
		return nil, utils.WrapErrorf(err, "get file-info[id=%d] from db fail", h.id)
	}

	return &getFileInfoResp{
		URL:  file.URL,
		Name: file.Name,
		Type: file.Type,
	}, nil
}
