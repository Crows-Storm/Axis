package rbac

import "strings"

// Action bit index. Must stay aligned with auth/domain/permission.Action.
type Action uint

const (
	ActionRead Action = iota
	ActionCreate
	ActionUpdate
	ActionDelete
	ActionEnable
	ActionClose
	ActionDisable
	ActionExport
	ActionAudit
	ActionPublish
)

type Mask uint64

func ActionToMask(a Action) Mask {
	return 1 << a
}

var actionCode = map[Action]string{
	ActionRead:    "READ",
	ActionCreate:  "CREATE",
	ActionUpdate:  "UPDATE",
	ActionDelete:  "DELETE",
	ActionEnable:  "ENABLE",
	ActionClose:   "CLOSE",
	ActionDisable: "DISABLE",
	ActionExport:  "EXPORT",
	ActionAudit:   "AUDIT",
	ActionPublish: "PUBLISH",
}

func ParseAction(code string) (Action, bool) {
	want := strings.ToUpper(strings.TrimSpace(code))
	for a, c := range actionCode {
		if c == want {
			return a, true
		}
	}
	return 0, false
}

func (m Mask) Has(a Action) bool {
	bit := ActionToMask(a)
	return m&bit == bit
}

func (m Mask) Codes() []string {
	var codes []string
	for a := Action(0); a < 64; a++ {
		if !m.Has(a) {
			continue
		}
		if c, ok := actionCode[a]; ok {
			codes = append(codes, c)
		}
	}
	return codes
}
