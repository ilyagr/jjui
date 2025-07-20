package test

import (
	"os"
	"strings"
	"testing"

	"github.com/idursun/jjui/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestParser_Parse(t *testing.T) {
	file, _ := os.Open("testdata/output.log")
	rows := parser.ParseRows(file)
	assert.Len(t, rows, 11)
}

func TestParser_Parse_NoCommitId(t *testing.T) {
	file, _ := os.Open("testdata/no-commit-id.log")
	rows := parser.ParseRows(file)
	assert.Len(t, rows, 1)
}

func TestParser_Parse_ShortId(t *testing.T) {
	file, _ := os.Open("testdata/short-id.log")
	rows := parser.ParseRows(file)
	assert.Len(t, rows, 2)
	assert.Equal(t, "X", rows[0].Commit.ChangeId)
	assert.Equal(t, "E", rows[0].Commit.CommitId)
	assert.Equal(t, "T", rows[1].Commit.ChangeId)
	assert.Equal(t, "79", rows[1].Commit.CommitId)
}

func TestParser_Parse_SingleLineWithDescription(t *testing.T) {
	file, _ := os.Open("testdata/single-line-with-description.log")
	rows := parser.ParseRows(file)
	assert.Len(t, rows, 1)
	assert.Equal(t, "x", rows[0].Commit.ChangeId)
	assert.Equal(t, "4", rows[0].Commit.CommitId)
}

func TestParser_Parse_CommitIdOnASeparateLine(t *testing.T) {
	file, _ := os.Open("testdata/commit-id.log")
	rows := parser.ParseRows(file)
	assert.Len(t, rows, 1)
	assert.Equal(t, "o", rows[0].Commit.ChangeId)
	assert.Equal(t, "5", rows[0].Commit.CommitId)
}

func TestParser_Parse_ConflictedLongIds(t *testing.T) {
	file, _ := os.Open("testdata/conflicted-change-id.log")
	rows := parser.ParseRows(file)
	assert.Len(t, rows, 3)
	assert.Equal(t, "p??", rows[0].Commit.ChangeId)
	assert.Equal(t, "qusvoztl??", rows[1].Commit.ChangeId)
	assert.Equal(t, "tyoqvzlm??", rows[2].Commit.ChangeId)
}

func TestParser_Parse_Disconnected(t *testing.T) {
	var lb LogBuilder
	lb.Write("*   id=abcde author=some@author id=xyrq")
	lb.Write("│   some documentation")
	lb.Write("~\n")
	lb.Write("*   id=abcde author=some@author id=xyrq")
	lb.Write("│   another commit")
	lb.Write("~\n")
	rows := parser.ParseRows(strings.NewReader(lb.String()))
	assert.Len(t, rows, 2)
}

func TestParser_Parse_Extend(t *testing.T) {
	var lb LogBuilder
	lb.Write("*   id=abcde author=some@author id=xyrq")
	lb.Write("│   some documentation")

	rows := parser.ParseRows(strings.NewReader(lb.String()))
	assert.Len(t, rows, 1)
	row := rows[0]

	extended := row.Lines[1].Extend(row.Indent)
	assert.Len(t, extended.Segments, 1)
}

func TestParser_Parse_WorkingCopy(t *testing.T) {
	var lb LogBuilder
	lb.Write("*   id=abcde author=some@author id=xyrq")
	lb.Write("│   some documentation")
	lb.Write("@   id=kdys author=some@author id=12cd")
	lb.Write("│   some documentation")

	rows := parser.ParseRows(strings.NewReader(lb.String()))
	assert.Len(t, rows, 2)
	row := rows[1]

	assert.True(t, row.Commit.IsWorkingCopy)
}

func TestParser_Parse_WorkingCopy2(t *testing.T) {
	input := `[1m[38;5;14m◆[0m  [1m[38;5;5mtrvy[0m[38;5;8mrr[39m [38;5;3mbryce.z.berg[39m [38;5;6m13 hours ago[39m [38;5;5mmain[39m [1m[38;5;4m38a2[0m[38;5;8m3f9[39m
│  diff stat: ...
[38;5;8m~[39m  [38;5;8m(elided revisions)[39m
│ ○  [1m[38;5;5mpqqp[0m[38;5;8mut[39m [38;5;3milyagr[39m@[38;5;3musers[39m [38;5;6m21 hours ago[39m [38;5;5msplittmp[39m [1m[38;5;4mfc9[0m[38;5;8m969c[39m
│ │  splittmp
│ │ ○  [1m[38;5;5mpotzm[0m[38;5;8mq[39m [38;5;3milyagr[39m@[38;5;3musers[39m [38;5;6m21 hours ago[39m [1m[38;5;4mf1d5[0m[38;5;8m512[39m
│ ╭─┤  run merge
│ │ ○  [1m[38;5;5msoo[0m[38;5;8mouq[39m [38;5;3milyagr[39m@[38;5;3musers[39m [38;5;6m3 weeks ago[39m [1m[38;5;4mbfc[0m[38;5;8mac43[39m
│ │ │  fix
│ │ ○  [1m[38;5;5mmqxwq[0m[38;5;8mt[39m [38;5;3mphilipmetzge[39m [38;5;6m1 month ago[39m [38;5;5mpr/4021[39m [1m[38;5;4m4fd2[0m[38;5;8m8a6[39m
│ │ │  run: Flesh out a bare implementation of 
│ │ │ [1m[38;5;2m@[0m  [1m[38;5;13mmxx[38;5;8mutw[39m [38;5;3milyagr[39m@[38;5;3musers[39m [38;5;14m21 hours ago[39m [38;5;12mcf7d17[38;5;8m0[39m[0m
│ │ │ │  I AM THE WORKING COPY SEE THE AT ON LINE ABOVE
│ │ │ │ ◌  [1m[38;5;5mzpw[0m[38;5;8mxnp[39m [38;5;3milyagr[39m@[38;5;3musers[39m [38;5;6m21 hours ago[39m [38;5;5milyaplus[39m [1m[38;5;4m7c33[0m[38;5;8mb3f[39m
│ ╭─────┤  [38;5;2m(empty)[39m newsplit merge
│ ◌ │ │ │    [1m[38;5;5mstnq[0m[38;5;8myy[39m [38;5;3milyagr[39m@[38;5;3musers[39m [38;5;6m21 hours ago[39m [38;5;5milya[39m [1m[38;5;4m79b[0m[38;5;8m5189[39m
╭─┼───┬───╮  [38;5;2m(empty)[39m Merge
│ │ │ │ │ ○  [1m[38;5;5mvuu[0m[38;5;8mxxm[39m [38;5;3milyagr[39m@[38;5;3musers[39m [38;5;6m21 hours ago[39m [38;5;5mevodiff[39m [1m[38;5;4m89b[0m[38;5;8m6808[39m
├─────────╯  cli 
`

	rows := parser.ParseRows(strings.NewReader(input))
	assert.Len(t, rows, 9)
	row := rows[5]

	// BUG. Line 5 *is* the working copy
	assert.True(t, !row.Commit.IsWorkingCopy)
	workingCopyIndex := -1
	for i := range rows {
		if rows[i].Commit.IsWorkingCopy {
			t.Logf("working copy at index %d: %+v", i, rows[i].Commit)
			workingCopyIndex = i
		}
	}
	// BUG: Should be 5, certainly not -1
	assert.Equal(t, workingCopyIndex, -1)
}

func TestParser_ChangeIdLikeDescription(t *testing.T) {
	file, _ := os.Open("testdata/change-id-like-description.log")
	rows := parser.ParseRows(file)
	assert.Len(t, rows, 1)
}
