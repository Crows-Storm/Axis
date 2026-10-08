package contextx

func bindTrace(b *Builder, get func(string) string, includeAuth bool) {
	reqID := get(MetadataRequestID)
	if reqID == "" {
		reqID = get(MetadataTraceID)
	}
	if reqID != "" {
		b.WithRequestID(reqID)
	}

	if corrID := get(MetadataCorrelationID); corrID != "" {
		b.WithCorrelationID(corrID)
	}

	if parent := get(MetadataParentSpanID); parent != "" {
		b.WithParentSpanID(parent)
	} else if spanID := get(MetadataSpanID); spanID != "" {
		b.WithParentSpanID(spanID)
	}

	if includeAuth {
		if auth := get(MetadataAuthorization); auth != "" {
			b.WithAuthorization(auth)
		}
	}

	if ip := get(MetadataClientIP); ip != "" {
		b.WithClientIP(ip)
	}
	if ua := get(MetadataUserAgent); ua != "" {
		b.WithUserAgent(ua)
	}
}

func injectTrace(r *Reader, set func(key, value string), includeAuth bool) {
	if reqID := r.RequestID(); reqID != "" {
		set(MetadataRequestID, reqID)
		set(MetadataTraceID, reqID)
	}
	if corrID := r.CorrelationID(); corrID != "" {
		set(MetadataCorrelationID, corrID)
	}
	if spanID := r.SpanID(); spanID != "" {
		set(MetadataParentSpanID, spanID)
	}
	if includeAuth {
		if auth := r.Authorization(); auth != "" {
			set(MetadataAuthorization, auth)
		}
	}
	if clientIP := r.ClientIP(); clientIP != "" {
		set(MetadataClientIP, clientIP)
	}
	if userAgent := r.UserAgent(); userAgent != "" {
		set(MetadataUserAgent, userAgent)
	}
}
