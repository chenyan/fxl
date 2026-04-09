package main

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/antonmedv/fx/internal/jsonx"
	"github.com/antonmedv/fx/internal/teatest"
)

func init() {
	w := colorprofile.NewWriter(io.Discard, nil)
	w.Profile = colorprofile.ANSI
	lipgloss.Writer = w
}

type options struct {
	showSizes       bool
	showLineNumbers bool
}

func prepare(t *testing.T, opts ...options) *teatest.TestModel {
	file, err := os.Open("testdata/example.json")
	require.NoError(t, err)

	json, err := io.ReadAll(file)
	require.NoError(t, err)

	head, err := jsonx.Parse(json)
	require.NoError(t, err)

	m := &model{
		top:          head,
		head:         head,
		bottom:       head,
		totalLines:   head.Bottom().LineNumber,
		eof:          true,
		wrap:         true,
		showCursor:   true,
		searchInput:  textinput.New(),
		search:       newSearch(),
		commandInput: textinput.New(),
	}

	if len(opts) > 0 {
		m.showSizes = opts[0].showSizes
		m.showLineNumbers = opts[0].showLineNumbers
	}

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)
	return tm
}

func read(t *testing.T, tm *teatest.TestModel) []byte {
	var out []byte
	teatest.WaitFor(t,
		tm.Output(),
		func(b []byte) bool {
			out = b
			return bytes.Contains(b, []byte("{"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second),
	)
	return out
}

func keyRune(r rune) tea.Msg {
	return tea.KeyPressMsg(tea.Key{Text: string(r), Code: r})
}

func keyStr(s string) []tea.Msg {
	ms := make([]tea.Msg, 0, len(s))
	for _, r := range s {
		ms = append(ms, keyRune(r))
	}
	return ms
}

func TestOutput(t *testing.T) {
	tm := prepare(t)

	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(keyRune('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestNavigation(t *testing.T) {
	tm := prepare(t)

	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(keyRune('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestCollapseRecursive(t *testing.T) {
	tm := prepare(t)

	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft, Mod: tea.ModShift}))
	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(keyRune('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestCollapseRecursiveWithSizes(t *testing.T) {
	tm := prepare(t, options{showSizes: true})

	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft, Mod: tea.ModShift}))
	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(keyRune('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
