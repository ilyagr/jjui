package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/test"
)

func Test_Update_RevsetWithEmptyInputKeepsDefaultRevset(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	defer commandRunner.Verify()

	ctx := test.NewTestContext(commandRunner)
	ctx.DefaultRevset = "assume-passed-from-cli"

	model := NewUI(ctx)
	model.Update(common.UpdateRevSetMsg(""))

	assert.Equal(t, ctx.DefaultRevset, ctx.CurrentRevset)
}
