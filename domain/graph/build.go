package graph

import (
	"context"
	"time"
)

type KGBuildConfig struct {
	ExtractorIDList []uint
	Desc            string
}

type KGBuildResult struct {
	BuildID              uint
	ExtractorSnapshotMap map[uint]uint // Extractor.ID -> BuildExtractor.ID
	DeletedEntities      map[string]struct{}
	SelectedEntities     map[string]struct{}
	StartTime            time.Time
	FinishTime           time.Time
}

func buildKG(setting *KGSetting, ctx context.Context, config *KGBuildConfig) (*KGBuildResult, error) {
	builder := kgBuilder{
		config:  config,
		result:  &KGBuildResult{},
		ctx:     ctx,
		setting: setting,
	}
	err := setting.GetMetadataDatabase().WithContext(ctx).Transaction(builder.build)
	return builder.result, err
}
