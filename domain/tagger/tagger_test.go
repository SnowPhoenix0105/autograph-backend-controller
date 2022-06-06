package tagger

import (
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/metadata"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yanyiwu/gojieba"
	"gorm.io/gorm"
	"testing"
)

func TestTagger_JiebaTokenize(t *testing.T) {
	tagger := Tagger{}
	tagger.reset()

	text := "【数组变量】作为右值时会退化为【指向数组首元素的指针】"
	words := tagger.jieba.Tokenize(text, gojieba.DefaultMode, true)

	for i, word := range words {
		t.Logf("[%d]%s", i, word.Str)
		assert.Equal(t, word.Str, text[word.Start:word.End])
	}
}

func TestTagger_Produce(t *testing.T) {
	logging.SetDefaultConfig(logging.GenerateTestConfig(t))

	database, err := metadata.CreateDatabase(metadata.GenerateTestConfig())
	require.Nil(t, err)

	var build metadata.Build
	err = database.Create(&build).Error
	require.Nil(t, err)

	extractor := metadata.BuildExtractor{BuildID: build.ID}
	err = database.Create(&extractor).Error
	require.Nil(t, err)

	out := []metadata.SchemaNodeOut{
		{
			NextNodes: map[string]metadata.RelationInfo{
				"指针": {
					Name:       "相关",
					Extractors: []uint{extractor.ID},
				},
			},
		},
		{
			NextNodes: map[string]metadata.RelationInfo{
				"数组": {
					Name:       "相关",
					Extractors: []uint{extractor.ID},
				},
			},
		},
	}
	source := metadata.SchemaNodeSource{
		Files:      nil,
		Extractors: []uint{extractor.ID},
	}
	nodes := []metadata.Node{
		{
			BuildID:    build.ID,
			Name:       "数组",
			OutJSON:    out[0].ToJSON(),
			SourceJSON: source.ToJSON(),
		},
		{
			BuildID:    build.ID,
			Name:       "指针",
			OutJSON:    out[1].ToJSON(),
			SourceJSON: source.ToJSON(),
		},
	}
	err = database.Create(&nodes).Error
	require.Nil(t, err)

	tagger, err := newTaggerWithBuildID(&TagSetting{
		Logger: logging.NewLogger(),
		GetMetadataDatabase: func() *gorm.DB {
			return database
		},
	}, context.TODO(), build.ID)
	require.Nil(t, err)

	text := "数组变量作为右值时会退化为指向数组首元素的指针"
	spoTriples := tagger.Produce(text)

	assert.True(t, len(spoTriples) >= 4)

	type SPO struct {
		S, P, O string
	}

	expect := []SPO{{"数组", "相关", "指针"}, {"指针", "相关", "数组"}}
	for i, spo := range spoTriples {
		head := text[spo.HeadEntity.Begin:spo.HeadEntity.End]
		tail := text[spo.TailEntity.Begin:spo.TailEntity.End]
		t.Logf("[%d] %s(%d:%d) --%s--> %s(%d:%d)",
			i,
			head,
			spo.HeadEntity.Begin,
			spo.HeadEntity.End,
			spo.Relation,
			tail,
			spo.TailEntity.Begin,
			spo.TailEntity.End,
		)

		res := SPO{head, spo.Relation, tail}

		found := false
		for _, exp := range expect {
			if exp == res {
				found = true
			}
		}

		assert.True(t, found)
	}
}
