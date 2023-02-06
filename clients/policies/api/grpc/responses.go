package grpc

type authorizeRes struct {
	authorized bool
}

type identityRes struct {
	id    string
	email string
}

type issueRes struct {
	value string
}

type addPolicyRes struct {
	authorized bool
}

type deletePolicyRes struct {
	deleted bool
}

type listPoliciesRes struct {
	policies []string
}
