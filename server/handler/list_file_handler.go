package handler

import (
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/server/common"
	"autograph-backend-controller/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func ListFile(ctx *gin.Context) {
	res, err := listFile()
	if err != nil {
		logging.Default().WithError(err).Errorf("ListFile produce error: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, common.MakeUnknownErrorResp())
	}

	ctx.JSON(http.StatusOK, res)
}

type listFileItem struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	CreateTime    int64  `json:"create_time"`
	CreateTimeStr string `json:"create_time_str"`
}

func listFile() ([]listFileItem, error) {
	var fileList []metadata.File
	res := metadata.DatabaseRaw().Find(&fileList)
	err := res.Error
	if err != nil {
		return nil, utils.WrapError(err, "select all files fail")
	}

	ret := make([]listFileItem, 0, res.RowsAffected)
	for _, file := range fileList {
		ret = append(ret, listFileItem{
			Name:          fmt.Sprintf("%s.%s", file.Name, file.Type),
			Type:          file.Type,
			CreateTime:    file.CreatedAt.Unix(),
			CreateTimeStr: file.CreatedAt.Format(time.RFC3339),
		})
	}

	return ret, nil
}
