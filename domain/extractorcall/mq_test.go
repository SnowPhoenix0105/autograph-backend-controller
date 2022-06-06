package extractorcall

import (
	"autograph-backend-controller/logging"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestRabbitMQManager(t *testing.T) {
	logging.SetDefaultConfig(logging.GenerateTestConfig(t))

	cfg := GenerateTestMQConnectionConfig()
	manager, err := newRabbitMQManager(cfg.ToURL(), []string{"First", "Second"})
	require.Nil(t, err)
	defer func(manager *rabbitMQManager) {
		assert.Nil(t, manager.Close())
	}(manager)

	err = manager.ListenOn("First", func(msg *amqp.Delivery) error {
		var sendObj SendSchema
		if err := json.Unmarshal(msg.Body, &sendObj); err != nil {
			return err
		}

		return manager.SendObjectByJSON("Second", ReceiveSchema{
			Text:    sendObj.Text,
			TextID:  sendObj.TextID,
			Offset:  sendObj.Offset,
			SPOList: nil,
		})
	})
	require.Nil(t, err)

	mutex := sync.Mutex{}
	var result []ReceiveSchema

	err = manager.ListenOn("Second", func(msg *amqp.Delivery) error {
		var receiveObj ReceiveSchema
		if err := json.Unmarshal(msg.Body, &receiveObj); err != nil {
			return err
		}

		mutex.Lock()
		result = append(result, receiveObj)
		mutex.Unlock()
		return nil
	})

	for i := 0; i < 24; i++ {
		err = manager.SendObjectByJSON("First", SendSchema{
			Text:   fmt.Sprintf("Text%d", i),
			TextID: uint(i),
			Offset: i * 10,
		})
		require.Nil(t, err)
	}

	ticker := time.NewTicker(200 * time.Microsecond)

	for i := 0; i < 100; i++ {
		mutex.Lock()
		length := len(result)
		mutex.Unlock()
		t.Logf("[%d] length=%d", i, length)
		if length >= 24 {
			break
		}
		<-ticker.C
	}

	ticker.Stop()
	assert.Equal(t, 24, len(result))

	for i, res := range result {
		assert.Equal(t, fmt.Sprintf("Text%d", i), res.Text)
		assert.Equal(t, uint(i), res.TextID)
		assert.Equal(t, i*10, res.Offset)
		assert.Zero(t, len(res.SPOList))
	}
}
