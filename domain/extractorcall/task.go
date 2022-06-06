package extractorcall

import (
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/utils"
	"time"
)

func DoExtraction(textIDs []uint, texts []string, name string, notifyEmail string) {
	db := metadata.DatabaseRaw()

	task := metadata.ExtractTask{
		Name:  name,
		Email: notifyEmail,
	}
	db.Create(&task)

	// 发送抽取信号
	for i, id := range textIDs {
		item := metadata.ExtractTaskItem{
			Status:        metadata.TaskStatusDoing,
			TextID:        id,
			ExtractTaskID: task.ID,
		}
		err := db.Create(&item).Error
		if err != nil {
			logging.Default().WithError(err).Errorf("create task{textID=%d, text=[%#v]} fail: %s", id, texts[i], err.Error())
		}

		err = Send(SendSchema{
			Text:   texts[i],
			TextID: id,
			TaskID: task.ID,
			Offset: 0,
		})

		if err != nil {
			logging.Default().WithError(err).Errorf("send text{id=%d, content=[%#v]} fail: %s", id, texts[i], err.Error())
			item.Status = metadata.TaskStatusFail
			db.Save(&item)
		}
	}

	// 监听抽取结果
	monitorOnExtractTask(task.ID)

	logging.Default().Infof("Task{ID=%d, Name=%#v, Email=%#v} Success", task.ID, task.Name, task.Email)
}

func monitorOnExtractTask(taskID uint) {
	tick := time.Tick(1 * time.Minute)

	for {
		<-tick

		var count int64

		err := metadata.DatabaseRaw().Model(&metadata.ExtractTaskItem{}).Where(metadata.ExtractTaskItem{
			ExtractTaskID: taskID,
			Status:        metadata.TaskStatusDoing,
		}).Count(&count).Error

		if err != nil {
			logging.Default().WithError(err).Errorf("check extract-task[%d] fail: %s", taskID, err.Error())
			continue
		}

		if count == 0 {
			logging.Default().Infof("check extract-task[%d]: finished!", taskID)

			break
		}

		logging.Default().Infof("check extract-task[%d]: not finish yet", taskID)
	}

	for i := 0; i < 3; i++ {
		err := sendExtractResult(taskID)
		if err == nil {
			break
		}

		logging.Default().WithError(err).Errorf("send extract result fail: %s", err.Error())
	}
}

func sendExtractResult(taskID uint) error {
	var task metadata.ExtractTask
	err := metadata.DatabaseRaw().Find(&task, taskID).Error
	if err != nil {
		return utils.WrapErrorf(err, "select task[%d] metadata fail", taskID)
	}

	if len(task.Email) == 0 {
		logging.Default().Warnf("task[%d] has no notifying email", taskID)
		return nil
	}

	// 统计task产生的实体
	var entities []metadata.Entity
	err = metadata.DatabaseRaw().Where(&metadata.Entity{
		TaskID: utils.UintToPtr(taskID),
	}).Find(&entities).Error

	if err != nil {
		return utils.WrapErrorf(err, "select entities metadata of task[%d] fail", taskID)
	}

	entityNameList := make([]string, len(entities))
	entityIndex := map[uint]int{} // EntityID -> index of var entities
	entityNameSet := make(map[string]struct{})
	entityUniqueNameList := make([]string, 0)

	// 异步构建索引
	indexingFinish := make(chan struct{})
	go func(outEntityList []string, outEntityIndex map[uint]int, finishChan chan<- struct{}) {
		for i, entity := range entities {
			entityIndex[entity.ID] = i
			entityNameList[i] = entity.Name
			_, exist := entityNameSet[entity.Name]
			if !exist {
				entityUniqueNameList = append(entityUniqueNameList, entity.Name)
				entityNameSet[entity.Name] = struct{}{}
			}
		}
		close(finishChan)
	}(entityNameList, entityIndex, indexingFinish)

	// 统计task产生的SPO
	var relations []metadata.Relation
	err = metadata.DatabaseRaw().Where(metadata.Relation{
		TaskID: utils.UintToPtr(taskID),
	}).Find(&relations).Error

	if err != nil {
		return utils.WrapErrorf(err, "select relations metadata of task[%d] fail", taskID)
	}

	<-indexingFinish

	spo := spoCollection{}
	for _, relation := range relations {
		spo.Add(entityNameList[entityIndex[relation.HeadID]], entityNameList[entityIndex[relation.TailID]], relation.Name)
	}

	// 发送抽取结果
	err = sendExtractTaskResultEmail(task.Email, task.Name, entityUniqueNameList, spo)
	if err != nil {
		return utils.WrapErrorf(err, "send email fail")
	}

	return nil
}
