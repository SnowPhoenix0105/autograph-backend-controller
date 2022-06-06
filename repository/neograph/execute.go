package neograph

import (
	"autograph-backend-controller/utils"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func Execute(cypher string, param map[string]interface{}) ([]*neo4j.Record, error) {
	return execute(globalDriver, cypher, param)
}

func execute(driver neo4j.Driver, cypher string, param map[string]interface{}) ([]*neo4j.Record, error) {
	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	res, err := session.Run(cypher, param)
	if err != nil {
		return nil, utils.WrapErrorf(err, "execute [%#v] fail", cypher)
	}

	records, err := res.Collect()
	if err != nil {
		return nil, utils.WrapError(err, "collect fail")
	}

	return records, nil
}
