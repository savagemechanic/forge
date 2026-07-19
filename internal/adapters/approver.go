package adapters

import "github.com/cloudspacelab/forge/internal/ports"

// AutoApprover is an approver that always approves (for non-interactive
// or trusted runs). Use with caution.
type AutoApprover struct{}

func (a *AutoApprover) RequestApproval(req ports.ApprovalRequest) ports.ApprovalResponse {
	return ports.ApprovalResponse{Approved: true, Reason: "auto-approved"}
}

// DenyApprover always denies — useful for dry-run / readonly sessions.
type DenyApprover struct{}

func (a *DenyApprover) RequestApproval(req ports.ApprovalRequest) ports.ApprovalResponse {
	return ports.ApprovalResponse{Approved: false, Reason: "read-only session"}
}
