package extractorcall

type SendSchema struct {
	Text   string `json:"text"`
	TextID uint   `json:"text_id"`
	TaskID uint   `json:"task_id"`
	Offset int    `json:"offset"`
}

type ReceiveSchemaEntity struct {
	Name  string `json:"name"`
	Begin int    `json:"begin"`
	End   int    `json:"end"`
}

type ReceiveSchemaSPO struct {
	ExtractorId uint                `json:"extractor_id"`
	Relation    string              `json:"relation"`
	HeadEntity  ReceiveSchemaEntity `json:"head_entity"`
	TailEntity  ReceiveSchemaEntity `json:"tail_entity"`
}

type ReceiveSchema struct {
	Text    string             `json:"text"`
	TextID  uint               `json:"text_id"`
	TaskID  *uint              `json:"task_id"`
	Offset  int                `json:"offset"`
	SPOList []ReceiveSchemaSPO `json:"spo_list"`
}
