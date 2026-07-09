package executor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"dscheduler/model"
)

// RunJob 调度中心向执行器下发任务（HTTP POST /run）
func RunJob(address string, p model.TriggerParam) (model.TriggerResult, error) {
	data, _ := json.Marshal(p)
	req, _ := http.NewRequest(http.MethodPost, "http://"+address+"/run", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return model.TriggerResult{Code: 500, Msg: err.Error()}, err
	}
	defer resp.Body.Close()
	var r model.TriggerResult
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return model.TriggerResult{Code: 500, Msg: err.Error()}, err
	}
	return r, nil
}

// Beat 探测执行器存活
func Beat(address string) (model.TriggerResult, error) {
	req, _ := http.NewRequest(http.MethodPost, "http://"+address+"/beat", nil)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return model.TriggerResult{Code: 500, Msg: err.Error()}, err
	}
	defer resp.Body.Close()
	var r model.TriggerResult
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return model.TriggerResult{Code: 500, Msg: err.Error()}, err
	}
	return r, nil
}
