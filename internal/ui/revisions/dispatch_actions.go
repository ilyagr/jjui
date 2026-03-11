package revisions

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/idursun/jjui/internal/ui/actions"
	keybindings "github.com/idursun/jjui/internal/ui/bindings"
	"github.com/idursun/jjui/internal/ui/dispatch"
	"github.com/idursun/jjui/internal/ui/operations"
)

func (m *Model) RouteOwnedIntent(owner string, intent tea.Msg) (tea.Cmd, bool) {
	if !dispatch.IsRevisionsOwner(owner) {
		return nil, false
	}

	if opScope, ok := m.CurrentOperation().(operations.ScopeProvider); ok {
		scope := string(opScope.Scope())
		if scope != "" && scope != string(ScopeRevisions) && ownerBelongsToScope(owner, scope) {
			return m.CurrentOperation().Update(intent), true
		}
	}

	return m.Update(intent), true
}

func (m *Model) HandleDispatchedAction(action keybindings.Action, args map[string]any) (tea.Cmd, bool) {
	if intent, ok := actions.ResolveByAction(action, args); ok {
		owner := dispatch.DeriveOwner(action)
		if cmd, handled := m.RouteOwnedIntent(owner, intent); handled {
			return cmd, true
		}
		return m.Update(intent), true
	}
	return nil, false
}

func ownerBelongsToScope(owner string, scope string) bool {
	if owner == "" || scope == "" {
		return false
	}
	return owner == scope || strings.HasPrefix(owner, scope+".")
}
