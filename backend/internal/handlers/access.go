package handlers

import (
	"context"
	"errors"

	"github.com/albertson/albertsonmissionplan/internal/store"
)

// leaderPriv holds effective leader capabilities.
//
// Model: anyone with an approved leader_credentials row is a full admin (can moderate
// experiences, edit ward-wide goals, and edit every organization tab). Access is gated
// entirely by the admin-approval flow — there is no separate email allowlist.
type leaderPriv struct {
	IsAdmin bool
}

// loadLeaderPriv resolves capabilities for the given session email by checking the database
// on each privileged request. This means denying/deleting an account immediately revokes
// access for its existing session on the next request (no stale-session window).
func (s *Server) loadLeaderPriv(ctx context.Context, email string) (leaderPriv, error) {
	if email == "" {
		return leaderPriv{}, errors.New("empty email")
	}
	ok, err := store.IsApprovedLeaderAccount(ctx, s.DB, email)
	if err != nil {
		return leaderPriv{}, err
	}
	if !ok {
		return leaderPriv{}, nil
	}
	return leaderPriv{IsAdmin: true}, nil
}

func (priv leaderPriv) canModerate() bool          { return priv.IsAdmin }
func (priv leaderPriv) canEditWardWideGoals() bool { return priv.IsAdmin }
func (priv leaderPriv) canEditOrg(_ string) bool   { return priv.IsAdmin }
func (priv leaderPriv) canUseWardPlanEditor() bool { return priv.IsAdmin }
