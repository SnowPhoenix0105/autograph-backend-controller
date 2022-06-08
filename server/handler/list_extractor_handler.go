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

func ListExtractor(ctx *gin.Context) {
	res, err := listExtractor()
	if err != nil {
		logging.Default().WithError(err).Errorf("ListVersion produce error: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, common.MakeUnknownErrorResp())
	}

	ctx.JSON(http.StatusOK, res)
}

type listExtractorItem struct {
	ID      uint   `json:"id"`
	Time    int64  `json:"time"`
	TimeStr string `json:"time_str"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
}

func listExtractor() ([]listExtractorItem, error) {
	extractorList := make([]metadata.Extractor, 0)
	res := metadata.DatabaseRaw().Find(&extractorList)
	err := res.Error
	if err != nil {
		return nil, utils.WrapError(err, "select all files fail")
	}

	ret := make([]listExtractorItem, 0, res.RowsAffected)
	for _, extractor := range extractorList {
		ret = append(ret, listExtractorItem{
			ID:      extractor.ID,
			Time:    extractor.CreatedAt.Unix(),
			TimeStr: extractor.CreatedAt.Format(time.RFC3339),
			Name:    extractor.Name,
			Desc:    extractor.Desc,
		})
	}

	return ret, nil
}
