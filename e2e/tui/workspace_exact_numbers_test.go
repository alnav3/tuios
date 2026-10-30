package tuie2e

import (
	"testing"

	"github.com/Gaurav-Gosain/tuitest"
)

// workspaces.exact_numbers: a switch lands on the number the person pressed,
// even when the numbers below it hold no panes. Without it, a switch past the
// last workspace holding a pane opens the next one instead (see
// TestWorkspaceSwitchPastTheLastOpensTheNext).
//
// The session under test starts with one pane on workspace 1 and no pane on
// any other workspace, which is the report's case: opt+3 with workspace 2
// empty.

const exactNumbersSession = "exact"

// exactNumbersSetup starts a session with one pane on workspace 1, writes the
// config body, and attaches a client to it.
func exactNumbersSetup(t *testing.T, body string) (*tuitest.Terminal, string) {
	t.Helper()
	base := t.TempDir()
	writeConfig(t, base, body)
	killDaemon(t, base)
	if out, err := tuiosCLI(t, base, "new", exactNumbersSession, "--detach"); err != nil {
		t.Fatalf("create session: %v: %s", err, out)
	}
	term := attachIn(t, base, exactNumbersSession, startOpts{})
	return term, base
}

// waitExactDockStatus waits for the dock's "<workspace>:<count>" field to read
// want.
func waitExactDockStatus(t *testing.T, term *tuitest.Terminal, want, what string) {
	t.Helper()
	if err := term.WaitFor(func(s tuitest.Screen) bool {
		m := renumberDockStatus(s)
		return m != nil && m[0] == want
	}, uiTimeout); err != nil {
		t.Fatalf("%s: the dock never read %s: %v\n%s", what, want, err, term.Snapshot())
	}
}

// TestWorkspaceExactNumbersSwitch turns the option on in the file: opt+3 with
// only workspace 1 holding a pane shows the empty 3, a pane opened there stays
// on 3, and the number reaches it again after a switch away.
func TestWorkspaceExactNumbersSwitch(t *testing.T) {
	term, base := exactNumbersSetup(t, "[workspaces]\nexact_numbers = true\n")

	if err := term.SendKeys(tuitest.Alt('3')); err != nil {
		t.Fatalf("press alt+3: %v", err)
	}
	waitExactDockStatus(t, term, "3:0", "opt+3 with only workspace 1 populated")
	if ws := currentWorkspace(t, base, exactNumbersSession); ws != 3 {
		t.Fatalf("the daemon shows workspace %d after opt+3, want 3", ws)
	}
	if n := panesOn(t, base, exactNumbersSession, 2); n != 0 {
		t.Fatalf("workspace 2 holds %d panes after opt+3, want none", n)
	}

	// A pane opened on the empty 3 stays there, and 2 stays empty.
	newWindow(t, term)
	waitPanesOn(t, term, base, exactNumbersSession, 3, 1, "after n on the empty 3")
	if n := panesOn(t, base, exactNumbersSession, 2); n != 0 {
		t.Fatalf("workspace 2 holds %d panes after a pane on 3, want none", n)
	}

	// Leaving and coming back by the number reaches the pane on 3.
	if err := term.SendKeys(tuitest.Alt('1')); err != nil {
		t.Fatalf("press alt+1: %v", err)
	}
	waitExactDockStatus(t, term, "1:1", "opt+1 back to the first workspace")
	if err := term.SendKeys(tuitest.Alt('3')); err != nil {
		t.Fatalf("press alt+3 again: %v", err)
	}
	waitExactDockStatus(t, term, "3:1", "opt+3 back to the pane on 3")
	if ws := currentWorkspace(t, base, exactNumbersSession); ws != 3 {
		t.Fatalf("the daemon shows workspace %d after the second opt+3, want 3", ws)
	}
	alive(t, term, "after switching by exact number")
}

// TestWorkspaceExactNumbersSetLive turns the option on while tuios runs with
// set-config: opt+3 lands on 2 before the change, as it always did, and on 3
// after it.
func TestWorkspaceExactNumbersSetLive(t *testing.T) {
	term, base := exactNumbersSetup(t, "")

	if err := term.SendKeys(tuitest.Alt('3')); err != nil {
		t.Fatalf("press alt+3: %v", err)
	}
	waitExactDockStatus(t, term, "2:0", "opt+3 with the option off")
	if ws := currentWorkspace(t, base, exactNumbersSession); ws != 2 {
		t.Fatalf("the daemon shows workspace %d after opt+3 with the option off, want 2", ws)
	}

	if err := term.SendKeys(tuitest.Alt('1')); err != nil {
		t.Fatalf("press alt+1: %v", err)
	}
	waitExactDockStatus(t, term, "1:1", "opt+1 back to the first workspace")

	if out, err := tuiosCLI(t, base, "set-config", "workspaces.exact_numbers", "true", "-s", exactNumbersSession); err != nil {
		t.Fatalf("set-config workspaces.exact_numbers: %v: %s", err, out)
	}
	if err := term.SendKeys(tuitest.Alt('3')); err != nil {
		t.Fatalf("press alt+3 after set-config: %v", err)
	}
	waitExactDockStatus(t, term, "3:0", "opt+3 with the option set live")
	if ws := currentWorkspace(t, base, exactNumbersSession); ws != 3 {
		t.Fatalf("the daemon shows workspace %d after opt+3 with the option on, want 3", ws)
	}
	alive(t, term, "after switching by exact number set live")
}
