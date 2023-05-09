// Code generated by github.com/Khan/genqlient, DO NOT EDIT.

package github

import (
	"context"
	"time"

	"github.com/Khan/genqlient/graphql"
)

// Autogenerated input type of ChangeUserStatus
type ChangeUserStatusInput struct {
	// A unique identifier for the client performing the mutation.
	ClientMutationId *string `json:"clientMutationId"`
	// The emoji to represent your status. Can either be a native Unicode emoji or an emoji name with colons, e.g., :grinning:.
	Emoji *string `json:"emoji"`
	// If set, the user status will not be shown after this date.
	ExpiresAt *time.Time `json:"expiresAt"`
	// Whether this status should indicate you are not fully available on GitHub, e.g., you are away.
	LimitedAvailability *bool `json:"limitedAvailability"`
	// A short description of your current status.
	Message *string `json:"message"`
	// The ID of the organization whose members will be allowed to see the status. If
	// omitted, the status will be publicly visible.
	OrganizationId *string `json:"organizationId"`
}

// GetClientMutationId returns ChangeUserStatusInput.ClientMutationId, and is useful for accessing the field via an interface.
func (v *ChangeUserStatusInput) GetClientMutationId() *string { return v.ClientMutationId }

// GetEmoji returns ChangeUserStatusInput.Emoji, and is useful for accessing the field via an interface.
func (v *ChangeUserStatusInput) GetEmoji() *string { return v.Emoji }

// GetExpiresAt returns ChangeUserStatusInput.ExpiresAt, and is useful for accessing the field via an interface.
func (v *ChangeUserStatusInput) GetExpiresAt() *time.Time { return v.ExpiresAt }

// GetLimitedAvailability returns ChangeUserStatusInput.LimitedAvailability, and is useful for accessing the field via an interface.
func (v *ChangeUserStatusInput) GetLimitedAvailability() *bool { return v.LimitedAvailability }

// GetMessage returns ChangeUserStatusInput.Message, and is useful for accessing the field via an interface.
func (v *ChangeUserStatusInput) GetMessage() *string { return v.Message }

// GetOrganizationId returns ChangeUserStatusInput.OrganizationId, and is useful for accessing the field via an interface.
func (v *ChangeUserStatusInput) GetOrganizationId() *string { return v.OrganizationId }

// __changeUserStatusInput is used internally by genqlient
type __changeUserStatusInput struct {
	Status ChangeUserStatusInput `json:"status"`
}

// GetStatus returns __changeUserStatusInput.Status, and is useful for accessing the field via an interface.
func (v *__changeUserStatusInput) GetStatus() ChangeUserStatusInput { return v.Status }

// changeUserStatusChangeUserStatusChangeUserStatusPayload includes the requested fields of the GraphQL type ChangeUserStatusPayload.
// The GraphQL type's documentation follows.
//
// Autogenerated return type of ChangeUserStatus
type changeUserStatusChangeUserStatusChangeUserStatusPayload struct {
	// Your updated status.
	Status *changeUserStatusChangeUserStatusChangeUserStatusPayloadStatusUserStatus `json:"status"`
}

// GetStatus returns changeUserStatusChangeUserStatusChangeUserStatusPayload.Status, and is useful for accessing the field via an interface.
func (v *changeUserStatusChangeUserStatusChangeUserStatusPayload) GetStatus() *changeUserStatusChangeUserStatusChangeUserStatusPayloadStatusUserStatus {
	return v.Status
}

// changeUserStatusChangeUserStatusChangeUserStatusPayloadStatusUserStatus includes the requested fields of the GraphQL type UserStatus.
// The GraphQL type's documentation follows.
//
// The user's description of what they're currently doing.
type changeUserStatusChangeUserStatusChangeUserStatusPayloadStatusUserStatus struct {
	Id string `json:"id"`
}

// GetId returns changeUserStatusChangeUserStatusChangeUserStatusPayloadStatusUserStatus.Id, and is useful for accessing the field via an interface.
func (v *changeUserStatusChangeUserStatusChangeUserStatusPayloadStatusUserStatus) GetId() string {
	return v.Id
}

// changeUserStatusResponse is returned by changeUserStatus on success.
type changeUserStatusResponse struct {
	// Update your status on GitHub.
	ChangeUserStatus *changeUserStatusChangeUserStatusChangeUserStatusPayload `json:"changeUserStatus"`
}

// GetChangeUserStatus returns changeUserStatusResponse.ChangeUserStatus, and is useful for accessing the field via an interface.
func (v *changeUserStatusResponse) GetChangeUserStatus() *changeUserStatusChangeUserStatusChangeUserStatusPayload {
	return v.ChangeUserStatus
}

// The query or mutation executed by changeUserStatus.
const changeUserStatus_Operation = `
mutation changeUserStatus ($status: ChangeUserStatusInput!) {
	changeUserStatus(input: $status) {
		status {
			id
		}
	}
}
`

func changeUserStatus(
	ctx context.Context,
	client graphql.Client,
	status ChangeUserStatusInput,
) (*changeUserStatusResponse, error) {
	req := &graphql.Request{
		OpName: "changeUserStatus",
		Query:  changeUserStatus_Operation,
		Variables: &__changeUserStatusInput{
			Status: status,
		},
	}
	var err error

	var data changeUserStatusResponse
	resp := &graphql.Response{Data: &data}

	err = client.MakeRequest(
		ctx,
		req,
		resp,
	)

	return &data, err
}