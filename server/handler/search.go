package handler

import (
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/neograph"
	"autograph-backend-controller/server/common"
	"autograph-backend-controller/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"net/http"
	"net/url"
	"strconv"
)

func Search(ctx *gin.Context) {
	handler := searchHandler{
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

type searchHandler struct {
	ctx *gin.Context

	// param
	query   string
	version uint

	// status
	nodeSet  map[searchNode]struct{}
	linkSet  map[searchLink]struct{}
	fileList []searchFile
}

func (h *searchHandler) checkParam() error {
	query := h.ctx.Query("q")

	if len(query) == 0 {
		return utils.WrapError(common.ErrRequestParamEmpty, "query is empty")
	}

	unescaped, err := url.QueryUnescape(query)
	if err != nil {
		return utils.WrapErrorf(err, "unescape [%#v] fail", query)
	}

	versionStr := h.ctx.Query("v")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return utils.WrapErrorf(err, "varsion atoi(%#v) fail:", versionStr)
	}

	h.query = unescaped
	h.version = uint(version)

	logging.Default().Infof("query=%#v, version=%d", h.query, h.version)

	return nil
}

type searchLink struct {
	Source string `json:"source"`
	Name   string `json:"name"`
	Target string `json:"target"`
}

type searchNode struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type searchGraph struct {
	Links []searchLink `json:"links"`
	Nodes []searchNode `json:"nodes"`
}

type searchFile struct {
	FileID   uint   `json:"file_id"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
}

type searchNodeSourceSchema struct {
	Files      []searchFile `json:"files"`
	Extractors []uint       `json:"extractors"`
}

type searchResp struct {
	Graph searchGraph  `json:"graph"`
	Files []searchFile `json:"files"`
	Query string       `json:"query"`
}

func (h *searchHandler) produce() (*searchResp, error) {

	h.nodeSet = make(map[searchNode]struct{})
	h.linkSet = make(map[searchLink]struct{})

	skips := map[int]bool{
		0: true,
		1: true,
	}
	cyphers := []string{
		`match 
			(e1:Entity{
				version: $version,
				name: $name
				})-[r1:Relation{
					version: $version
					}]->(e2:Entity{
						version: $version
					}),
			(e3:Entity{
				version: $version
				})-[r2:Relation{
					version: $version
					}]->(e1)
			return e1, e2, e3, r1, r2
		`,
		`match 
			(e1:Entity{
				version: $version,
				name: $name
				})-[r1:Relation{
					version: $version
					}]->(e2:Entity{
						version: $version
					})-[r2:Relation{
						version: $version
						}]->(e3:Entity{
							version: $version
						}),
			(e4:Entity{
				version: $version
				})-[r3:Relation{
					version: $version
					}]->(e5:Entity{
					version: $version
						})-[r4:Relation{
							version: $version
							}]->(e1)
			return e1, e2, e3, e4, e5, r1, r2, r3, r4
		`,
		`match 
			(e1:Entity{
				version: $version,
				name: $name
				})-[r1:Relation{
					version: $version
					}]->(e2:Entity{
						version: $version
					})
			return e1, e2, r1
		`,
		`match 
			(e1:Entity{
				version: $version,
				name: $name
				})-[r1:Relation{
					version: $version
					}]->(e2:Entity{
						version: $version
					})-[r2:Relation{
						version: $version
						}]->(e3:Entity{
							version: $version
						})
			return e1, e2, e3, r1, r2
		`,
	}

	for i, cypher := range cyphers {
		if skips[i] {
			continue
		}

		if err := h.produceCypher(cypher); err != nil {
			return nil, utils.WrapErrorf(err, "produce cypher [%d] fail", i)
		}
	}

	var nodeList []searchNode
	for node := range h.nodeSet {
		nodeList = append(nodeList, node)
	}

	var linkList []searchLink
	for link := range h.linkSet {
		linkList = append(linkList, link)
	}

	ret := searchResp{
		Graph: searchGraph{
			Links: linkList,
			Nodes: nodeList,
		},
		Files: h.fileList,
		Query: h.query,
	}

	return &ret, nil
}

func (h *searchHandler) produceCypher(cypher string) error {
	res, err := neograph.Execute(cypher, map[string]interface{}{
		"version": h.version,
		"name":    h.query,
	})
	if err != nil {
		return utils.WrapErrorf(err, "execute query with [version=%d, query=%#v] fail", h.version, h.query)
	}

	for _, record := range res {
		if err := h.produceRecord(record); err != nil {
			return utils.WrapError(err, "produce record fail")
		}
	}

	return nil
}

func (h *searchHandler) produceRecord(record *neo4j.Record) error {
	for i, valueIface := range record.Values {
		switch value := valueIface.(type) {
		case neo4j.Node:
			err := h.produceNode(&value)
			if err != nil {
				return utils.WrapErrorf(err, "produce node [%d] fail")
			}
		case neo4j.Relationship:
			err := h.produceRelation(&value)
			if err != nil {
				return utils.WrapErrorf(err, "produce relation [%d] fail")
			}
		default:
			logging.Default().Warnf("key=%#v, unknown type of value=%#v", record.Keys[i], value)
		}
	}

	return nil
}

func (h *searchHandler) produceNode(node *neo4j.Node) error {
	nameIface, exist := node.Props["name"]
	if !exist {
		return fmt.Errorf("node name key not found in prop [%#v]", node.Props)
	}

	name, ok := nameIface.(string)
	if !ok {
		return fmt.Errorf("node name is not string [%#v]", nameIface)
	}

	h.nodeSet[searchNode{
		Name: name,
		ID:   strconv.Itoa(int(node.Id)),
	}] = struct{}{}

	if len(h.fileList) == 0 && h.query == name {
		sourceIface, exist := node.Props["source"]
		if !exist {
			return fmt.Errorf("node source key not found in prop [%#v]", node.Props)
		}

		sourceStr, ok := sourceIface.(string)
		if !ok {
			return fmt.Errorf("node source is not string [%#v]", nameIface)
		}

		var source searchNodeSourceSchema
		if err := json.Unmarshal([]byte(sourceStr), &source); err != nil {
			return utils.WrapErrorf(err, "unmarshal source json [%#v] fail", sourceStr)
		}

		h.fileList = source.Files
	}

	return nil
}

func (h *searchHandler) produceRelation(relation *neo4j.Relationship) error {
	nameIface, exist := relation.Props["name"]
	if !exist {
		return fmt.Errorf("relation name key not found in prop [%#v]", relation.Props)
	}

	name, ok := nameIface.(string)
	if !ok {
		return fmt.Errorf("relation name is not string [%#v]", nameIface)
	}

	h.linkSet[searchLink{
		Source: strconv.Itoa(int(relation.StartId)),
		Name:   name,
		Target: strconv.Itoa(int(relation.EndId)),
	}] = struct{}{}

	return nil
}
