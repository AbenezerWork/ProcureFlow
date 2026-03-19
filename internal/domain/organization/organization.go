package organization

import (
	"time"

	"github.com/google/uuid"
)

type MembershipRole string

const (
	MembershipRoleOwner              MembershipRole = "owner"
	MembershipRoleAdmin              MembershipRole = "admin"
	MembershipRoleRequester          MembershipRole = "requester"
	MembershipRoleApprover           MembershipRole = "approver"
	MembershipRoleProcurementOfficer MembershipRole = "procurement_officer"
	MembershipRoleViewer             MembershipRole = "viewer"
)

type MembershipStatus string

const (
	MembershipStatusInvited   MembershipStatus = "invited"
	MembershipStatusActive    MembershipStatus = "active"
	MembershipStatusSuspended MembershipStatus = "suspended"
	MembershipStatusRemoved   MembershipStatus = "removed"
)

type Organization struct {
	ID              uuid.UUID
	Name            string
	Slug            string
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ArchivedAt      *time.Time
}

type Membership struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	UserID          uuid.UUID
	Role            MembershipRole
	Status          MembershipStatus
	CreatedByUserID uuid.UUID
	InvitedAt       time.Time
	ActivatedAt     *time.Time
	SuspendedAt     *time.Time
	RemovedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UserOrganization struct {
	Organization Organization
	Role         MembershipRole
	Status       MembershipStatus
}
