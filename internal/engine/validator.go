package engine

import (
	"slices"
	"strings"

	"nvim-quest/internal/lessons"
)

func Validate(simulator *Simulator, challenge lessons.Challenge) bool {
	success := challenge.Success
	if success.CursorOnWord != "" && wordAtCursor(simulator.Buffer) != success.CursorOnWord {
		return false
	}
	if success.BufferEquals != nil && !slices.Equal(simulator.Buffer.Lines, success.BufferEquals) {
		return false
	}
	if success.LineDeleted != "" && containsLine(simulator.Buffer.Lines, success.LineDeleted) {
		return false
	}
	if success.WordDeleted != "" && strings.Contains(simulator.Buffer.Text(), success.WordDeleted) {
		return false
	}
	if success.ContainsText != "" && !strings.Contains(simulator.Buffer.Text(), success.ContainsText) {
		return false
	}
	if success.SearchMatchActive != "" && (!simulator.SearchActive || simulator.SearchTerm != success.SearchMatchActive || wordAtCursor(simulator.Buffer) != success.SearchMatchActive) {
		return false
	}
	return true
}

func wordAtCursor(buffer Buffer) string {
	line := buffer.CurrentLine()
	if line == "" {
		return ""
	}
	col := min(buffer.Cursor.Col, len(line)-1)
	start, end := col, col
	for start > 0 && !isBoundary(line[start-1]) {
		start--
	}
	for end < len(line) && !isBoundary(line[end]) {
		end++
	}
	return line[start:end]
}

func isBoundary(character byte) bool {
	return character == ' ' || character == '\t'
}

func containsLine(lines []string, target string) bool {
	return slices.Contains(lines, target)
}
