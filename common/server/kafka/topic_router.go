package kafka

import (
	"fmt"
	"strings"

	"github.com/Crows-Storm/Axis/common/domain/event"
)

// TopicRouter 将事件名映射到 Kafka Topic
// 规则: {prefix}.{bounded-context}.{aggregate}.{event-name}
// 例:   order.order.aggregate.order-created
type TopicRouter struct {
	prefix    string            // 如 "myapp"
	overrides map[string]string // 事件名 -> topic 的显式覆盖
}

func NewTopicRouter(prefix string) *TopicRouter {
	return &TopicRouter{
		prefix:    prefix,
		overrides: make(map[string]string),
	}
}

// Override 允许为特定事件指定自定义 topic
func (r *TopicRouter) Override(eventName, topic string) {
	r.overrides[eventName] = topic
}

// Resolve 根据事件确定目标 topic
func (r *TopicRouter) Resolve(evt event.DomainEvent) string {
	if topic, ok := r.overrides[evt.EventName()]; ok {
		return topic
	}
	// 默认规则: prefix.event-name (kebab-case)
	return fmt.Sprintf("%s.%s", r.prefix, toKebabCase(evt.EventName()))
}

func toKebabCase(s string) string {
	// 简单的驼峰/下划线转 kebab-case
	var result []rune
	for i, r := range s {
		if r == '_' || r == ' ' {
			result = append(result, '-')
		} else if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-', r+32)
		} else {
			result = append(result, r)
		}
	}
	return strings.ToLower(string(result))
}
