package revisions

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/idursun/jjui/internal/parser"
	"github.com/idursun/jjui/internal/screen"
	"github.com/idursun/jjui/internal/ui/operations"
	"github.com/idursun/jjui/internal/ui/render"
)

type itemRenderer struct {
	row           parser.Row
	isHighlighted bool
	selectedStyle lipgloss.Style
	textStyle     lipgloss.Style
	dimmedStyle   lipgloss.Style
	matchedStyle  lipgloss.Style
	op            operations.Operation
	SearchText    string
	AceJumpPrefix *string
	isChecked     bool
}

func (ir itemRenderer) getSegmentStyle(segment screen.Segment) lipgloss.Style {
	style := segment.Style
	if ir.isHighlighted {
		style = style.Inherit(ir.selectedStyle)
	} else {
		style = style.Inherit(ir.textStyle)
	}
	return style
}

func (ir itemRenderer) renderSegment(tb *render.TextBuilder, segment *screen.Segment) {
	baseStyle := ir.getSegmentStyle(*segment)
	if ir.SearchText == "" {
		tb.Styled(segment.Text, baseStyle)
		return
	}

	lowerText := strings.ToLower(segment.Text)
	searchText := ir.SearchText
	if !strings.Contains(lowerText, searchText) {
		tb.Styled(segment.Text, baseStyle)
		return
	}

	matchStyle := baseStyle.Inherit(ir.matchedStyle)
	start := 0
	for {
		offset := strings.Index(lowerText[start:], searchText)
		if offset == -1 {
			if start < len(segment.Text) {
				tb.Styled(segment.Text[start:], baseStyle)
			}
			break
		}
		idx := start + offset
		if idx > start {
			tb.Styled(segment.Text[start:idx], baseStyle)
		}
		end := idx + len(searchText)
		if end > len(segment.Text) {
			end = len(segment.Text)
		}
		tb.Styled(segment.Text[idx:end], matchStyle)
		start = end
		if start >= len(segment.Text) {
			break
		}
	}
}
