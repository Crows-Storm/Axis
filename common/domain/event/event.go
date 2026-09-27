package event

import (
	"maps"

	"github.com/Crows-Storm/Axis/common/util"
)

type Message interface {
	Name() string
}

type Envelope[T Message] struct {
	Message  T
	Metadata Metadata
}

// Metadata contains some data related to a Message that are not functional
// for the Message itself, but instead functioning as supporting information
// to provide additional context.
type Metadata map[string]string

// With returns a new Metadata reference holding the value addressed using
// the specified key.
func (m Metadata) With(key, value string) Metadata {
	if m == nil {
		m = make(Metadata)
	}
	m[key] = value
	return m
}

// Merge merges the other Metadata provided in input with the current map.
// Returns a pointer to the extended metadata map.
func (m Metadata) Merge(other Metadata) Metadata {
	if m == nil {
		return other
	}
	maps.Copy(m, other)
	return m
}

type GenericEnvelope Envelope[Message]

type EventCategory string

const (
	Created EventCategory = "CREATED"
	Updated EventCategory = "UPDATED"
	Deleted EventCategory = "DELETED"

	ApplyPassword EventCategory = "APPLY_PASSWORD"

	Activated EventCategory = "ACTIVATED"
	Disabled  EventCategory = "DISABLED"

	Custom EventCategory = "CUSTOM"
	Mark   EventCategory = "MARK"
)

// ToGenericEnvelope maps the Envelope instance into a GenericEnvelope one.
func (e Envelope[T]) ToGenericEnvelope() GenericEnvelope {
	return GenericEnvelope{
		Message:  e.Message,
		Metadata: e.Metadata,
	}
}

// The GenerateEventID to generate a EventID
func GenerateEventID() string {
	return util.UUIDV4().String()
}
