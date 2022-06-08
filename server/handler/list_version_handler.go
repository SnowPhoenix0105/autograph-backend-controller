package handler

import (
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/server/common"
	"autograph-backend-controller/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func ListVersion(ctx *gin.Context) {
	res, err := listVersion()
	if err != nil {
		logging.Default().WithError(err).Errorf("ListVersion produce error: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, common.MakeUnknownErrorResp())
	}

	ctx.JSON(http.StatusOK, res)
}

type listVersionItem struct {
	ID      uint   `json:"id"`
	Time    int64  `json:"time"`
	TimeStr string `json:"time_str"`
	Desc    string `json:"desc"`
}

func listVersion() ([]listVersionItem, error) {
	versionList := make([]metadata.Build, 0)
	res := metadata.DatabaseRaw().Find(&versionList)
	err := res.Error
	if err != nil {
		return nil, utils.WrapError(err, "select all files fail")
	}

	ret := make([]listVersionItem, 0, res.RowsAffected)
	for _, file := range versionList {
		ret = append(ret, listVersionItem{
			ID:      file.ID,
			Time:    file.CreatedAt.Unix(),
			TimeStr: file.CreatedAt.Format(time.RFC3339),
			Desc:    file.Desc,
		})
	}

	return ret, nil
}
