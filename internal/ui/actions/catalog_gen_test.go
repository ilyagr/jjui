package actions

import (
	"testing"

	keybindings "github.com/idursun/jjui/internal/ui/bindings"
	"github.com/idursun/jjui/internal/ui/intents"
	"github.com/stretchr/testify/require"
)

func TestResolveIntent_ApplyForceFromArgs(t *testing.T) {
	intent, ok := ResolveIntent(OwnerSquash, keybindings.Action("revisions.squash.apply"), map[string]any{"force": true})
	require.True(t, ok)
	apply, ok := intent.(intents.Apply)
	require.True(t, ok)
	require.True(t, apply.Force)
}

func TestResolveIntent_TargetPickerApplyForceFromArgs(t *testing.T) {
	intent, ok := ResolveIntent(OwnerTargetPicker, keybindings.Action("revisions.target_picker.apply"), map[string]any{"force": true})
	require.True(t, ok)
	apply, ok := intent.(intents.TargetPickerApply)
	require.True(t, ok)
	require.True(t, apply.Force)
}

func TestResolveIntent_UnknownOwnerOrAction(t *testing.T) {
	_, ok := ResolveIntent("unknown.owner", keybindings.Action("revisions.apply"), nil)
	require.False(t, ok)

	_, ok = ResolveIntent(OwnerSquash, keybindings.Action("ui.open_git"), nil)
	require.False(t, ok)
}

func TestIsRevisionsOwner(t *testing.T) {
	require.True(t, IsRevisionsOwner(OwnerRevisions))
	require.True(t, IsRevisionsOwner(OwnerRebase))
	require.True(t, IsRevisionsOwner(OwnerDetailsConfirmation))

	require.False(t, IsRevisionsOwner(OwnerUi))
	require.False(t, IsRevisionsOwner(OwnerGit))
	require.False(t, IsRevisionsOwner(OwnerRevset))
	require.False(t, IsRevisionsOwner(OwnerStatusInput))
}
