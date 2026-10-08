package permission

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

const (
	PermRead    Mask = 1 << ActionRead
	PermCreate  Mask = 1 << ActionCreate
	PermUpdate  Mask = 1 << ActionUpdate
	PermDelete  Mask = 1 << ActionDelete
	PermEnable  Mask = 1 << ActionEnable
	PermClose   Mask = 1 << ActionClose
	PermDisable Mask = 1 << ActionDisable
	PermExport  Mask = 1 << ActionExport
	PermAudit   Mask = 1 << ActionAudit
	PermPublish Mask = 1 << ActionPublish

	PermReadOnly Mask = PermRead
	PermCRUD     Mask = PermRead | PermCreate | PermUpdate | PermDelete
	PermAll      Mask = PermRead | PermCreate | PermUpdate | PermDelete |
		PermEnable | PermClose | PermDisable |
		PermExport | PermAudit | PermPublish
)

// ActionToMask maps an action bit index to its uint64 mask value (1 << bit).
func ActionToMask(a Action) Mask {
	return 1 << a
}

var ActionName = map[Action]string{
	ActionRead:    "View",
	ActionCreate:  "Add",
	ActionUpdate:  "Edit",
	ActionDelete:  "Delete",
	ActionEnable:  "Activate",
	ActionClose:   "Close",
	ActionDisable: "Disable",
	ActionExport:  "Export",
	ActionAudit:   "Audit",
	ActionPublish: "Publish",
}

var ActionCode = map[Action]string{
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

// ParseAction resolves an action code such as "READ" or "DELETE".
func ParseAction(code string) (Action, bool) {
	for a, c := range ActionCode {
		if c == code {
			return a, true
		}
	}
	return 0, false
}

// Has reports whether every bit of the given action is set: (m & bit) == bit.
func (m Mask) Has(a Action) bool {
	bit := ActionToMask(a)
	return m&bit == bit
}

// HasMask reports whether all required bits are present: (m & required) == required.
func (m Mask) HasMask(required Mask) bool {
	return m&required == required
}

// HasAny reports whether any required bit is present: (m & required) != 0.
func (m Mask) HasAny(required Mask) bool {
	return m&required != 0
}

// Add returns a new mask with the action bit set (OR).
func (m Mask) Add(a Action) Mask {
	return m | ActionToMask(a)
}

// Remove returns a new mask with the action bit cleared (AND NOT).
func (m Mask) Remove(a Action) Mask {
	return m &^ ActionToMask(a)
}

// Toggle returns a new mask with the action bit flipped (XOR).
func (m Mask) Toggle(a Action) Mask {
	return m ^ ActionToMask(a)
}

// List returns every action whose bit is set on this mask.
func (m Mask) List() []Action {
	var result []Action
	for a := Action(0); a < 64; a++ {
		if m.Has(a) {
			result = append(result, a)
		}
	}
	return result
}

// Names returns display names for bits that are set.
func (m Mask) Names() []string {
	var names []string
	for _, a := range m.List() {
		if n, ok := ActionName[a]; ok {
			names = append(names, n)
		}
	}
	return names
}

// Codes returns action codes (READ, CREATE, ...) for bits that are set.
func (m Mask) Codes() []string {
	var codes []string
	for _, a := range m.List() {
		if c, ok := ActionCode[a]; ok {
			codes = append(codes, c)
		}
	}
	return codes
}
