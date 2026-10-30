package tuie2e

import (
	"strings"
	"testing"

	"github.com/Gaurav-Gosain/tuitest"
)

// The report this file exists for: on the scrolling (niri) layout with wide
// columns, opening a terminal puts the new column after the focused one with
// only a sliver of it on screen. Focus moves into the new pane, so the keys go
// somewhere the user cannot see, and the strip stays where it was because the
// least-scroll rule counts that sliver as visible.
//
// appearance.scroll_column_width is 99 and the cap 100 here, which is the
// width the report came from. At the shipped 55% most of the new column is on
// screen and there is nothing to notice, which is why this needs its own
// config.
//
// The daemon and standalone paths make the pane in different places, so both
// are run. The assertion is that the shell in the new pane answers on screen:
// a pane off the edge cannot render its own output.
//
// NEGATIVE CONTROL: with the reveal removed from either creation path, the
// typed marker never appears, because the new pane is one cell wide on screen.
// The run is in NEGATIVE_CONTROLS.md.
func TestOpeningAPaneOnTheStripBringsTheWholeColumnOnScreen(t *testing.T) {
	for _, rc := range []struct {
		name   string
		daemon bool
	}{
		{name: "daemon", daemon: true},
		{name: "standalone"},
	} {
		t.Run(rc.name, func(t *testing.T) {
			base := t.TempDir()
			writeConfig(t, base, "[appearance]\nscroll_column_width = 99\nscroll_column_max = 100\n")

			var a *tuitest.Terminal
			if rc.daemon {
				killDaemon(t, base)
				if out, err := tuiosCLI(t, base, "new", "stripnew", "--detach"); err != nil {
					t.Fatalf("create session: %v: %s", err, out)
				}
				a = attachIn(t, base, "stripnew", startOpts{cols: 100, rows: 30})
			} else {
				a = startIn(t, base, startOpts{cols: 100, rows: 30})
				waitBoot(t, a)
				newWindow(t, a)
			}
			renameWindow(t, a, "ALPHA")
			enableTiling(t, a)

			if err := a.SendKeys(tuitest.Ctrl('p')); err != nil {
				t.Fatalf("open the palette: %v", err)
			}
			waitPaletteOpen(t, a, "to pick the scrolling layout")
			if err := a.SendKeys("scrolling", tuitest.Enter); err != nil {
				t.Fatalf("pick the scrolling layout: %v", err)
			}
			waitPaletteClosed(t, a, "after picking the scrolling layout")

			// The new pane becomes the second column, after ALPHA, and takes
			// the focus. Ninety-nine percent of the view is ALPHA, so only
			// about one cell of the new column is on screen until the strip is
			// told to show it.
			newWindow(t, a)
			waitWindowCount(t, a, 2, "the new pane")

			enterTerminalMode(t, a)
			runInShell(t, a, `echo NEW""PANE-$((6*7))`, "NEWPANE-42", shellTimeout)

			if band := stripBand(a); !strings.Contains(band, "NEWPANE-42") {
				t.Fatalf("the output of the pane that has the focus is not on screen:\n%s", band)
			}
			alive(t, a, "after opening a pane on the strip")
		})
	}
}
