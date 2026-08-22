//go:build !static && !rgfw

package object

import rl "github.com/gen2brain/raylib-go/raylib"

func closeGGWindow() {
	rl.CloseWindow()
}
