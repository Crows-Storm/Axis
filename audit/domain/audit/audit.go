package audit

import (
	"time"

	"github.com/Crows-Storm/Axis/common/domain"
	"github.com/Crows-Storm/Axis/common/domain/audit"
	"github.com/Crows-Storm/Axis/common/domain/event"
)

type Audit struct {
	domain.AggregateRoot

	Id        int64
	Operator  Operator
	Resource  AuditResource
	Context   AuditContext
	Payload   ChangePayload
	Type      event.EventCategory
	Result    audit.OperationResult
	TraceId   string    // trace id / request id
	Timestamp time.Time // Occurred At: 1787740620000 epoch millisecond
}

type Operator struct {
	Id       int64  // user id
	RoleCode string // e.g. ADMIN / MANAGE / USER
}

type AuditResource struct {
}

type AuditContext struct {
}

type ChangePayload struct {
}
