package tagger

import (
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/utils"
	"context"
	"encoding/json"
	"gorm.io/gorm"
)

type indexBuilder struct {
	// inputs
	ctx     context.Context
	buildID uint

	// outputs
	relationIndex map[SOTuple]string
	entityIndex   map[string]struct{}
}

func (b *indexBuilder) Build(tx *gorm.DB) error {
	b.relationIndex = make(map[SOTuple]string)
	b.entityIndex = make(map[string]struct{})

	err := b.build(tx)
	if err != nil {
		return utils.WrapError(err, "build fail")
	}

	return nil
}

func (b *indexBuilder) build(tx *gorm.DB) error {
	var batchData []metadata.Node

	res := tx.Where(&metadata.Node{BuildID: b.buildID}).
		FindInBatches(&batchData, 128, func(tx *gorm.DB, batchNum int) error {
			for i := 0; i < len(batchData); i++ {

				name := batchData[i].Name
				outJSON := batchData[i].OutJSON

				b.entityIndex[name] = struct{}{}

				var outSchema metadata.SchemaNodeOut
				err := json.Unmarshal([]byte(outJSON), &outSchema)

				if err != nil {
					return utils.WrapError(err, "outSchema json unmarshal fail")
				}

				for next, relationInfo := range outSchema.NextNodes {
					soTuple := SOTuple{
						HeadEntity: name,
						TailEntity: next,
					}
					b.relationIndex[soTuple] = relationInfo.Name
				}
			}

			batchData = nil
			return nil
		})

	return res.Error
}
