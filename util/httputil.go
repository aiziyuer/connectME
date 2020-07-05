package util

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func NewRequest(client *http.Client) *resty.Request {
	var request *resty.Request

	if client != nil {
		request = resty.
			NewWithClient(client).
			SetLogger(zap.S()).
			SetRetryCount(5).
			SetRetryWaitTime(time.Second * 3).
			SetDebug(viper.GetBool("DEBUG")).R().
			EnableTrace()
	} else {
		request = resty.New().
			SetLogger(zap.S()).
			SetRetryCount(5).
			SetRetryWaitTime(time.Second * 3).
			SetDebug(viper.GetBool("DEBUG")).R().
			EnableTrace()
	}

	return request
}
