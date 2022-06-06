package graph

import (
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type KGSetting struct {
	GetMetadataDatabase func() *gorm.DB
	Logger              *logrus.Logger
}

var globalSetting KGSetting

func Init(setting *KGSetting) {
	globalSetting = *setting
}

func BuildKG(ctx context.Context, buildInfo *KGBuildConfig) (*KGBuildResult, error) {
	return buildKG(&globalSetting, ctx, buildInfo)
}

func TransKGToCSV(buildID uint) ([]byte, []byte, error) {
	return transKGToCSV(&globalSetting, buildID)
}
