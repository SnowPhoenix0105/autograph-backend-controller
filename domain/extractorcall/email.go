package extractorcall

import (
	"autograph-backend-controller/utils"
	emailutils "autograph-backend-controller/utils/email"
	"fmt"
	"strings"
)

const extractEmailHTMLTemplate = `
<h1>关系抽取成功</h1>
<p>文件名：%s</p>
<p>不重复实体数量：%d</p>
<p>不重复SPO三元组数量：%d</p>

<h2>实体列表</h2>
<p>%s</p>

<h2>SPO列表</h2>
<p>%s</p>

<p></p>
<p>更多信息请前往系统查看</p>
`

type spoCollection map[string]map[string]map[string]struct{}

func (s spoCollection) Add(head, tail, relation string) {
	tailRelEntry, exist := s[head]
	if !exist {
		tailRelEntry = make(map[string]map[string]struct{})
		s[head] = tailRelEntry
	}

	relSet, exist := tailRelEntry[tail]
	if !exist {
		relSet = make(map[string]struct{})
		tailRelEntry[tail] = relSet
	}

	relSet[relation] = struct{}{}
}

func (s spoCollection) Plain() []string {
	ret := make([]string, 0)
	for head, tailRelEntry := range s {
		for tail, relSet := range tailRelEntry {
			for relation := range relSet {
				ret = append(ret, fmt.Sprintf("(%s)-[%s]->(%s)", head, relation, tail))
			}
		}
	}
	return ret
}

func sendExtractTaskResultEmail(email, filename string, entityNameList []string, spo spoCollection) error {
	err := emailutils.SendHtml(email, "【知识图谱管理系统】关系抽取完成", renderExtractResultPage(filename, entityNameList, spo))
	if err != nil {
		return utils.WrapErrorf(err, "send email to [%s] fail", email)
	}

	return nil
}

func renderExtractResultPage(filename string, entityList []string, spo spoCollection) string {
	entityListStr := strings.Join(entityList, "<br/>")
	spoList := spo.Plain()
	spoListStr := strings.Join(spoList, "<br/>")
	return fmt.Sprintf(extractEmailHTMLTemplate, filename, len(entityList), len(spoList), entityListStr, spoListStr)
}
