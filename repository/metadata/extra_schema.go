package metadata

import "encoding/json"

func toJSON(schema interface{}) string {
	bytes, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

type SchemaExtractorExtraDeploy struct {
	DeployYAML string
}

type RelationInfo struct {
	Name       string     `json:"name"`
	Files      []FileInfo `json:"files"`
	Extractors []uint     `json:"extractors"`
}

type SchemaNodeOut struct {
	NextNodes map[string]RelationInfo `json:"next_nodes"`
}

func (no *SchemaNodeOut) ToJSON() string {
	return toJSON(no)
}

type FileInfo struct {
	FileID   uint   `json:"file_id"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
}

type SchemaNodeSource struct {
	Files      []FileInfo `json:"files"`
	Extractors []uint     `json:"extractors"`
}

func (ns *SchemaNodeSource) ToJSON() string {
	return toJSON(ns)
}
