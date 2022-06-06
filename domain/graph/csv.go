package graph

import (
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func transKGToCSV(setting *KGSetting, buildID uint) ([]byte, []byte, error) {
	db := setting.GetMetadataDatabase()

	builder := &csvBuilder{
		buildID: buildID,
		db:      db,
		logger:  setting.Logger,
	}

	if err := builder.buildCSV(); err != nil {
		return nil, nil, utils.WrapError(err, "build csv fail")
	}

	return builder.entityCSV.Bytes(), builder.relationCSV.Bytes(), nil
}

type csvBuilder struct {
	// input
	buildID uint
	db      *gorm.DB
	logger  *logrus.Logger

	// output
	entityCSV   bytes.Buffer
	relationCSV bytes.Buffer
}

func (b *csvBuilder) buildCSV() error {
	var nodes []metadata.Node
	err := b.db.Model(&metadata.Node{}).Where(&metadata.Node{
		BuildID: b.buildID,
	}).Find(&nodes).Error

	if err != nil {
		return utils.WrapErrorf(err, "collect nodes [buildID=%d] fail", b.buildID)
	}

	// 写文件头
	b.entityCSV.WriteString("version,name,source")
	b.relationCSV.WriteString("version,head,rel,tail")

	// 写各个csv文件内容
	for _, node := range nodes {
		if err := b.produceNode(&node); err != nil {
			return utils.WrapErrorf(err, "produce node [%s] fail", node.Name)
		}
	}

	return nil
}

func (b *csvBuilder) produceNode(node *metadata.Node) error {
	if err := b.recordEntity(node); err != nil {
		return utils.WrapError(err, "record entity fail")
	}

	if err := b.recordRelation(node); err != nil {
		return utils.WrapError(err, "record relation fail")
	}

	return nil
}

func (b *csvBuilder) recordEntity(node *metadata.Node) error {
	_, err := b.entityCSV.WriteString(fmt.Sprintf("\n%d,%#v,%#v", node.BuildID, node.Name, node.SourceJSON))
	if err != nil {
		return utils.WrapErrorf(err, "record entity [%#v] fail", node.Name)
	}

	return nil
}

func (b *csvBuilder) recordRelation(node *metadata.Node) error {
	var outSchema metadata.SchemaNodeOut
	if err := json.Unmarshal([]byte(node.OutJSON), &outSchema); err != nil {
		return utils.WrapErrorf(err, "json unmarshal out-rel of node [%#v] fail", node.Name)
	}

	for tail, rel := range outSchema.NextNodes {
		if node.Name == tail {
			continue
		}

		_, err := b.relationCSV.WriteString(fmt.Sprintf("\n%d,%#v,%#v,%#v", node.BuildID, node.Name, rel.Name, tail))
		if err != nil {
			return utils.WrapErrorf(err, "record spo <%#v, %#v, %#v> error", node.Name, rel.Name, tail)
		}
	}

	return nil
}
