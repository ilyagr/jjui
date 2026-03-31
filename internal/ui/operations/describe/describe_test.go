package describe

import (
	"testing"

	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/layout"
	"github.com/idursun/jjui/internal/ui/operations"
	"github.com/idursun/jjui/internal/ui/render"
	"github.com/idursun/jjui/test"
	"github.com/stretchr/testify/assert"
)

func TestRenderToDisplayContext_UsesDynamicHeight(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.GetDescription("change")).SetOutput([]byte("this description should wrap onto multiple lines"))
	defer commandRunner.Verify()

	ctx := test.NewTestContext(commandRunner)
	op := NewOperation(ctx, &jj.Commit{ChangeId: "change", CommitId: "commit"})

	dl := render.NewDisplayContext()
	height := op.RenderToDisplayContext(
		dl,
		&jj.Commit{ChangeId: "change", CommitId: "commit"},
		operations.RenderOverDescription,
		layout.Rect(0, 0, 12, 10),
		layout.Pos(0, 0),
	)

	assert.Greater(t, height, 1)
}
