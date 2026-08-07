package persistent

import "github.com/alvadorncorp/gopackages/functional/internal/core/transient"

// checkBuilderWindow panics with panicBuilderUsedAfterBuild when win is nil.
// Builders clear win after Freeze so post-Build methods never call into
// transient APIs that would leak "transient: window used after Freeze".
func checkBuilderWindow(win *transient.Window) {
	if win == nil {
		panic(panicBuilderUsedAfterBuild)
	}
}

// freezeBuilderWindow freezes an open window and clears *win so later
// checkBuilderWindow calls treat the builder as already Built.
func freezeBuilderWindow(win **transient.Window) {
	checkBuilderWindow(*win)
	(*win).Freeze()
	*win = nil
}
