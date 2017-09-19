package rangeupload

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-wheel/postman/httptask"

	"github.com/lets-go-go/logger"
	uuid "github.com/satori/go.uuid"
)

const (
	uploadUniqueCode = "unique-code"
	uploadRange      = "range"
	uploadNow        = "now"
	uploadTotal      = "total"

	headerKeyDataVersion = "Data-Version"
	headerKeyDataRange   = "Data-Range"
	headerKeyDataLength  = "Data-Length"

	urlKeyUniqueCode = "uniqueCode"
	urlKeyVersion    = "version"
)

type rangeUploadParam struct {
	FileSrc    string `json:"fileSrc"`
	TaskID     string `json:"taskId"`
	Token      string `json:"token"`
	FfsID      string `json:"ffsId"`
	FileID     string `json:"fileId"`
	Md5        string `json:"md5"`
	Version    string `json:"version"`
	UniqueCode string `json:"uniqueCode"`
	NowLength  int64  `json:"nowLength"`
	Total      int64  `json:"total"`
}

// Upload 定制断点续传请求
func Upload(tmplFile, taskDataFile string) (httptask.TaskGroupResult, error) {
	var httptasks httptask.HTTPTasks
	content, err := ioutil.ReadFile(tmplFile)
	if err != nil {
		return handleError("", err, "", nil)
	}
	if err = json.Unmarshal(content, &httptasks); err != nil {
		return handleError("", err, "", nil)
	}

	logger.Debugf("tasks=%+v", httptasks.Items)

	uploadTasks := make(map[string]httptask.RequestItem)
	for _, item := range httptasks.Items {
		uploadTasks[item.Name] = item
	}

	var params rangeUploadParam

	content, err = ioutil.ReadFile(taskDataFile)
	if err != nil {
		return handleError("", err, "", nil)
	}

	if err = json.Unmarshal(content, &params); err != nil {
		return handleError("", err, "", nil)
	}

	if params.UniqueCode == "" {
		// step1 新任务，还没有获取UniqueCode
		if item, ok := uploadTasks[uploadUniqueCode]; ok {
			if err := uniqueCode(item, &params); err != nil {
				return handleError("", err, taskDataFile, &params)
			}
		}

	}

	if params.NowLength != 0 {
		// 校验已上传长度
		if item, ok := uploadTasks[uploadNow]; ok {
			if err := getNowLen(item, &params); err != nil {
				return handleError("", err, taskDataFile, &params)
			}
		}
	}

	const blockSize = 1024 * 1024
	var start, length int64

	for {
		// 循环上传
		if params.NowLength == params.Total {
			break
		}

		start = params.NowLength
		length = blockSize
		if blockSize > (params.Total - start) {
			length = params.Total - start
		}

		if item, ok := uploadTasks[uploadRange]; ok {
			uploadSize, err := rangeUploadFile(item, &params, start, length)
			if err != nil {
				return handleError("", err, taskDataFile, &params)
			}

			params.NowLength += uploadSize
		}
	}

	// 上传成功
	if item, ok := uploadTasks[uploadTotal]; ok {
		if err := uploadFinish(item, &params); err != nil {
			return handleError("", err, taskDataFile, &params)
		}
	}

	return handleError(params.TaskID, nil, taskDataFile, &params)

}

func handleError(taskID string, err error, taskDataFile string, params *rangeUploadParam) (httptask.TaskGroupResult, error) {
	var groupResults httptask.TaskGroupResult

	groupResults.GroupID = taskID
	var taskresult httptask.TaskResult

	taskresult.TaskItemID = taskID

	taskresult.Time = time.Now().Format("20060102150405")

	taskresult.Error = fmt.Sprintf("%v", err)

	groupResults.Results = append(groupResults.Results, taskresult)

	newData, _ := json.Marshal(params)

	ioutil.WriteFile(taskDataFile, newData, 0777)

	return groupResults, err
}

func uniqueCode(item httptask.RequestItem, params *rangeUploadParam) error {

	fileHash, err := GetFileHash(params.FileSrc)
	if err != nil {
		return err
	}
	params.Md5 = fileHash.MD5
	params.Total = fileHash.Size
	params.FfsID = fileHash.UUID

	urlParams := url.Values{
		"token": []string{params.Token},
		"type":  []string{"1"},
		"md5":   []string{params.Md5},
		"ffsId": []string{params.FfsID},
	}

	item.Request.URL = fmt.Sprintf("%s?%s", item.Request.URL, urlParams.Encode())

	body, err := httptask.DoRequest(item)
	if err != nil {
		return err
	}

	var response struct {
		Code       string `json:"code"`
		UniqueCode string `json:"uniqueCode"`
	}

	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return err
	}
	if response.Code == "0" {
		params.UniqueCode = response.UniqueCode
	} else {
		return fmt.Errorf("biz error code=%s", response.Code)
	}

	return nil
}

func rangeUploadFile(item httptask.RequestItem, params *rangeUploadParam, start, length int64) (int64, error) {

	var end int64

	end = start + length - 1

	var contentRange, contentLength string

	contentRange = fmt.Sprintf("%d-%d", start, end)
	contentLength = strconv.FormatInt(params.Total, 10)

	if params.Version == "" {
		item.Request.Header = []httptask.Header{
			{Key: "Data-Range", Value: contentRange},
			{Key: "Data-Length", Value: contentLength},
		}
	} else {
		item.Request.Header = []httptask.Header{
			{Key: "Data-Range", Value: contentRange},
			{Key: "Data-Length", Value: contentLength},
			{Key: "Data-Version", Value: params.Version},
		}
	}

	item.Request.Body.Urlencoded = []httptask.Urlencoded{
		{Key: "uniqueCode", Type: "text", Value: params.UniqueCode},
		{
			Key:    "upload",
			Type:   "rangefile",
			Value:  params.FileSrc,
			Start:  start,
			Length: length,
		},
	}

	logger.Tracef("reqeust body=%+v", item.Request.Body.Urlencoded)

	body, err := httptask.DoRequest(item)
	if err != nil {
		return 0, err
	}

	var response struct {
		Code    string `json:"code"`
		FileID  string `json:"fileId"`
		Version string `json:"version"`
	}

	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return 0, err
	}
	if response.Code == "0" {
		params.Version = response.Version
		params.FileID = response.FileID
	} else {
		return 0, fmt.Errorf("biz error code=%s", response.Code)
	}

	return length, nil
}

func getNowLen(item httptask.RequestItem, params *rangeUploadParam) error {
	urlParams := url.Values{
		"version":    []string{params.Version},
		"uniqueCode": []string{params.UniqueCode},
	}

	item.Request.URL = fmt.Sprintf("%s?%s", item.Request.URL, urlParams.Encode())

	body, err := httptask.DoRequest(item)
	if err != nil {
		return err
	}

	var response struct {
		Code      string `json:"code"`
		NowLength string `json:"nowLength"`
	}

	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return err
	}
	if response.Code == "0" {
		params.NowLength, _ = strconv.ParseInt(response.NowLength, 10, 0)
	} else {
		return fmt.Errorf("biz error code=%s", response.Code)
	}

	return nil
}

func uploadFinish(item httptask.RequestItem, params *rangeUploadParam) error {
	urlParams := url.Values{
		"total":      []string{strconv.FormatInt(params.Total, 10)},
		"uniqueCode": []string{params.UniqueCode},
	}

	item.Request.URL = fmt.Sprintf("%s?%s", item.Request.URL, urlParams.Encode())

	body, err := httptask.DoRequest(item)
	if err != nil {
		return err
	}

	var response struct {
		Code string `json:"code"`
	}

	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return err
	}

	if response.Code != "0" {
		return fmt.Errorf("biz error code=%s", response.Code)
	}

	return nil
}

const fileChunk = 8192

//FileHash struct for MD5's
type FileHash struct {
	Filename string
	Size     int64
	MD5      string
	UUID     string
}

//GetFileHash Get takes a file and returns the FileHash struct
func GetFileHash(fileLocation string) (*FileHash, error) {
	var fileHash = FileHash{}

	var err error

	file, err := os.Open(fileLocation)
	if err != nil {
		return &fileHash, err
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	fileHash.Filename = fileInfo.Name()
	fileHash.Size = fileInfo.Size()
	fileHash.UUID = uuid.NewV4().String()

	hash := md5.New()

	blocks := uint64(math.Ceil(float64(fileHash.Size) / float64(fileChunk)))

	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(fileChunk, float64(fileHash.Size-int64(i*fileChunk))))
		buf := make([]byte, blocksize)

		file.Read(buf)
		hash.Write(buf) // append into the hash
	}

	fileHash.MD5 = hex.EncodeToString(hash.Sum(nil))

	return &fileHash, nil
}
