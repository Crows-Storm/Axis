package rule

import "context"

// TODO: TDD Current Status -> Red
type SanitizerRule interface {
	match(ctx context.Context, data []byte) bool
	handle(ctx context.Context, data []byte) []byte
}

type SanitizerRuleEngine struct {
	ctx context.Context
}
