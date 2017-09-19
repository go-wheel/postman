package httptask

// TaskResult 单任务结果
type TaskResult struct {
	TaskItemID string `json:"itemid"`
	Result     string `json:"result"`
	Error      string `json:"error"`
	Time       string `json:"time"`
}

// TaskGroupResult 任务组结果
type TaskGroupResult struct {
	GroupID string       `json:"groupid"`
	Results []TaskResult `json:"results"`
}
