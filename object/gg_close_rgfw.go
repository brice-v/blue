//go:build !static && rgfw

package object

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// closeGGWindow works around a hang in raylib's CloseWindow when built with
// the rgfw backend: teardown can spin forever after unloading GL resources
// (reproduced with a minimal program on X11, upstream issue pending).
// Run the C teardown detached; if it does not finish quickly, leave it behind
// and let normal process exit reap the thread. On non-rgfw builds this is a
// plain synchronous CloseWindow.
func closeGGWindow() {
	done := make(chan struct{})
	go func() {
		rl.CloseWindow()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
	}
}
