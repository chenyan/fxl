package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/antonmedv/fx/internal/ident"
	. "github.com/antonmedv/fx/internal/jsonx"
	"github.com/antonmedv/fx/internal/theme"
	"github.com/antonmedv/fx/internal/utils"
)

func (m *model) View() tea.View {
	if m.suspending {
		return tea.NewView("")
	}

	if m.showHelp {
		return tea.NewView(m.help.View())
	}

	if m.showPreview {
		searchBar := m.previewSearchStatusBar()
		if searchBar != "" {
			return tea.NewView(m.preview.View() + "\n" + searchBar)
		}
		statusBar := flex(m.termWidth, m.cursorPath(), m.fileName)
		return tea.NewView(m.preview.View() + "\n" + theme.CurrentTheme.StatusBar(statusBar))
	}

	var screen []byte
	var cursorLineNumber int

	appendTreeLine := func(lineNumber int, n *Node) []byte {
		var line []byte
		if m.showLineNumbers {
			lineNumbersWidth := len(strconv.Itoa(m.totalLines))
			if n.LineNumber == 0 {
				line = append(line, bytes.Repeat([]byte{' '}, lineNumbersWidth)...)
			} else {
				lineNumStr := fmt.Sprintf("%*d", lineNumbersWidth, n.LineNumber)
				line = append(line, theme.CurrentTheme.LineNumber(lineNumStr)...)
			}
			line = append(line, ' ', ' ')
		}

		for i := 0; i < int(n.Depth); i++ {
			line = append(line, ident.IdentBytes...)
		}

		focusTree := !m.jsonlMode || !m.jsonlListFocus
		isSelected := m.cursor == lineNumber && focusTree
		if isSelected {
			if n.LineNumber == 0 {
				cursorLineNumber = n.Parent.LineNumber
			} else {
				cursorLineNumber = n.LineNumber
			}
		}
		if !m.showCursor {
			isSelected = false // don't highlight the cursor while iterating search results
		}

		isRef := false
		isRefSelected := false

		if n.Key != "" {
			line = append(line, m.prettyKey(n, isSelected)...)
			line = append(line, theme.Colon...)

			_, isRef = isRefNode(n)
			isRefSelected = isRef && isSelected
			isSelected = false // don't highlight the key's value
		}

		line = append(line, m.prettyPrint(n, isSelected, isRef)...)

		if n.IsCollapsed() {
			if n.Kind == Object {
				if n.Collapsed.Key != "" {
					line = append(line, theme.CurrentTheme.Preview(n.Collapsed.Key)...)
					line = append(line, theme.ColonPreview...)
					if len(n.Collapsed.Value) > 0 &&
						len(n.Collapsed.Value) < 42 &&
						n.Collapsed.Kind != Object &&
						n.Collapsed.Kind != Array {
						line = append(line, theme.CurrentTheme.Preview(n.Collapsed.Value)...)
						if n.Size > 1 {
							line = append(line, theme.CommaPreview...)
							line = append(line, theme.Dot3...)
						}
					} else {
						line = append(line, theme.Dot3...)
					}
				}
				line = append(line, theme.CloseCurlyBracket...)
			} else if n.Kind == Array {
				line = append(line, theme.Dot3...)
				line = append(line, theme.CloseSquareBracket...)
			}
			if n.End != nil && n.End.Comma {
				line = append(line, theme.Comma...)
			}
		}
		if n.Comma {
			line = append(line, theme.Comma...)
		}

		if m.showSizes && n.Size > 0 {
			var w string
			if n.Size == 1 {
				if n.Kind == Array {
					w = "item"
				} else if n.Kind == Object {
					w = "key"
				}
			} else {
				if n.Kind == Array {
					w = "items"
				} else if n.Kind == Object {
					w = "keys"
				}
			}
			line = append(line, theme.CurrentTheme.Size(fmt.Sprintf(" (%d %s)", n.Size, w))...)
		}

		if isRefSelected {
			line = append(line, theme.CurrentTheme.Preview("  ctrl+g goto")...)
		}

		return line
	}

	printedLines := 0
	n := m.head

	if m.jsonlMode && len(m.jsonlRoots) > 0 {
		for lineNumber := 0; lineNumber < m.viewHeight(); lineNumber++ {
			left := m.jsonlLeftLineBytes(lineNumber)
			screen = append(screen, left...)
			screen = append(screen, theme.CurrentTheme.StatusBar("│")...)
			if n == nil {
				if m.eof {
					screen = append(screen, theme.Empty...)
				}
				screen = append(screen, '\n')
				printedLines++
				continue
			}
			screen = append(screen, appendTreeLine(lineNumber, n)...)
			screen = append(screen, '\n')
			printedLines++
			n = n.Next
		}
	} else {
		for lineNumber := 0; lineNumber < m.viewHeight(); lineNumber++ {
			if n == nil {
				break
			}
			screen = append(screen, appendTreeLine(lineNumber, n)...)
			screen = append(screen, '\n')
			printedLines++
			n = n.Next
		}

		for i := printedLines; i < m.viewHeight(); i++ {
			if m.eof {
				screen = append(screen, theme.Empty...)
			}
			screen = append(screen, '\n')
		}
	}

	if m.gotoSymbolInput.Focused() && m.fuzzyMatch != nil {
		var matchedStr []byte
		str := m.fuzzyMatch.Str
		for i := 0; i < len(str); i++ {
			if utils.Contains(i, m.fuzzyMatch.Pos) {
				matchedStr = append(matchedStr, theme.CurrentTheme.Search(string(str[i]))...)
			} else {
				matchedStr = append(matchedStr, theme.CurrentTheme.StatusBar(string(str[i]))...)
			}
		}
		repeatCount := m.termWidth - len(str)
		if repeatCount > 0 {
			matchedStr = append(matchedStr, theme.CurrentTheme.StatusBar(strings.Repeat(" ", repeatCount))...)
		}
		screen = append(screen, matchedStr...)
	} else {
		statusBarWidth := m.termWidth
		var indicator string
		if m.eof {
			percent := int(float64(cursorLineNumber) / float64(m.totalLines) * 100)
			if cursorLineNumber == 1 {
				percent = min(1, percent)
			}
			indicator = fmt.Sprintf("%d%%", percent)
		} else {
			indicator = fmt.Sprintf(" %s", m.spinner.View())
			statusBarWidth += 2 // adjust for spinner
		}

		info := fmt.Sprintf("%s %s", indicator, m.fileName)
		if m.jsonlMode && len(m.jsonlRoots) > 0 {
			info = fmt.Sprintf("%s | %s", m.jsonlStatusHint(), info)
		}
		statusBar := flex(statusBarWidth, m.cursorPath(), info)
		screen = append(screen, theme.CurrentTheme.StatusBar(statusBar)...)
	}

	if m.yank {
		screen = append(screen, '\n')
		screen = append(screen, []byte("(y)value  (p)path  (k)key  (b)key+value")...)
	} else if m.showShowSelector {
		screen = append(screen, '\n')
		screen = append(screen, []byte("(s)sizes  (l)line numbers")...)
	} else if m.gotoSymbolInput.Focused() {
		screen = append(screen, '\n')
		screen = append(screen, m.gotoSymbolInput.View()...)
	} else if m.commandInput.Focused() {
		screen = append(screen, '\n')
		screen = append(screen, m.commandInput.View()...)
	} else if m.searchInput.Focused() {
		screen = append(screen, '\n')
		screen = append(screen, m.searchInput.View()...)
	} else if m.searchInput.Value() != "" {
		screen = append(screen, '\n')
		re, ci := regexCase(m.searchInput.Value())
		re = "/" + re + "/"
		if ci {
			re += "i"
		}
		if m.searching {
			status := fmt.Sprintf("%s searching...", m.spinner.View())
			screen = append(screen, flex(m.termWidth, re, status)...)
		} else if m.search.err != nil {
			screen = append(screen, flex(m.termWidth, re, m.search.err.Error())...)
		} else if len(m.search.results) == 0 {
			screen = append(screen, flex(m.termWidth, re, "not found")...)
		} else {
			cursor := fmt.Sprintf("found: [%v/%v]", m.search.cursor+1, len(m.search.results))
			screen = append(screen, flex(m.termWidth, re, cursor)...)
		}
	}

	s := string(screen)
	v := tea.NewView(s)
	v.AltScreen = true
	if _, noMouse := os.LookupEnv("FX_NO_MOUSE"); !noMouse {
		v.MouseMode = tea.MouseModeCellMotion
	}
	return v
}

func (m *model) centerLine(n *Node) {
	middle := m.visibleLines() / 2

	for range middle {
		m.up()
	}

	m.selectNodeInView(n)
}
