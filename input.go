package main

import (
	//	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func HandleKeys(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.KeyEscape:
		w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	case glfw.KeyE:
		w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		if glfw.RawMouseMotionSupported() {
			w.SetInputMode(glfw.RawMouseMotion, glfw.True)
		}

	}
}

func HandleMouseMovement(w *glfw.Window, xpos, ypos float64) {
	width, height := w.GetFramebufferSize()
	viewMat = mgl32.LookAt(
		0, 0, 1/viewRange,
		float32(2*xpos/float64(width)-1), float32(2*ypos/float64(height)-1), 1,
		0, 1.0, 0,
	)
	//	fmt.Println(program)
	UpdateUniformMat4fv("view", program, &viewMat[0])
	//	fmt.Println(viewMat)
}
