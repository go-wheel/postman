package backendhttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-wheel/postman/httptask"
	"github.com/go-wheel/postman/rangeupload"

	"github.com/lets-go-go/logger"
)

var (
	newTaskList map[string]bool
	mu          sync.Mutex // guards fileList
)

func init() {
	newTaskList = make(map[string]bool)
}

// LoadAllOldTasks Load All Tasks from task path
func LoadAllOldTasks() {

	go func() {
		if taskList, err := ioutil.ReadDir(taskPath); err == nil {
			handleTaskList(taskList)
		} else {
			logger.Fatalf("read task dir error=%v", err)
			return
		}

		logger.Debugf("all old task done")

		HandleNewTask()
	}()
}

// 处理普通http请求列表
func handleTaskList(taskList []os.FileInfo) {
	for _, v := range taskList {
		logger.Infof("current task file=%s", v.Name())

		if v.IsDir() {
			if v.Name() == "range" {
				handleRangeUploadList()
			}
		} else {
			filePath := path.Join(taskPath, v.Name())
			handleTaskFile(filePath)
		}
	}
}

// 处理单个http请求文件，可能有多个任务
func handleTaskFile(taskFile string) {
	content, err := ioutil.ReadFile(taskFile)
	if err != nil {
		logger.Debugf("read task file error:%+v", err)
		return
	}

	groupResults, _ := handleTaskGroup(content)
	if err != nil {
		logger.Errorf("handleRangeUploadTask error:%+v", err)
		AppendNewTask(taskFile)
		return
	}

	completeTask(taskFile, groupResults)

}

// 处理断点续传任务列表
func handleRangeUploadList() {
	rangeUploadList, err := ioutil.ReadDir(rangeUploadTaskPath)
	if err != nil {
		logger.Errorf("do range upload task list error=%+v", err)
	}
	for _, v := range rangeUploadList {
		logger.Infof("current range upload task file=%s", v.Name())

		if !v.IsDir() {
			filePath := path.Join(rangeUploadTaskPath, v.Name())
			handleRangeUploadTask(filePath)
		}
	}
}

// 处理断点续传任务，单个任务
func handleRangeUploadTask(taskFile string) {
	taskGroupResult, err := rangeupload.Upload(rangeUploadTmplPath, taskFile)
	if err != nil {
		logger.Errorf("handleRangeUploadTask error:%+v", err)
		AppendNewTask(taskFile)
		return
	}

	completeTask(taskFile, taskGroupResult)
}

// AppendNewTask Append New Task
func AppendNewTask(fileName string) {
	mu.Lock()
	defer mu.Unlock()
	newTaskList[fileName] = false

	logger.Debugf("add new task file=%s", fileName)
}

// HandleNewTask Handle New Task
func HandleNewTask() {

	timeout := 5 * time.Second
	checkTimer := time.NewTimer(timeout)

	go func() {
		for {
			select {
			case <-checkTimer.C:

				for filePath, done := range newTaskList {
					if !done {
						handleNewTask(filePath)
					}
				}

				checkTimer.Reset(timeout)
			}
		}
	}()
}

// 处理新任务
func handleNewTask(filePath string) {
	logger.Debugf("current file=%v", filePath)
	if strings.Index(filePath, rangeUploadTaskPath) != -1 {
		handleRangeUploadTask(filePath)
	} else {
		handleTaskFile(filePath)
	}

}

// 普通http任务
func handleTaskGroup(content []byte) (httptask.TaskGroupResult, error) {
	var groupResults httptask.TaskGroupResult
	var err error
	if httptasks := httptask.Parse(content); httptasks != nil {
		groupResults.GroupID = httptasks.Info.Name
		for _, item := range httptasks.Items {
			var taskresult httptask.TaskResult

			taskresult.TaskItemID = item.Name

			if result, err := httptask.DoRequest(item); err != nil {
				taskresult.Error = fmt.Sprintf("%v", err)
			} else {
				taskresult.Result = result
			}

			taskresult.Time = time.Now().Format("20060102150405")
			groupResults.Results = append(groupResults.Results, taskresult)
		}
	}
	return groupResults, err
}

func completeTask(taskFile string, groupResults httptask.TaskGroupResult) {

	resultData, _ := json.Marshal(groupResults)
	fileName := filepath.Base(taskFile)

	resultFileName := fmt.Sprintf("result-%s", fileName)
	resultFile := path.Join(resultPath, resultFileName)

	completeFile := path.Join(resultPath, fileName)
	if err := os.Rename(taskFile, completeFile); err != nil {
		logger.Warnf("move task file failed:%+v", completeFile)
	}

	if err := ioutil.WriteFile(resultFile, resultData, 0777); err != nil {
		logger.Warnf("write result file failed:%+v", fileName)
	}

	mu.Lock()
	defer mu.Unlock()

	newTaskList[taskFile] = true
}
