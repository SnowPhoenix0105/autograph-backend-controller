package graph

import (
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/metadata"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
)

func TestBuildKG(t *testing.T) {
	logging.SetDefaultConfig(logging.GenerateTestConfig(t))

	database, err := metadata.CreateDatabase(metadata.GenerateTestConfig())
	require.Nil(t, err)

	setting := KGSetting{
		GetMetadataDatabase: func() *gorm.DB {
			return database
		},
		Logger: logging.NewLogger(),
	}

	ptrID := func(id uint) *uint {
		return &id
	}

	/*
		File0 --> Text0,  Text1,  ..., Text19
		File1 --> Text20, Text21, ..., Text39
		null  --> Text40, Text41, ..., Text44
	*/
	file := make([]metadata.File, 2)
	for i := 0; i < 2; i++ {
		name := fmt.Sprintf("File%d", i)
		hash := md5.Sum([]byte(name))
		file[i] = metadata.File{
			Type: "txt",
			URL:  "/" + name,
			Name: name,
			Hash: hash[:],
		}
	}
	err = database.Create(&file).Error
	require.Nil(t, err)

	text := make([]metadata.Text, 45)
	for i := 0; i < 45; i++ {
		var fileID *uint = nil
		if i/20 < 2 {
			fileID = ptrID(file[i/20].ID)
		}
		text[i] = metadata.Text{
			Content: fmt.Sprintf("Text%d", i),
			FileID:  fileID,
		}
	}
	err = database.Create(&text).Error
	require.Nil(t, err)

	/*
		Text[n] --> Node[2n], Node[2n+1] for n in [0, 45)
		null    --> Node90, Node91, ..., Node99
	*/

	getTextID := func(n int) *uint {
		if n < 90 {
			return ptrID(text[n/2].ID)
		}
		return nil
	}

	/*
						Node0, Node1, Node2, Node3, Node4, ..., Node98, Node99, R1, R2
		Extractor0 -->  E0,           E2,           E4,         E98,
		Extractor1 -->  E100,  E101,  E102,  E103,  E104,       E198,    E199
		Extractor2 -->  E200,  E201,  E202,  E203,  E204,       E298,    E299,   R1
		Extractor3 -->  E300,  E301,  E302,  E303,  E304,       E398,    E399, 	     R2
		Extractor4 -->  E400,  E401,  E402,  E403,  E404,       E498,    E499,   R1, R2

		Extractor5 ADD --> Node1  E[2n+1], n in [0, 25) (1, 3, ..., 49)
				   ADD --> Node98 E[2n+1]n n in [25, 48) (51, 53, ..., 95)
				   DEL --> Node1 E97
				   DEL --> Node94 E99
				   DEL --> Node98(E51) --R2--> Node1(E1)
	*/
	extractor := make([]metadata.Extractor, 7)
	for i := 0; i < 7; i++ {
		typ := metadata.ExtractorTypeModel
		if i == 6 {
			typ = metadata.ExtractorTypeHumanIntervention
		}
		extractor[i] = metadata.Extractor{
			Name: fmt.Sprintf("Extractor%d", i-1),
			Desc: fmt.Sprintf("Desc of Extractor%d", i-1),
			Type: typ,
		}
	}
	err = database.Create(&extractor).Error
	require.Nil(t, err)
	extractor = extractor[1:]

	entity := make([]metadata.Entity, 500)

	for i := 0; i < 100; i++ {
		entity[i].Type = metadata.EntityTypeAdd
		entity[i].ExtractorID = extractor[0].ID

		if i%2 == 0 {
			entity[i].Name = fmt.Sprintf("Node%d", i)
		} else {
			entity[i].ExtractorID = extractor[5].ID
			if i < 50 {
				entity[i].Name = "Node1"
			} else if i < 96 {
				entity[i].Name = "Node98"
			} else {
				entity[i].Type = metadata.EntityTypeDel
				if i == 97 {
					entity[i].Name = "Node1"
				} else { // i == 99
					entity[i].Name = "Node94"
				}
			}
		}
	}

	for i := 0; i < 100; i++ {
		for j := 1; j < 5; j++ {
			entity[j*100+i] = metadata.Entity{
				Name:        fmt.Sprintf("Node%d", i),
				Type:        metadata.EntityTypeAdd,
				ExtractorID: extractor[j].ID,
				TextID:      getTextID(i),
			}
		}
	}

	err = database.Create(&entity).Error
	require.Nil(t, err)

	/*

		    +---------R2------------+   +----------R2----------+
		    ↓                       |   ↓                      |
		Node0 --R1--> Node1 --R1--> Node2 --R1--> ... --R1--> Node99
		↑   ↑         |   ↑         |                          |
		|   +----R2---+   +----R2---+                          |
		|                                                      |
		+-------------------------R2---------------------------+
	*/
	r1Relations := make([]metadata.Relation, 198)
	for i := 0; i < 99; i++ {
		r1Relations[i] = metadata.Relation{
			Name:        "R1",
			Type:        metadata.EntityTypeAdd,
			ExtractorID: extractor[2].ID,
			TextID:      getTextID(i),
			HeadID:      entity[200+i].ID,
			TailID:      entity[200+i+1].ID,
		}
		r1Relations[99+i] = metadata.Relation{
			Name:        "R1",
			Type:        metadata.EntityTypeAdd,
			ExtractorID: extractor[4].ID,
			TextID:      getTextID(i),
			HeadID:      entity[400+i].ID,
			TailID:      entity[400+i+1].ID,
		}
	}
	err = database.Create(&r1Relations).Error
	require.Nil(t, err)
	r1Relations = nil

	for i := 1; i < 100; i++ {
		var tmp []metadata.Relation
		for j := 0; j < i; j++ {
			tmp = append(tmp, metadata.Relation{
				Name:        "R2",
				Type:        metadata.EntityTypeAdd,
				ExtractorID: extractor[3].ID,
				TextID:      getTextID(i),
				HeadID:      entity[300+i].ID,
				TailID:      entity[300+j].ID,
			})
			tmp = append(tmp, metadata.Relation{
				Name:        "R2",
				Type:        metadata.EntityTypeAdd,
				ExtractorID: extractor[4].ID,
				TextID:      getTextID(i),
				HeadID:      entity[400+i].ID,
				TailID:      entity[400+j].ID,
			})
		}
		err = database.Create(&tmp).Error
		require.Nil(t, err)
	}

	err = database.Create(&metadata.Relation{
		Name:        "R2",
		HeadID:      entity[51].ID,
		TailID:      entity[1].ID,
		Type:        metadata.EntityTypeDel,
		ExtractorID: extractor[5].ID,
	}).Error
	require.Nil(t, err)

	extractorIdList := make([]uint, len(extractor))
	for i := 0; i < len(extractor); i++ {
		extractorIdList[i] = extractor[i].ID
	}

	res, err := buildKG(&setting, context.TODO(), &KGBuildConfig{
		ExtractorIDList: extractorIdList,
		Desc:            "TestBuild",
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	assert.Equal(t, map[string]struct{}{"Node1": {}, "Node94": {}}, res.DeletedEntities)

	selected := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		if i == 1 || i == 94 {
			continue
		}
		selected[fmt.Sprintf("Node%d", i)] = struct{}{}
	}
	assert.Equal(t, selected, res.SelectedEntities)

	t.Logf("Build KG Cost: %v", res.FinishTime.Sub(res.StartTime))
	t.Logf("%#v", res)

	var build metadata.Build
	err = database.Take(&build, res.BuildID).Error
	require.Nil(t, err)
	require.Equal(t, "TestBuild", build.Desc)

	var extractorsSnapshot []metadata.BuildExtractor
	err = database.Where(&metadata.BuildExtractor{BuildID: res.BuildID}).Find(&extractorsSnapshot).Error
	require.Nil(t, err)
	require.Equal(t, len(res.ExtractorSnapshotMap), len(extractorsSnapshot))
	for i := 0; i < len(extractor); i++ {
		snapshotID := res.ExtractorSnapshotMap[extractor[i].ID]
		var target *metadata.BuildExtractor = nil
		for j := 0; j < len(extractorsSnapshot); j++ {
			if extractorsSnapshot[j].ID == snapshotID {
				target = &extractorsSnapshot[j]
				break
			}
		}
		require.NotNil(t, target)
		require.Equal(t, extractor[i].Name, target.Name)
		require.Equal(t, extractor[i].Desc, target.Desc)
	}

	// 检查已删除实体和关系

	var delNode []metadata.Node
	err = database.Where(&metadata.Node{BuildID: res.BuildID, Name: "Node1"}).Find(&delNode).Error
	require.Nil(t, err)
	require.Zero(t, len(delNode))

	err = database.Where(&metadata.Node{BuildID: res.BuildID, Name: "Node94"}).Find(&delNode).Error
	require.Nil(t, err)
	require.Zero(t, len(delNode))

	var node metadata.Node
	err = database.Where(metadata.Node{BuildID: res.BuildID, Name: "Node98"}).Take(&node).Error
	require.Nil(t, err)

	var outSchema metadata.SchemaNodeOut
	err = json.Unmarshal([]byte(node.OutJSON), &outSchema)
	require.Nil(t, err)

	_, ok := outSchema.NextNodes["Node1"]
	require.False(t, ok)

	// 检查 R2 关系

	for i := 0; i < 100; i++ {

		name := fmt.Sprintf("Node%d", i)

		if i == 1 || i == 94 {

			var delNode []metadata.Node
			err = database.Where(metadata.Node{BuildID: res.BuildID, Name: name}).Find(&delNode).Error
			require.Nil(t, err)
			require.Zero(t, len(delNode))

			continue
		}

		var node metadata.Node
		err = database.Where(metadata.Node{BuildID: res.BuildID, Name: name}).Take(&node).Error
		require.Nil(t, err)

		err = json.Unmarshal([]byte(node.OutJSON), &outSchema)
		require.Nil(t, err)

		for j := 0; j < i; j++ {
			nextNodeName := fmt.Sprintf("Node%d", j)

			if j == 1 || j == 94 {

				_, ok = outSchema.NextNodes[nextNodeName]
				require.Falsef(t, ok, "%#v --R2--> %#v", name, nextNodeName)

				continue
			}

			info := outSchema.NextNodes[nextNodeName]
			require.Equal(t, "R2", info.Name)

			if i >= 80 {
				require.Zero(t, len(info.Files))
			} else {
				require.Equal(t, []metadata.FileInfo{{
					FileID:   file[i/40].ID,
					FileName: file[i/40].Name,
				}}, info.Files)
			}

			require.Equal(t, 2, len(info.Extractors))
			require.Contains(t, info.Extractors, res.ExtractorSnapshotMap[extractor[3].ID])
			require.Contains(t, info.Extractors, res.ExtractorSnapshotMap[extractor[4].ID])
		}
	}

	// 检查R1关系

	for i := 0; i < 99; i++ {

		name := fmt.Sprintf("Node%d", i)

		if i == 1 || i == 94 {

			var delNode []metadata.Node
			err = database.Where(metadata.Node{BuildID: res.BuildID, Name: name}).Find(&delNode).Error
			require.Nil(t, err)
			require.Zero(t, len(delNode))

			continue
		}

		var node metadata.Node
		err = database.Where(metadata.Node{BuildID: res.BuildID, Name: name}).Take(&node).Error
		require.Nil(t, err)

		err = json.Unmarshal([]byte(node.OutJSON), &outSchema)
		require.Nil(t, err)

		nextNodeName := fmt.Sprintf("Node%d", i+1)

		if i == 0 || i == 93 {

			_, ok = outSchema.NextNodes[nextNodeName]
			require.Falsef(t, ok, "%#v --R2--> %#v", name, nextNodeName)

			continue
		}

		info := outSchema.NextNodes[nextNodeName]
		require.Equal(t, "R1", info.Name)

		if i >= 80 {
			require.Zero(t, len(info.Files))
		} else {
			require.Equal(t, []metadata.FileInfo{{
				FileID:   file[i/40].ID,
				FileName: file[i/40].Name,
			}}, info.Files)
		}

		require.Equal(t, 2, len(info.Extractors))
		require.Contains(t, info.Extractors, res.ExtractorSnapshotMap[extractor[2].ID])
		require.Contains(t, info.Extractors, res.ExtractorSnapshotMap[extractor[4].ID])
	}

	// 检查 Source

	for i := 0; i < 100; i++ {

		name := fmt.Sprintf("Node%d", i)

		if i == 1 || i == 94 {

			var delNode []metadata.Node
			err = database.Where(metadata.Node{BuildID: res.BuildID, Name: name}).Find(&delNode).Error
			require.Nil(t, err)
			require.Zero(t, len(delNode))

			continue
		}

		var node metadata.Node
		err = database.Where(metadata.Node{BuildID: res.BuildID, Name: name}).Take(&node).Error
		require.Nil(t, err)

		var sourceSchema metadata.SchemaNodeSource

		err = json.Unmarshal([]byte(node.SourceJSON), &sourceSchema)
		require.Nil(t, err)

		if i >= 80 {
			require.Zero(t, len(sourceSchema.Files))
		} else {
			require.Equal(t, []metadata.FileInfo{{
				FileID:   file[i/40].ID,
				FileName: file[i/40].Name,
			}}, sourceSchema.Files)
		}

		require.True(t, len(sourceSchema.Extractors) > 3)
		require.Contains(t, sourceSchema.Extractors, res.ExtractorSnapshotMap[extractor[1].ID])
		require.Contains(t, sourceSchema.Extractors, res.ExtractorSnapshotMap[extractor[2].ID])
		require.Contains(t, sourceSchema.Extractors, res.ExtractorSnapshotMap[extractor[3].ID])
		require.Contains(t, sourceSchema.Extractors, res.ExtractorSnapshotMap[extractor[4].ID])

		if i%2 == 1 {
			require.Equal(t, 4, len(sourceSchema.Extractors))
		} else {
			require.True(t, len(sourceSchema.Extractors) > 4)
			require.Contains(t, sourceSchema.Extractors, res.ExtractorSnapshotMap[extractor[0].ID])
			if i == 98 {
				require.Equal(t, 6, len(sourceSchema.Extractors))
				require.Contains(t, sourceSchema.Extractors, res.ExtractorSnapshotMap[extractor[5].ID])
			} else {
				require.Equal(t, 5, len(sourceSchema.Extractors))
			}
		}
	}
}
