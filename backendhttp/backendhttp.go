package backendhttp

import (
	"flag"
	"path"
	"path/filepath"
)

var (
	taskPath            string
	rangeUploadTaskPath string
	rangeUploadTmplPath string
	resultPath          string
)

// "args": ["-taskpath=./example/req", "-resultpath=./example/result", "-rangetpl=./example/tpl/range-tpl.json"],
func init() {
	flag.StringVar(&taskPath, "taskpath", "./example/req", "taskpath")
	flag.StringVar(&rangeUploadTmplPath, "rangetpl", "./example/tpl/range-tpl.json", "range templete path")
	flag.StringVar(&resultPath, "resultpath", "./example/result", "resultpath")
}

func Init() {

	// 断点续传目录在任务的range子目录
	rangeUploadTaskPath = path.Join(taskPath, "range")
	rangeUploadTaskPath = path.Clean(rangeUploadTaskPath)
	rangeUploadTaskPath = filepath.Clean(rangeUploadTaskPath)
	LoadAllOldTasks()

	StartTaskWatcher()
}
