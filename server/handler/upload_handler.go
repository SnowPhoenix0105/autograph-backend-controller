package handler

import (
	"autograph-backend-controller/domain/extractorcall"
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/filesave"
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/server/common"
	"autograph-backend-controller/utils"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strings"
)

func UploadFile(ctx *gin.Context) {
	handler := uploadFileHandler{
		ctx: ctx,
	}

	if err := handler.checkParam(); err != nil {
		logging.NewLogger().WithError(err).Errorf("parse req error: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, common.MakeUnknownErrorResp())
		return
	}

	if err := handler.produce(); err != nil {
		logging.NewLogger().WithError(err).Errorf("produce error: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, common.MakeUnknownErrorResp())
		return
	}

	ctx.JSON(http.StatusOK, common.MakeSuccessResp(nil))
}

type uploadFileHandler struct {
	ctx *gin.Context

	// params
	fileName string
	fileData []byte
}

func (h *uploadFileHandler) checkParam() error {
	contentType := h.ctx.GetHeader("Content-Type")
	if !strings.Contains(contentType, "multipart/form-data") {
		return utils.WrapErrorf(common.ErrContentTypeNotMultipartFormData,
			"actual Content-Type = [%s] not 'multipart/form-data'", contentType)
	}

	multipart, err := h.ctx.MultipartForm()
	if err != nil {
		return utils.WrapErrorf(err, "read multipart header fail")
	}

	fileHeaders := multipart.File["file"]

	for _, header := range fileHeaders {
		file, err := header.Open()
		if err != nil {
			return utils.WrapError(err, "open multipart file fail")
		}

		data, err := ioutil.ReadAll(file)
		if err != nil {
			return utils.WrapError(err, "read multipart file fail")
		}

		h.fileName = header.Filename
		h.fileData = data
	}

	return nil
}

func (h *uploadFileHandler) produce() error {
	textInfo, err := h.saveFile()
	if err != nil {
		return utils.WrapError(err, "call saveFile fail")
	}

	textIDs, err := h.saveText(textInfo)
	if err != nil {
		return utils.WrapError(err, "call saveText fail")
	}

	h.startExtraction(textIDs, textInfo.TextList)
	return nil
}

func (h *uploadFileHandler) saveFile() (filesave.SaveFileResp, error) {

	resp, err := filesave.SaveFile(h.fileData)
	if err != nil {
		return filesave.SaveFileResp{}, utils.WrapError(err, "save multipart file fail")
	}

	return resp, nil

}

func (h *uploadFileHandler) saveText(textInfo filesave.SaveFileResp) ([]uint, error) {
	hashBytes, err := hex.DecodeString(textInfo.Hash)
	if err != nil {
		return nil, utils.WrapError(err, "decode hash fail")
	}

	file := metadata.File{
		Model: gorm.Model{},
		Extra: metadata.Extra{},
		Type:  textInfo.Type,
		URL:   textInfo.URL,
		Name:  h.removeSuffix(h.fileName),
		Hash:  hashBytes,
	}
	err = metadata.DatabaseRaw().Create(&file).Error

	if err != nil {
		return nil, utils.WrapError(err, "save file metadata fail")
	}

	var texts []metadata.Text
	for _, text := range textInfo.TextList {
		texts = append(texts, metadata.Text{
			Content: text,
			FileID:  utils.UintToPtr(file.ID),
		})
	}

	metadata.DatabaseRaw().CreateInBatches(&texts, 128)

	var textIDList []uint
	for _, text := range texts {
		textIDList = append(textIDList, text.ID)
	}

	return textIDList, nil
}

func (h *uploadFileHandler) startExtraction(textIDs []uint, textList []string) {
	email := ""
	user, exist := h.ctx.Get(common.RequestContextKeyUser)
	if exist {
		userInfo, ok := user.(*common.UserInfo)
		if ok {
			email = userInfo.Email
		} else {
			logging.Default().Errorf("ctx.Get(%s) get [%#v] not (*server.UserInfo)", common.RequestContextKeyUser, user)
		}
	}

	go extractorcall.DoExtraction(textIDs, textList, h.fileName, email)
}

func (h *uploadFileHandler) removeSuffix(origin string) string {
	index := strings.LastIndexByte(origin, '.')
	if index < 0 || len(origin)-index > 5 {
		return origin
	}

	return origin[:index]
}
