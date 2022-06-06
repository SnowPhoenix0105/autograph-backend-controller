package handler

import (
	"autograph-backend-controller/domain/graph"
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/filesave"
	"autograph-backend-controller/repository/neograph"
	"autograph-backend-controller/server/common"
	"autograph-backend-controller/utils"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func BuildVersion(ctx *gin.Context) {
	handler := buildVersionHandler{
		ctx: ctx,
	}

	if err := handler.checkParam(); err != nil {
		logging.Default().WithError(err).Errorf("parse req error: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, common.MakeUnknownErrorResp())
		return
	}

	version, err := handler.produce()
	if err != nil {
		logging.Default().WithError(err).Errorf("produce error: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, common.MakeUnknownErrorResp())
		return
	}

	ctx.JSON(http.StatusOK, common.MakeSuccessResp(buildVersionRespSchema{Version: version}))
}

type buildVersionHandler struct {
	ctx *gin.Context

	// params
	extractorIDList []uint
	desc            string
}

type buildVersionReqSchema struct {
	ExtractorList []uint `json:"extractor_list"`
	Desc          string `json:"desc"`
}

type buildVersionRespSchema struct {
	Version uint `json:"version"`
}

func (h *buildVersionHandler) checkParam() error {
	var req buildVersionReqSchema
	if err := h.ctx.Bind(&req); err != nil {
		return utils.WrapError(err, "bind req fail")
	}

	if len(req.ExtractorList) == 0 {
		return utils.WrapError(common.ErrRequestParamEmpty, "param extractor_list is empty")
	}

	if len(req.Desc) == 0 {
		return utils.WrapError(common.ErrRequestParamEmpty, "param desc is empty")
	}

	h.extractorIDList = req.ExtractorList
	h.desc = req.Desc

	return nil
}

func (h *buildVersionHandler) produce() (uint, error) {
	buildInfo, err := graph.BuildKG(context.TODO(), &graph.KGBuildConfig{
		ExtractorIDList: h.extractorIDList,
		Desc:            h.desc,
	})

	if err != nil {
		return 0, utils.WrapErrorf(err, "build kg with extractors=%#v, desc=%#v fail", h.extractorIDList, h.desc)
	}

	entityData, relationData, err := graph.TransKGToCSV(buildInfo.BuildID)
	if err != nil {
		return 0, utils.WrapErrorf(err, "transform kg [%d] to csv fail", buildInfo.BuildID)
	}

	entityFileInfo, err := filesave.SaveFile(entityData)
	if err != nil {
		return 0, utils.WrapErrorf(err, "save entity csv fail", buildInfo.BuildID)
	}

	relationFileInfo, err := filesave.SaveFile(relationData)
	if err != nil {
		return 0, utils.WrapErrorf(err, "save relation csv fail", buildInfo.BuildID)
	}

	entityFileURL := fmt.Sprintf("http://%s/raw/%s", filesave.GetConfig().FullHost(), entityFileInfo.URL)
	relationFileURL := fmt.Sprintf("http://%s/raw/%s", filesave.GetConfig().FullHost(), relationFileInfo.URL)

	entityCypher := `
		load csv 
		with headers 
		from $url 
		as line 
		create(e:Entity{
			version:toInteger(line.version),
			name:toString(line.name),
			source:toString(line.source)
		});
	`
	relationCypher := `
		load csv
		with headers
		from $url
		as line
		match (h:Entity{
			version:toInteger(line.version),
			name:toString(line.head)
		}),(t:Entity{
			version:toInteger(line.version),
			name:toString(line.tail)
		})
		merge (h)-[r:Relation{
			version:toInteger(line.version),
			name:toString(line.rel)
		}]->(t)
	`

	_, err = neograph.Execute(entityCypher, map[string]interface{}{
		"url": entityFileURL,
	})
	if err != nil {
		return 0, utils.WrapError(err, "load csv to neo4j fail")
	}

	_, err = neograph.Execute(relationCypher, map[string]interface{}{
		"url": relationFileURL,
	})
	if err != nil {
		return 0, utils.WrapError(err, "load csv to neo4j fail")
	}

	return buildInfo.BuildID, nil
}
