package render

// ClampStartLine constrains a scroll start line to valid bounds.
func ClampStartLine(startLine int, listHeight int, itemCount int, itemHeight int) int {
	if startLine < 0 {
		startLine = 0
	}
	totalLines := itemCount * itemHeight
	maxStart := totalLines - listHeight
	if maxStart < 0 {
		maxStart = 0
	}
	if startLine > maxStart {
		startLine = maxStart
	}
	return startLine
}
