package metadata

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigration(t *testing.T) {
	cfg := GenerateTestConfig()
	cfg.CheckMigration = true
	_, err := CreateDatabase(cfg)
	assert.Nil(t, err)
}
