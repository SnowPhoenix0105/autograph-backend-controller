package tagger

import (
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TagSetting struct {
	Logger              *logrus.Logger
	GetMetadataDatabase func() *gorm.DB
}

var globalSetting TagSetting

func Init(setting *TagSetting) {
	globalSetting = *setting
}

func NewTaggerWithBuildID(ctx context.Context, buildID uint) (*Tagger, error) {
	return newTaggerWithBuildID(&globalSetting, ctx, buildID)
}
