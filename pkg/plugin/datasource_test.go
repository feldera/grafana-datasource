package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestQueryData(t *testing.T) {
	ds := Datasource{
		baseUrl: "http://localhost:28080",
		pipeline: "blah",
	}

	query := []byte(`{"queryText": "SELECT * FROM v0 where ts BETWEEN $__timeFrom() AND $__timeTo() LIMIT 10"}`)

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A", JSON: query, TimeRange: backend.TimeRange{From: time.Now(), To: time.Now()}},
			},
		},
	)
	if err != nil {
		t.Error(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}
}
