package redis

const (
	policyPrefix   = "policy."
	authorize      = policyPrefix + "authorize"
	authorizeByKey = policyPrefix + "authorizebykey"
	addPolicy      = policyPrefix + "add"
	updatePolicy   = policyPrefix + "update"
	listPolicy     = policyPrefix + "list"
	deletePolicy   = policyPrefix + "delete"
)

type event interface {
	Encode(operation string) map[string]interface{}
}

var (
	_ event = (*policyEvent)(nil)
)

type policyEvent struct {
	entityType string
	clientID   string
	groupID    string
	actions    []string
}

func (cte policyEvent) Encode(operation string) map[string]interface{} {
	return map[string]interface{}{
		"group_id":  cte.groupID,
		"client_id": cte.clientID,
		"actions":   cte.actions,
		"operation": operation,
	}
}
