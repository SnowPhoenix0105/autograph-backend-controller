package extractorcall

import (
	"autograph-backend-controller/utils/email"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSendExtractTaskResultEmail(t *testing.T) {
	email.Init(email.GenerateTestConfig())

	entList := make([]string, 0)
	for i := 0; i < 20; i++ {
		entList = append(entList, fmt.Sprintf("ent_%d", i))
	}

	spo := spoCollection{}
	for i, ent := range entList {
		for j := i + 1; j < len(entList); j++ {
			spo.Add(ent, entList[j], fmt.Sprintf("rel_%d", i*j))
		}
	}

	err := sendExtractTaskResultEmail("autograph_receiver@163.com", "test", entList, spo)
	assert.Nil(t, err)
}
