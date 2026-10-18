package adapter

import (
	"fmt"
	"strings"

	commexchange "github.com/Crows-Storm/Axis/common/domain/exchange"
)

type Registry struct {
	clients map[commexchange.Platform]ExchangeClient
}

func NewRegistry(clients ...ExchangeClient) Registry {
	m := make(map[commexchange.Platform]ExchangeClient, len(clients))
	for _, c := range clients {
		if c == nil {
			continue
		}
		m[c.Platform()] = c
	}
	return Registry{clients: m}
}

func (r Registry) Client(platform string) (ExchangeClient, error) {
	p, ok := commexchange.ParsePlatform(platform)
	if !ok {
		return nil, fmt.Errorf("unknown platform %q", strings.TrimSpace(platform))
	}
	c, ok := r.clients[p]
	if !ok {
		return nil, fmt.Errorf("unknown platform %q", p)
	}
	return c, nil
}
