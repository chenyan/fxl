package main

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/antonmedv/fx/internal/teatest"
)

func TestGotoLine(t *testing.T) {
	tm := prepare(t, options{showLineNumbers: true})

	tm.Send(keyRune(':'))
	for _, msg := range keyStr("5") {
		tm.Send(msg)
	}
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(keyRune('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestGotoLineCollapsed(t *testing.T) {
	tm := prepare(t, options{showLineNumbers: true})

	tm.Send(keyRune('E'))

	tm.Send(keyRune(':'))
	for _, msg := range keyStr("5") {
		tm.Send(msg)
	}
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(keyRune('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestGotoLineInputInvalid(t *testing.T) {
	tm := prepare(t, options{showLineNumbers: true})

	tm.Send(keyRune('E'))

	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	tm.Send(keyRune(':'))
	for _, msg := range keyStr("invalid") {
		tm.Send(msg)
	}
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(keyRune('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestGotoLineInputGreaterThanTotalLines(t *testing.T) {
	tm := prepare(t, options{showLineNumbers: true})

	tm.Send(keyRune(':'))
	for _, msg := range keyStr("500") {
		tm.Send(msg)
	}
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(keyRune('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestGotoLineInputLessThanOne(t *testing.T) {
	tm := prepare(t, options{showLineNumbers: true})

	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	tm.Send(keyRune(':'))
	tm.Send(keyRune('-'))
	tm.Send(keyRune('2'))
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(keyRune('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestGotoLineKeepsHistory(t *testing.T) {
	tm := prepare(t, options{showLineNumbers: true})

	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))

	tm.Send(keyRune(':'))
	tm.Send(keyRune('4'))
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	tm.Send(keyRune(':'))
	for _, msg := range keyStr("14") {
		tm.Send(msg)
	}
	tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	tm.Send(keyRune('['))

	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(keyRune('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
