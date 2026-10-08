package event

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	// Prefix is the fixed first segment of every Kafka topic.
	Prefix = "axis"
	// AuditTopic receives a copy of each business event.
	AuditTopic = Prefix + ".audit"
)

// Topic is axis.{domain}.{name}. Each of domain and name is one segment.
func Topic(domain, name string) (string, error) {
	if !singleSegment(domain) || !singleSegment(name) {
		return "", fmt.Errorf("event topic: domain %q and name %q must each be one segment", domain, name)
	}
	return Prefix + "." + domain + "." + name, nil
}

func singleSegment(value string) bool {
	if value == "" || strings.Contains(value, ".") {
		return false
	}
	for _, r := range value {
		if unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
