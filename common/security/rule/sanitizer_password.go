package rule

import (
	"context"
)

type SanitizerPassword struct {
}

func NewSanitizerPassword() *SanitizerPassword {
	return &SanitizerPassword{}
}

func (s *SanitizerPassword) sanitize(ctx context.Context) {
	//TODO implement me
	panic("implement me")
}
