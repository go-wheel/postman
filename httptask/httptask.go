package httptask

import (
	"net/http"

	"github.com/lets-go-go/httpclient"
	"github.com/lets-go-go/logger"
)

func DoRequest(requst RequestItem) (string, error) {
	logger.Debugf("request item:%+v", requst)

	method := requst.Request.Method
	reqURL := requst.Request.URL
	headers := requst.Request.Header
	client := httpclient.Access(method, reqURL, 0)
	if len(headers) > 0 {
		for _, header := range headers {
			client.AddHeader(header.Key, header.Value)
		}
	}

	if method == http.MethodGet {
		for _, vals := range requst.Request.Body.Urlencoded {
			if vals.Type == "text" {
				client.AddQuery(vals.Key, vals.Value)
			}
		}
	} else {
		for {
			if requst.Request.Body == nil {
				break
			}
			if requst.Request.Body.Mode == "formdata" {
				if len(requst.Request.Body.Formdata) > 0 {
					break
				}

				body := map[string]string{}
				for _, formdata := range requst.Request.Body.Formdata {
					if formdata.Type == "text" {
						body[formdata.Key] = formdata.Value
					}
				}

				if len(body) > 0 {
					client.SendBody(body)
				}

				break
			} else if requst.Request.Body.Mode == "urlencoded" {
				for _, formdata := range requst.Request.Body.Urlencoded {
					if formdata.Type == "text" {
						client.AddField(formdata.Key, formdata.Value)
					} else if formdata.Type == "file" {
						client = client.AttachFile(formdata.Key, formdata.Value, "")
					} else if formdata.Type == "rangefile" {
						client = client.AttachFileRange(formdata.Key, formdata.Value, "", formdata.Start, formdata.Length)
					}
				}
				break
			} else if requst.Request.Body.Mode == "raw" {
				break
			} else if requst.Request.Body.Mode == "file" {
				break
			}
		}

	}

	return client.Text()
}
