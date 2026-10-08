package event

import "github.com/Crows-Storm/Axis/common/contextx"

// WithCause stamps a reaction event: same correlation as the consumed wire,
// causation is the parent event_id. Command-originated envelopes must not call this.
func WithCause(env Envelope[Message], cause Wire) Envelope[Message] {
	md := env.Metadata
	if corr := lookupWire(cause, contextx.MetadataCorrelationID); corr != "" {
		md = md.With(contextx.MetadataCorrelationID, corr)
	}
	if cause.EventID != "" {
		md = md.With(contextx.MetadataCausationId, cause.EventID)
	}
	env.Metadata = md
	return env
}

func lookupWire(wire Wire, key string) string {
	if wire.Metadata != nil {
		if v := wire.Metadata[key]; v != "" {
			return v
		}
	}
	if wire.Headers != nil {
		return wire.Headers[key]
	}
	return ""
}
