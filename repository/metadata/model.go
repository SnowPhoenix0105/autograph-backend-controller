package metadata

import (
	"database/sql"
	"gorm.io/gorm"
)

/*
Extra 用于扩展信息，或者保存多态的信息，通过JSON格式。不直接单独作为一个数据库对象，类似gorm.Model。

	ExtraType 标记JSON的schema；
	ExtraJSON 额外信息的JSON主体；
*/
type Extra struct {
	ExtraType sql.NullString `gorm:"type:varchar(16)"`
	ExtraJSON sql.NullString `gorm:"type:text"`
}

//////////////////////////////// 元信息，包含详细、原始的生产信息 ////////////////////////////////////

/*
ExtractTask 描述了一次关系抽取任务

	Name 任务名，可以为文件名
	Email 任务结束后通知的邮箱
*/
type ExtractTask struct {
	gorm.Model
	Extra

	Name  string
	Email string

	ExtractItem []ExtractTaskItem `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

/*
ExtractTaskItem 描述一次关系抽取任务中一个句子的结果
*/
type ExtractTaskItem struct {
	gorm.Model
	Extra

	Status uint `gorm:"comment:DOING=1,DONE=2,FAIL=3"`

	TextID        uint
	ExtractTaskID uint
}

/*
File 记录了文件的元信息。

	Extra 为扩展预留；
	Type 文件的类型；
	URL 文件的内部路径，如果是用文件系统储存的，则为文件路径；
	Name 文件名，用于下载文件后的命名；
	Hash 文件的MD5摘要，用于判断重复；

	ProducedText 多对一关系，表示由本文件产生的文本；
*/
type File struct {
	gorm.Model
	Extra
	Type string `gorm:"type:varchar(8) not null"`
	URL  string `gorm:"type:varchar(128)"`
	Name string `gorm:"type:varchar(32)"`
	Hash []byte `gorm:"type:binary(16);index:idx_files_hash"`

	ProducedText []Text `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

/*
Text 记录了文本的元信息。

	Extra 为扩展预留；
	Content 文本内容；

	FileID 一对多关系，此文本来源的文件；
	ProducedEntities 多对一关系，由此文本抽取出的实体；
	ProducedRelations 多对一关系，由此文本抽取出的关系；
*/
type Text struct {
	gorm.Model
	Extra
	Content string `gorm:"type:text not null"`

	FileID            *uint
	ExtractItems      []ExtractTaskItem `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ProducedEntities  []Entity          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ProducedRelations []Relation        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

/*
Extractor 记录了抽取器（包括算法模型和人为干预）模型的元信息。

	Extra 为扩展预留，当Type为Model时，记录了模型的部署信息；
	Name 模型名称；
	Desc 模型描述；
	Type 表示抽取器是算法模型还是人为干预，1表示算法模型，2表示人为干预。

	ProducedEntities 多对一关系，由此模型抽取出的实体；
	ProducedRelations 多对一关系，由此模型抽取出的关系；
*/
type Extractor struct {
	gorm.Model
	Extra
	Name string `gorm:"type:varchar(64) not null"`
	Desc string
	Type uint `gorm:"comment:Model=1, HumanIntervention=2"`
	URL  string

	ProducedEntities  []Entity   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ProducedRelations []Relation `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

/*
Entity 记录了实体的元信息。

	Extra 为扩展预留；
	Name 实体名；
	Type 表示该元信息表示增加实体还是删除实体，1表示增加，2表示删除（一般来说只有Extractor是人为干预时才可能为删除）

	TextID	一对多关系，此实体来源的文本；
	ExtractorID 一对多关系，抽取出次实体的模型；
*/
type Entity struct {
	gorm.Model
	Extra
	Name string `gorm:"type:varchar(32) not null;index:idx_entities_name"`
	Type uint   `gorm:"comment:Add=1,Del=2"`

	ExtractorID  uint
	TaskID       *uint
	TextID       *uint
	HeadEntityOf []Relation `gorm:"foreignKey:HeadID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TailEntityOf []Relation `gorm:"foreignKey:TailID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

/*
Relation 记录了关系的元信息。

	Extra 为扩展预留；
	Name 关系名
	Type 表示该元信息表示增加实体还是删除实体，1表示增加，2表示删除（一般来说只有Extractor是人为干预时才可能为删除）
*/
type Relation struct {
	gorm.Model
	Extra
	Name string `gorm:"type:varchar(32) not null;index:idx_relations_name"`
	Type uint   `gorm:"comment:Add=1,Del=2"`

	ExtractorID uint
	TextID      *uint
	TaskID      *uint
	HeadID      uint `gorm:"type:bigint unsigned not null"`
	TailID      uint `gorm:"type:bigint unsigned not null"`
}

/////////////////////////////// 构建信息，包含构建知识图谱的信息///////////////////////////////////

/*
Build 记录了一次图谱构建的元信息。

	Extra 为扩展预留；
	Desc 构建的描述；
*/
type Build struct {
	gorm.Model
	Extra
	Desc string
}

/*
BuildExtractor 记录了一次知识图谱构建中抽取器的快照。

	Extra 为扩展预留；
	BuildID 构建号；
	Name 抽取器名；
	Desc 抽取器描述；
	Type 表示抽取器是算法模型还是人为干预，1表示算法模型，2表示人为干预。
*/
type BuildExtractor struct {
	gorm.Model
	Extra
	BuildID uint   `gorm:"index:idx_name_version"`
	Name    string `gorm:"type:varchar(64) not null;index:idx_name_version"`
	Desc    string
	Type    uint `gorm:"comment:Model=1, HumanIntervention=2"`
}

/*
Node 记录了一次知识图谱构建中的节点，以及节点出边和入边。反范式设计，因为只需要一次写入。

	BuildID 构建号；
	Name 实体名；
	OutJSON 通过JSON字符串描述节点的出边，描述的是Node与Node的多对多关系；
	// InJSON 通过JSON字符串描述节点的入边，描述的是Node与Node的多对多关系；
	SourceJSON 通过JSON字符串描述节点的来源信息，描述的是Node与File、Node与BuildExtractor的两组多对多关系；
*/
type Node struct {
	gorm.Model
	Extra
	BuildID uint   `gorm:"index:idx_name_version"`
	Name    string `gorm:"type:varchar(32) not null;index:idx_name_version"`
	OutJSON string
	// InJSON     string
	SourceJSON string
}

///////////////////////////// 其它信息，包含系统所需的各种数据 /////////////////////////////////////////
