package client

import (
	"net"
	"net/http"
	"time"
)

var (
	Transport  *http.Transport
	HttpClient *http.Client
)

// init all signal Transport and HttpClient object
func init() {
	Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	HttpClient = &http.Client{
		Transport: Transport,
		Timeout:   5 * time.Second,
	}
}
