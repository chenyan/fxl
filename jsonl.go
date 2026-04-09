package main

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	. "github.com/antonmedv/fx/internal/jsonx"
	"github.com/antonmedv/fx/internal/theme"
)

func jsonlLeftColumnWidth(termW int) int {
	w := termW / 3
	if w < 18 {
		w = 18
	}
	if w > 42 {
		w = 42
	}
	if w >= termW-8 {
		w = max(12, termW-8)
	}
	return w
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

func jsonlItemLabel(n *Node) string {
	const max = 36
	if n == nil {
		return ""
	}
	switch n.Kind {
	case Object:
		if n.HasChildren() && n.Next != nil && n.Next != n.End && n.Next.Key != "" {
			k := n.Next.Key
			if u, err := strconv.Unquote(k); err == nil {
				k = u
			}
			return truncateRunes("{"+k+": …}", max)
		}
		return "{}"
	case Array:
		if n.Size > 0 {
			return truncateRunes(fmt.Sprintf("[…] (%d)", n.Size), max)
		}
		return "[]"
	case String:
		s := n.Value
		if u, err := strconv.Unquote(s); err == nil {
			s = u
		}
		s = strings.ReplaceAll(s, "\n", " ")
		return truncateRunes(s, max)
	case Number, Bool, Null:
		return truncateRunes(n.Value, max)
	default:
		return truncateRunes(n.Value, max)
	}
}

func (m *model) jsonlLeftWidth() int {
	return jsonlLeftColumnWidth(m.termWidth)
}

func (m *model) scrollJSONLListIntoView() {
	nItems := len(m.jsonlRoots)
	if nItems == 0 {
		return
	}
	h := m.viewHeight()
	if m.jsonlIdx < m.jsonlListOff {
		m.jsonlListOff = m.jsonlIdx
	}
	if m.jsonlIdx >= m.jsonlListOff+h {
		m.jsonlListOff = m.jsonlIdx - h + 1
	}
	if m.jsonlListOff < 0 {
		m.jsonlListOff = 0
	}
	if m.jsonlListOff > nItems-1 {
		m.jsonlListOff = max(0, nItems-1)
	}
}

func (m *model) applyJSONLView() {
	if len(m.jsonlRoots) == 0 {
		return
	}
	if m.jsonlIdx >= len(m.jsonlRoots) {
		m.jsonlIdx = len(m.jsonlRoots) - 1
	}
	if m.jsonlIdx < 0 {
		m.jsonlIdx = 0
	}
	sel := m.jsonlRoots[m.jsonlIdx]
	m.top = sel
	m.bottom = sel
	m.head = sel
	m.cursor = 0
	m.scrollJSONLListIntoView()
	rw := m.viewWidth()
	if m.wrap {
		Wrap(sel, rw)
	}
	m.totalLines = sel.Bottom().LineNumber
}

func (m *model) jsonlStatusHint() string {
	if len(m.jsonlRoots) == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d  e expand  E collapse  tab:list  ,/.", m.jsonlIdx+1, len(m.jsonlRoots))
}

func (m *model) jsonlLeftLineBytes(row int) []byte {
	w := m.jsonlLeftWidth()
	idx := m.jsonlListOff + row
	if idx >= len(m.jsonlRoots) {
		return []byte(strings.Repeat(" ", w))
	}
	n := m.jsonlRoots[idx]
	label := jsonlItemLabel(n)
	numW := 4
	if len(m.jsonlRoots) >= 100 {
		numW = 5
	}
	num := fmt.Sprintf("%*d ", numW, idx+1)
	avail := w - len(num)
	if avail < 4 {
		avail = 4
	}
	lab := truncateRunes(label, avail)
	pad := w - len(num) - len(lab)
	if pad < 0 {
		pad = 0
	}
	line := num + lab + strings.Repeat(" ", pad)
	if len(line) > w {
		line = line[:w]
	}
	var styled string
	switch {
	case idx == m.jsonlIdx && m.jsonlListFocus:
		styled = theme.CurrentTheme.Cursor(line)
	case idx == m.jsonlIdx:
		styled = theme.CurrentTheme.Preview(line)
	default:
		styled = line
	}
	return []byte(styled)
}

func (m *model) handleJSONLKey(msg tea.KeyPressMsg) (handled bool, out tea.Model, cmd tea.Cmd) {
	if !m.jsonlMode || len(m.jsonlRoots) == 0 {
		return false, m, nil
	}

	if msg.String() == "tab" {
		m.jsonlListFocus = !m.jsonlListFocus
		return true, m, nil
	}

	if key.Matches(msg, keyMap.JSONLPrev) {
		if m.jsonlIdx > 0 {
			m.jsonlIdx--
			m.applyJSONLView()
			m.recordHistory()
		}
		return true, m, nil
	}
	if key.Matches(msg, keyMap.JSONLNext) {
		if m.jsonlIdx < len(m.jsonlRoots)-1 {
			m.jsonlIdx++
			m.applyJSONLView()
			m.recordHistory()
		}
		return true, m, nil
	}

	if m.jsonlListFocus {
		if key.Matches(msg, keyMap.Expand) {
			m.jsonlListFocus = false
			return true, m, nil
		}
		switch {
		case key.Matches(msg, keyMap.Up):
			if m.jsonlIdx > 0 {
				m.jsonlIdx--
				m.applyJSONLView()
				m.recordHistory()
			}
			return true, m, nil
		case key.Matches(msg, keyMap.Down):
			if m.jsonlIdx < len(m.jsonlRoots)-1 {
				m.jsonlIdx++
				m.applyJSONLView()
				m.recordHistory()
			}
			return true, m, nil
		}
	}
	return false, m, nil
}
