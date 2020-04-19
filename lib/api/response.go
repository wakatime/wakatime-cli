package api

import (
	"encoding/json"
	"fmt"
)

type Result struct {
	Status    int
	Heartbeat Heartbeat
}

func parseResults(data []byte) ([]Result, error) {
	var responseBody struct {
		Responses [][]json.RawMessage `json:"responses"`
	}

	err := json.Unmarshal(data, &responseBody)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal response body: %s", err)
	}

	var results []Result
	for _, r := range responseBody.Responses {
		type heartbeatData struct {
			Data *Heartbeat `json:"data"`
		}

		var result Result
		err := json.Unmarshal(r[0], &heartbeatData{Data: &result.Heartbeat})
		if err != nil {
			return nil, fmt.Errorf("failed json unmarshal heartbeat: %s", err)
		}

		err = json.Unmarshal(r[1], &result.Status)
		if err != nil {
			return nil, fmt.Errorf("failed json unmarshal status: %s", err)
		}

		results = append(results, result)
	}

	return results, nil
}
