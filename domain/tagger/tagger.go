package tagger

import (
	"autograph-backend-controller/utils"
	"context"
	"fmt"
	"github.com/yanyiwu/gojieba"
	"gorm.io/gorm"
)

type Range struct {
	Begin int
	End   int
}

type EntityInfo struct {
	Range
	Name string
}

type SPOTriple struct {
	HeadEntity Range
	TailEntity Range
	Relation   string
}

type SOTuple struct {
	HeadEntity string
	TailEntity string
}

type Tagger struct {
	jieba         *gojieba.Jieba
	entityIndex   map[string]struct{}
	relationIndex map[SOTuple]string
}

func newTaggerWithBuildID(setting *TagSetting, ctx context.Context, buildID uint) (*Tagger, error) {
	ret := Tagger{}
	ret.reset()

	err := ret.applyKG(ctx, setting.GetMetadataDatabase(), buildID)
	if err != nil {
		return nil, utils.WrapError(err, fmt.Sprintf("apply KG with buildID=[%d] fail", buildID))
	}

	return &ret, nil
}

func (t *Tagger) reset(jiebaPath ...string) {
	t.jieba = gojieba.NewJieba(jiebaPath...)
}

func (t *Tagger) applyIndex(entityIndex map[string]struct{}, relationIndex map[SOTuple]string) {
	entityAppend := len(t.entityIndex) != 0
	if !entityAppend {
		t.entityIndex = entityIndex
	}

	if len(t.relationIndex) == 0 {
		t.relationIndex = relationIndex
	} else {
		for k, v := range relationIndex {
			t.relationIndex[k] = v
		}
	}

	for entity := range entityIndex {

		t.jieba.AddWord(entity)

		if entityAppend {
			t.entityIndex[entity] = struct{}{}
		}
	}
}

func (t *Tagger) applyKG(ctx context.Context, db *gorm.DB, buildID uint) error {
	builder := indexBuilder{
		ctx:     ctx,
		buildID: buildID,
	}

	err := db.WithContext(ctx).Transaction(builder.Build)
	if err != nil {
		return utils.WrapError(err, "build index fail")
	}

	t.applyIndex(builder.entityIndex, builder.relationIndex)

	return nil
}

func (t *Tagger) wordsToEntities(words []gojieba.Word) []EntityInfo {
	var ret []EntityInfo

	for _, word := range words {

		_, ok := t.entityIndex[word.Str]
		if !ok {
			continue
		}

		ret = append(ret, EntityInfo{
			Range: Range{
				Begin: word.Start,
				End:   word.End,
			},
			Name: word.Str,
		})
	}

	return ret
}

func (t *Tagger) appendTripleIfRelationExists(triples []SPOTriple, head, tail EntityInfo) []SPOTriple {
	relation, ok := t.relationIndex[SOTuple{head.Name, tail.Name}]
	if ok {
		return append(triples, SPOTriple{
			HeadEntity: head.Range,
			TailEntity: tail.Range,
			Relation:   relation,
		})
	}
	return triples
}

func (t *Tagger) Produce(text string) []SPOTriple {
	words := t.jieba.Tokenize(text, gojieba.DefaultMode, true)
	entities := t.wordsToEntities(words)

	var ret []SPOTriple

	for i := 1; i < len(entities); i++ {
		for j := 0; j < i; j++ {
			front := entities[j]
			back := entities[i]

			ret = t.appendTripleIfRelationExists(ret, front, back)
			ret = t.appendTripleIfRelationExists(ret, back, front)
		}
	}

	return ret
}
