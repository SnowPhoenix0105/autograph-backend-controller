package graph

import (
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/utils"
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type kgBuilder struct {
	config  *KGBuildConfig
	result  *KGBuildResult
	ctx     context.Context
	tx      *gorm.DB
	setting *KGSetting
}

func (b *kgBuilder) build(tx *gorm.DB) error {
	b.tx = tx

	defer func() {
		b.result.FinishTime = time.Now()
	}()

	err := b.createBuild()
	if err != nil {
		return utils.WrapError(err, "create build fail")
	}

	humanInterventionExtractorIDList, err := b.createExtractorSnapshot()
	if err != nil {
		return utils.WrapError(err, "create snapshot of extractor fail")
	}

	entities, err := b.collectEntities(humanInterventionExtractorIDList)
	if err != nil {
		return utils.WrapError(err, "collect entities fail")
	}

	err = b.createNodes(entities)
	if err != nil {
		return utils.WrapError(err, "create nodes fail")
	}

	return nil
}

////////////// 构建的基本信息 /////////////////

/*
createBuild 创建一条构建记录，并填充 b.result 的 BuildID 和 StartTime 字段
*/
func (b *kgBuilder) createBuild() error {
	build := metadata.Build{
		Desc: b.config.Desc,
	}

	if err := b.tx.Create(&build).Error; err != nil {
		return utils.WrapError(err, "insert build to db fail")
	}

	b.result.BuildID = build.ID
	b.result.StartTime = build.CreatedAt
	return nil
}

/*
createExtractorSnapshot 创建所有 b.config 中的抽取模型的快照， 并填充 b.result.ExtractorSnapshotMap 字段。
返回 Type 为 HumanIntervention 的 Extractor.ID
*/
func (b *kgBuilder) createExtractorSnapshot() ([]uint, error) {
	var extractors []metadata.Extractor
	if err := b.tx.Find(&extractors, b.config.ExtractorIDList).Error; err != nil {
		return nil, utils.WrapError(err, "select extractors fail")
	}

	buildExtractors := make([]metadata.BuildExtractor, len(extractors))
	for i := 0; i < len(buildExtractors); i++ {
		buildExtractors[i] = metadata.BuildExtractor{
			BuildID: b.result.BuildID,
			Name:    extractors[i].Name,
			Desc:    extractors[i].Desc,
			Type:    extractors[i].Type,
		}
	}

	if err := b.tx.Create(&buildExtractors).Error; err != nil {
		return nil, utils.WrapError(err, "insert extractors to db fail")
	}

	snapshotMap := make(map[uint]uint, len(b.config.ExtractorIDList))
	for i := 0; i < len(buildExtractors); i++ {
		snapshotMap[extractors[i].ID] = buildExtractors[i].ID
	}
	b.result.ExtractorSnapshotMap = snapshotMap

	ret := make([]uint, 0)
	for i := 0; i < len(extractors); i++ {
		if extractors[i].Type == metadata.ExtractorTypeHumanIntervention {
			ret = append(ret, extractors[i].ID)
		}
	}

	return ret, nil
}

////////////// 确定要构建的 Node 的范围 //////////////

/*
collectDeletedEntities 获取被删除实体的集合，即收集所有 Type=Del 的 Entity 的 Name。
*/
func (b *kgBuilder) collectDeletedEntities(humanInterventionExtractorIDList []uint) (del map[string]struct{}, err error) {
	rows, err := b.tx.Model(&metadata.Entity{}).
		Distinct("name").
		Select("name").
		Where("extractor_id in ? and type = ?", humanInterventionExtractorIDList, metadata.EntityTypeDel).
		Rows()
	if err != nil {
		return nil, utils.WrapError(err, "select from database fail")
	}
	defer rows.Close()

	del = make(map[string]struct{})
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, utils.WrapError(err, "scan result fail")
		}
		del[name] = struct{}{}
	}

	return del, nil
}

/*
collectEntities 获取所有未被删除的实体的列表，收集所有 Entity 的 Name，然后排除 Type=Del 的 Entity 的 Name。
填充 b.result.DeletedEntities 和 b.result.SelectedEntities。
*/
func (b *kgBuilder) collectEntities(humanInterventionExtractorIDList []uint) (map[string]struct{}, error) {
	del, err := b.collectDeletedEntities(humanInterventionExtractorIDList)
	if err != nil {
		return nil, utils.WrapError(err, "collect deleted entities fail")
	}

	b.result.DeletedEntities = del
	b.setting.Logger.Infof("deleted node set=%v", del)

	rows, err := b.tx.Model(&metadata.Entity{}).
		Distinct("name").
		Select("name").
		Where("extractor_id in ? and type != ?", b.config.ExtractorIDList, metadata.EntityTypeTmp).
		Rows()
	if err != nil {
		return nil, utils.WrapError(err, "select entities from db fail")
	}
	defer rows.Close()

	ret := make(map[string]struct{}, 0)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, utils.WrapError(err, "scan result fail")
		}

		// 排除删除的实体
		_, ok := del[name]
		if ok {
			continue
		}

		ret[name] = struct{}{}
	}

	b.result.SelectedEntities = ret
	b.setting.Logger.Infof("selected node set(len=%d)=%v", len(ret), ret)

	return ret, nil
}

////////////// 构建 Node ///////////////

/*
getFileIDByTextID 通过 Text.ID 获取 File.ID。
*/
func (b *kgBuilder) getFileIDByTextID(textID *uint) (*uint, error) {
	if textID == nil {
		return nil, nil
	}

	var text metadata.Text
	if err := b.tx.Take(&text, *textID).Error; err != nil {
		return nil, utils.WrapError(err, fmt.Sprintf("select text with id=[%d] fail", *textID))
	}

	return text.FileID, nil
}

func (b *kgBuilder) getFileNameAndTypeByID(fileID uint) (string, string, error) {
	var file metadata.File
	if err := b.tx.Take(&file, fileID).Error; err != nil {
		return "", "", utils.WrapError(err, fmt.Sprintf("select file with id=[%d] fail", fileID))
	}

	return fmt.Sprintf("%s.%s", file.Name, file.Type), file.Type, nil
}

/*
buildSourceForNode 通过构建 Node 的 Entity 列表，构建 Node 的来源信息
*/
func (b *kgBuilder) buildSourceForNode(entities []metadata.Entity) (metadata.SchemaNodeSource, error) {
	ret := metadata.SchemaNodeSource{}

	// files

	files := make([]metadata.FileInfo, 0)
	checked := make(map[uint]struct{})

	for i := 0; i < len(entities); i++ {
		textID := entities[i].TextID

		fileID, err := b.getFileIDByTextID(textID)
		if err != nil {
			return ret, utils.WrapError(err, "get fileID with textID fail")
		}

		if fileID == nil {
			continue
		}

		_, ok := checked[*fileID]
		if ok {
			continue
		}

		fileName, fileType, err := b.getFileNameAndTypeByID(*fileID)
		if err != nil {
			return ret, utils.WrapError(err, "get fileName with fileID fail")
		}

		checked[*fileID] = struct{}{}
		files = append(files, metadata.FileInfo{
			FileID:   *fileID,
			FileName: fileName,
			FileType: fileType,
		})
	}

	// extractors

	extractors := make([]uint, 0)
	checked = make(map[uint]struct{})

	for i := 0; i < len(entities); i++ {
		extractorID := entities[i].ExtractorID

		_, ok := checked[extractorID]
		if ok {
			continue
		}

		checked[extractorID] = struct{}{}
		extractors = append(extractors, b.result.ExtractorSnapshotMap[extractorID])
	}

	ret.Files = files
	ret.Extractors = extractors

	return ret, nil
}

type relationTuple struct {
	files      []*uint
	extractors []uint
	isDel      bool
}

type relationInfo map[string]*relationTuple

func (i relationInfo) add(relationName string, file *uint, extractor uint, isAdd bool) {
	val, ok := i[relationName]
	if !ok {
		val = &relationTuple{}
		i[relationName] = val
	}

	val.files = append(val.files, file)
	val.isDel = val.isDel || !isAdd
	val.extractors = append(val.extractors, extractor)
}

func (i relationInfo) maxCountNotDeleteName() string {
	maxCount := 0
	maxCountName := ""

	for name, tuple := range i {

		if tuple.isDel {
			continue
		}

		if len(tuple.files) > maxCount {
			maxCount = len(tuple.files)
			maxCountName = name
		}
	}
	return maxCountName
}

func (i relationInfo) files(relationName string) map[uint]struct{} {
	raw := i[relationName].files
	ret := make(map[uint]struct{}, len(raw))

	for _, ptr := range raw {
		if ptr != nil {
			ret[*ptr] = struct{}{}
		}
	}

	return ret
}

func (i relationInfo) extractors(relationName string) map[uint]struct{} {
	target := i[relationName]
	ret := make(map[uint]struct{}, len(target.extractors))
	for _, eid := range target.extractors {
		ret[eid] = struct{}{}
	}
	return ret
}

/*
collectRelations 统计某个实体的所有出（入）边。返回值的 key 为相邻节点的名字，value为与改变匹配的 Relation 的统计信息。
*/
func (b *kgBuilder) collectRelations(relations []metadata.Relation, nodeIDs []uint) (map[string]relationInfo, error) {
	ret := make(map[string]relationInfo)

	for i := 0; i < len(relations); i++ {

		var targetEntity metadata.Entity
		if err := b.tx.Take(&targetEntity, nodeIDs[i]).Error; err != nil {
			return nil, utils.WrapError(err, fmt.Sprintf("select entity with id=[%d] fail", nodeIDs[i]))
		}

		_, ok := b.result.SelectedEntities[targetEntity.Name]
		if !ok {
			continue
		}

		info, ok := ret[targetEntity.Name]
		if !ok {
			info = make(relationInfo)
			ret[targetEntity.Name] = info
		}

		fileID, err := b.getFileIDByTextID(relations[i].TextID)
		if err != nil {
			return nil, utils.WrapError(err, "get fileID with textID fail")
		}

		extractorID := b.result.ExtractorSnapshotMap[relations[i].ExtractorID]

		info.add(relations[i].Name, fileID, extractorID, relations[i].Type == metadata.EntityTypeAdd)
	}

	return ret, nil
}

/*
buildOutForNode 通过同一个 Node 的 Entity 列表，构建出边信息。
*/
func (b *kgBuilder) buildOutForNode(entities []metadata.Entity) (metadata.SchemaNodeOut, error) {
	ret := metadata.SchemaNodeOut{}

	ids := make([]uint, len(entities))
	for i := 0; i < len(ids); i++ {
		ids[i] = entities[i].ID
	}

	var outRelations []metadata.Relation
	err := b.tx.
		Where("head_id in ?", ids).
		Find(&outRelations).Error
	if err != nil {
		return ret, utils.WrapError(err, "select relations with headIDs fail")
	}

	next := make([]uint, len(outRelations))
	for i := 0; i < len(outRelations); i++ {
		next[i] = outRelations[i].TailID
	}

	out, err := b.collectRelations(outRelations, next)
	if err != nil {
		return ret, utils.WrapError(err, "collect relations fail")
	}

	nextNodes := map[string]metadata.RelationInfo{}
	for nextName, info := range out {
		relationName := info.maxCountNotDeleteName()

		if len(relationName) == 0 {
			continue
		}

		fids := info.files(relationName)
		files := make([]metadata.FileInfo, 0, len(fids))
		for fid := range fids {
			name, typ, err := b.getFileNameAndTypeByID(fid)
			if err != nil {
				return ret, utils.WrapError(err, "get fileName with id fail")
			}
			files = append(files, metadata.FileInfo{
				FileID:   fid,
				FileName: name,
				FileType: typ,
			})
		}

		eids := info.extractors(relationName)
		extractors := make([]uint, 0, len(eids))
		for eid := range eids {
			extractors = append(extractors, eid)
		}

		nextNodes[nextName] = metadata.RelationInfo{
			Name:       relationName,
			Files:      files,
			Extractors: extractors,
		}
	}

	ret.NextNodes = nextNodes

	return ret, nil
}

func (b *kgBuilder) createNode(name string) error {
	var entities []metadata.Entity
	err := b.tx.
		Where("name = ? and extractor_id in ?", name, b.config.ExtractorIDList).
		Find(&entities).Error
	if err != nil {
		return utils.WrapError(err, "select entities with name and extractorIds fail")
	}

	// 出边
	nodeOut, err := b.buildOutForNode(entities)
	if err != nil {
		return utils.WrapError(err, "build outs for node fail")
	}

	// 溯源
	nodeSource, err := b.buildSourceForNode(entities)
	if err != nil {
		return utils.WrapError(err, "build source for node fail")
	}

	// 创建 node
	node := metadata.Node{
		Model:      gorm.Model{},
		Extra:      metadata.Extra{},
		BuildID:    b.result.BuildID,
		Name:       name,
		OutJSON:    nodeOut.ToJSON(),
		SourceJSON: nodeSource.ToJSON(),
	}

	if err := b.tx.Create(&node).Error; err != nil {
		return utils.WrapError(err, "insert node fail")
	}

	return nil
}

func (b *kgBuilder) createNodes(entities map[string]struct{}) error {

	for entity := range entities {

		err := b.createNode(entity)
		if err != nil {
			return utils.WrapError(err, "create node fail")
		}
	}

	return nil
}
