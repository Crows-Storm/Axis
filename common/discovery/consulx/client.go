package consulx

import (
	"net"
	"net/http"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

type Config struct {
	Address string
	Token   string
	Timeout time.Duration // 用于非 blocking query 的请求（如注册、注销）
}

func NewClient(cfg *Config) (*consulapi.Client, error) {

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Minute,
		ExpectContinueTimeout: 1 * time.Second,
	}

	consulCfg := consulapi.DefaultConfig()
	consulCfg.Address = cfg.Address
	consulCfg.Token = cfg.Token
	consulCfg.HttpClient = &http.Client{
		Transport: transport,
	}

	return consulapi.NewClient(consulCfg)
}
