package extractorcall

import (
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/utils"
	"encoding/json"
	"errors"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

func Send(info SendSchema) error {
	return globalMQManager.SendObjectByJSON(QueueExtractorInput, info)
}

type rangeTuple struct {
	begin int
	end   int
}

func saveReceivedData(tx *gorm.DB, data ReceiveSchema) error {
	entityIndex := make(map[rangeTuple]*metadata.Entity)

	// 保存实体
	for _, spo := range data.SPOList {
		headOrigin := spo.HeadEntity

		headKey := rangeTuple{
			begin: headOrigin.Begin,
			end:   headOrigin.End,
		}

		if entityIndex[headKey] == nil {
			head := metadata.Entity{
				Name:        headOrigin.Name,
				Type:        metadata.EntityTypeAdd,
				ExtractorID: spo.ExtractorId,
				TextID:      utils.UintToPtr(data.TextID),
				TaskID:      data.TaskID,
			}

			if err := tx.Create(&head).Error; err != nil {
				return utils.WrapError(err, "save head-entity to db fail")
			}

			entityIndex[headKey] = &head
		}

		tailOrigin := spo.TailEntity

		tailKey := rangeTuple{
			begin: tailOrigin.Begin,
			end:   tailOrigin.End,
		}

		if entityIndex[tailKey] == nil {
			tail := metadata.Entity{
				Name:        tailOrigin.Name,
				Type:        metadata.EntityTypeAdd,
				ExtractorID: spo.ExtractorId,
				TextID:      utils.UintToPtr(data.TextID),
				TaskID:      data.TaskID,
			}

			if err := tx.Create(&tail).Error; err != nil {
				return utils.WrapError(err, "save tail-entity to db fail")
			}

			entityIndex[tailKey] = &tail
		}
	}

	// 保存关系
	for _, spo := range data.SPOList {
		headOrigin := spo.HeadEntity

		headKey := rangeTuple{
			begin: headOrigin.Begin,
			end:   headOrigin.End,
		}

		tailOrigin := spo.TailEntity

		tailKey := rangeTuple{
			begin: tailOrigin.Begin,
			end:   tailOrigin.End,
		}

		head := entityIndex[headKey]
		tail := entityIndex[tailKey]

		relation := metadata.Relation{
			Name:        spo.Relation,
			Type:        metadata.EntityTypeAdd,
			ExtractorID: spo.ExtractorId,
			TextID:      utils.UintToPtr(data.TextID),
			HeadID:      head.ID,
			TailID:      tail.ID,
			TaskID:      data.TaskID,
		}
		if err := tx.Create(&relation).Error; err != nil {
			return utils.WrapError(err, "save relation to db fail")
		}
	}

	// 更新任务
	if data.TaskID != nil {

		var taskItem metadata.ExtractTaskItem
		err := tx.Where(&metadata.ExtractTaskItem{
			ExtractTaskID: *data.TaskID,
			TextID:        data.TextID,
		}).First(&taskItem).Error

		if err != nil {
			return utils.WrapError(err, "update task fail")
		}

		taskItem.Status = metadata.TaskStatusDone
		tx.Save(&taskItem)
	}

	return nil
}

var errEmptyBody = errors.New("message body is empty")

func buildReceive(getMetadataDatabase func() *gorm.DB) func(msg *amqp.Delivery) error {
	return func(msg *amqp.Delivery) error {
		if len(msg.Body) == 0 {
			return utils.WrapError(errEmptyBody, "msg.Body is empty")
		}

		var data ReceiveSchema
		if err := json.Unmarshal(msg.Body, &data); err != nil {
			return utils.WrapErrorf(err, "json unmarshal fail with[%#v]", msg.Body)
		}

		err := getMetadataDatabase().Transaction(func(tx *gorm.DB) error {
			return saveReceivedData(tx, data)
		})

		if err != nil {
			return utils.WrapError(err, "save data to db fail")
		}

		return err
	}
}
