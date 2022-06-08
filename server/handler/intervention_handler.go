package handler

import (
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/server/common"
	"autograph-backend-controller/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func Intervention(ctx *gin.Context) {
	handler := interventionHandler{ctx: ctx}

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

type interventionHandler struct {
	ctx *gin.Context

	req *interventionReq
}

type spoSchema struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

type interventionReq struct {
	AddEntity   []string    `json:"add_entity"`
	AddRelation []spoSchema `json:"add_relation"`
	DelEntity   []string    `json:"del_entity"`
	DelRelation []spoSchema `json:"del_relation"`
}

type interventionResp struct {
	ExtractorName string `json:"extractor_name"`
	ExtractorID   uint   `json:"extractor_id"`
}

func (h *interventionHandler) checkParam() error {
	var req interventionReq
	if err := h.ctx.Bind(&req); err != nil {
		return utils.WrapError(err, "bind req fail")
	}

	h.req = &req

	return nil
}

func (h *interventionHandler) produce() (*interventionResp, error) {
	entityCollection := make(map[string]uint)
	relationCollection := make(map[spoSchema]uint)

	// TMP实体 & ADD关系
	for _, spo := range h.req.AddRelation {
		entityCollection[spo.Object] = metadata.EntityTypeTmp
		entityCollection[spo.Subject] = metadata.EntityTypeTmp
		relationCollection[spo] = metadata.EntityTypeAdd
	}

	// TMP实体 & DEL关系
	for _, spo := range h.req.DelRelation {
		entityCollection[spo.Object] = metadata.EntityTypeTmp
		entityCollection[spo.Subject] = metadata.EntityTypeTmp
		relationCollection[spo] = metadata.EntityTypeDel
	}

	// ADD实体
	for _, ent := range h.req.AddEntity {
		entityCollection[ent] = metadata.EntityTypeAdd
	}

	// DEL实体
	for _, ent := range h.req.DelEntity {
		entityCollection[ent] = metadata.EntityTypeDel
	}

	if len(entityCollection) == 0 {
		return &interventionResp{
			ExtractorName: "",
			ExtractorID:   0,
		}, nil
	}

	extractor := metadata.Extractor{
		Name: fmt.Sprintf("人工干预%s", time.Now().Format(time.RFC3339)),
		Desc: fmt.Sprintf("人工干预%s", time.Now().Format(time.RFC3339)),
		Type: metadata.ExtractorTypeHumanIntervention,
	}
	entityIndex := make(map[string]int)
	entities := make([]metadata.Entity, 0, len(entityCollection))
	relations := make([]metadata.Relation, 0, len(relationCollection))

	// 入库
	if err := metadata.DatabaseRaw().Transaction(func(tx *gorm.DB) error {
		if err := metadata.DatabaseRaw().Create(&extractor).Error; err != nil {
			return utils.WrapError(err, "create extractor fail")
		}
		for ent, typ := range entityCollection {
			entityIndex[ent] = len(entities)
			entities = append(entities, metadata.Entity{
				Name:        ent,
				Type:        typ,
				ExtractorID: extractor.ID,
			})
		}

		if err := metadata.DatabaseRaw().Create(&entities).Error; err != nil {
			return utils.WrapError(err, "create entities fail")
		}

		if len(relationCollection) == 0 {
			return nil
		}

		for spo, typ := range relationCollection {
			relations = append(relations, metadata.Relation{
				Name:        spo.Predicate,
				Type:        typ,
				ExtractorID: extractor.ID,
				HeadID:      entities[entityIndex[spo.Subject]].ID,
				TailID:      entities[entityIndex[spo.Object]].ID,
			})
		}

		if err := metadata.DatabaseRaw().Create(&relations).Error; err != nil {
			return utils.WrapError(err, "create relations fail")
		}

		return nil

	}); err != nil {
		return nil, utils.WrapError(err, "insert data fail")
	}

	return &interventionResp{
		ExtractorName: extractor.Name,
		ExtractorID:   extractor.ID,
	}, nil
}
