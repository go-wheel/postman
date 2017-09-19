package httptask

import (
	"encoding/json"

	"github.com/lets-go-go/logger"
)

// Parse parse
func Parse(content []byte) *HTTPTasks {

	var tasks HTTPTasks

	if err := json.Unmarshal(content, &tasks); err != nil {
		logger.Errorf("task unmarshal error:%+v", err)
		return nil
	}

	return &tasks
}

// FormData http post request form data
type FormData struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Urlencoded http post request urlencoded
type Urlencoded struct {
	Type   string `json:"type"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	Start  int64  `json:"start"`
	Length int64  `json:"length"`
}

// Body http post request form data
type Body struct {
	Mode       string       `json:"mode"`
	Formdata   []FormData   `json:"formdata"`
	Urlencoded []Urlencoded `json:"urlencoded"`
	Value      string       `json:"value"`
}

// Header http post request header
type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Request http request
type Request struct {
	URL    string   `json:"url"`
	Method string   `json:"method"`
	Header []Header `json:"header"`
	Body   *Body    `json:"body"`
}

// RequestItem one request item
type RequestItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Request     Request `json:"request"`
}

// TaskInfo task basic information
type TaskInfo struct {
	Name        string `json:"name"`
	TaskGroupID string `json:"_postman_id"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
}

// HTTPTasks http tasks
type HTTPTasks struct {
	Info  TaskInfo      `json:"info"`
	Items []RequestItem `json:"item"`
}
