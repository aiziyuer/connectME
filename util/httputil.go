package util

import (
	"github.com/go-resty/resty/v2"
	"net/http"
)

func NewRequest(client *http.Client) *resty.Request {
	var request *resty.Request

	if client != nil {
		request = resty.
			NewWithClient(client).
			SetRetryCount(3).
			SetDebug(true).R().
			EnableTrace()
	} else {
		request = resty.New().
			SetRetryCount(3).
			SetDebug(true).R().
			EnableTrace()
	}

	return request
}
