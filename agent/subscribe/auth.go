package subscribe

import (
	fmt "fmt"

	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/agent/consul/state"
	"github.com/hashicorp/consul/agent/consul/stream"
	"github.com/hashicorp/consul/agent/structs"
)

// EnforceACL takes an acl.Authorizer and returns the decision for whether the
// event is allowed to be sent to this client or not.
func enforceACL(authz acl.Authorizer, e stream.Event) acl.EnforcementDecision {
	switch {
	case e.IsEndOfSnapshot(), e.IsEndOfEmptySnapshot():
		return acl.Allow
	}

	p, ok := e.Payload.(state.EventPayload)
	if !ok {
		panic(fmt.Sprintf("unexpected event payload: %T", e.Payload))
	}

	switch v := p.Obj.(type) {
	case *structs.CheckServiceNode:

		/*
			TODO:

			// If it's not populated it's likely a bug so don't send it (or panic on
			// nils). This might catch us out if we ever send partial messages but
			// hopefully test will show that up early.
			if v.ServiceHealth == nil || v.ServiceHealth.CheckServiceNode == nil {
				return acl.Deny
			}
			csn := v.ServiceHealth.CheckServiceNode
		*/

		csn := v
		if csn.Node == nil || csn.Service == nil || csn.Node.Node == "" || csn.Service.Service == "" {
			return acl.Deny
		}

		// TODO(streaming): we need the AuthorizerContext for ent
		if dec := authz.NodeRead(csn.Node.Node, nil); dec != acl.Allow {
			return acl.Deny
		}

		// TODO(streaming): we need the AuthorizerContext for ent
		// Enterprise support for streaming events - they don't have enough data to
		// populate it yet.
		if dec := authz.ServiceRead(csn.Service.Service, nil); dec != acl.Allow {
			return acl.Deny
		}
		return acl.Allow
	}

	return acl.Deny
}
