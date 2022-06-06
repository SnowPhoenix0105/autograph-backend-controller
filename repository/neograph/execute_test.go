package neograph

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestExecute(t *testing.T) {
	driver, err := initDriver(&GenerateTestConfig().Neo4j)
	require.Nil(t, err)
	defer driver.Close()

	err = driver.VerifyConnectivity()
	require.Nil(t, err)

	type TestStruct struct {
		Name string
		Info map[string]interface{}
	}

	inputDesc := TestStruct{
		Name: "test",
		Info: map[string]interface{}{
			"a": 1,
			"b": 2,
		},
	}

	desc, err := json.Marshal(inputDesc)
	require.Nil(t, err)

	ts := time.Now().Unix()

	t.Logf("%d", ts)

	// create e1
	{
		res, err := execute(driver, "create (e1:Test{name: $name, desc: $desc, ts: $ts})", map[string]interface{}{
			"name": "e1",
			"ts":   ts,
			"desc": string(desc),
		})
		require.Nil(t, err)

		for i, record := range res {
			t.Logf("create e1 [%d], %#v -> %#v", i, record.Keys, record.Values)
		}
	}

	// create e1
	{
		res, err := execute(driver, "create (e2:Test{name: $name, desc: $desc, ts: $ts})", map[string]interface{}{
			"name": "e2",
			"ts":   ts,
			"desc": string(desc),
		})
		require.Nil(t, err)

		for i, record := range res {
			t.Logf("create e2 [%d], %#v -> %#v", i, record.Keys, record.Values)
		}
	}

	// create r
	{
		res, err := execute(driver, "match (e1:Test{ts: $ts}), (e2:Test{ts: $ts}) merge (e1)-[r:TestR{ts: $ts}]->(e2)", map[string]interface{}{
			"ts": ts,
		})
		require.Nil(t, err)

		for i, record := range res {
			t.Logf("create r [%d], %#v -> %#v", i, record.Keys, record.Values)
		}
	}

	{
		res, err := execute(driver, "match (e1:Test{ts: $ts})-[r:TestR{ts: $ts}]->(e2:Test{ts: $ts}) return e1, r, e2", map[string]interface{}{
			"ts": ts,
		})
		require.Nil(t, err)

		for i, record := range res {
			t.Logf("query [%d], %#v -> %#v", i, record.Keys, record.Values)
		}
	}
}
