package main

import (
	"os"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func HandleKeys(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.KeyEscape:
		w.SetShouldClose(true)
		w.Destroy()
		os.Exit(0)
	case glfw.KeyL:
		AddState = byte('l')
		// TODO: Add Key Detection for other shapes
	case glfw.KeyUp:
		eyePos[2] += 0.05
	case glfw.KeyDown:
		eyePos[2] -= 0.05
	case glfw.KeyRight:
		eyePos[0] += 0.05
	case glfw.KeyLeft:
		eyePos[0] -= 0.05
	case glfw.KeySpace:
		eyePos[1] += 0.05
	case glfw.KeyZ:
		eyePos[1] -= 0.05
	}

	viewMat = mgl32.LookAtV(
		eyePos,
		LookAt,
		mgl32.Vec3{0, 1, 0},
	)
	//	fmt.Println(program)
	UpdateUniformMat4fv("view", program, &viewMat[0])
	//fmt.Println(eyePos)
}

func HandleMouseMovement(w *glfw.Window, xpos, ypos float64) {
	switch BtnState {
	case byte('R'):

	default:
		width, height := w.GetFramebufferSize()
		CurrPoint.P[0] = -float32(2*xpos/float64(width) - 1)
		CurrPoint.P[1] = float32(2*ypos/float64(height) - 1)
		LookAt = eyePos.Sub(CurrPoint.P)
		viewMat = mgl32.LookAtV(
			eyePos,
			LookAt,
			mgl32.Vec3{0, 1, 0},
		)
		//	fmt.Println(program)
		UpdateUniformMat4fv("view", program, &viewMat[0])
		//	fmt.Println(viewMat)
	}
}

func HandleMouseButton(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch button {
	case glfw.MouseButtonLeft:
		switch AddState {
		case byte('l'):
			// TODO: Add CurrPoint to an array of Lines in Bvg
		case 0:
			for _, btn := range Btns {
				if PtPolyCollision(P(float32(MouseX), float32(MouseY), 1), btn.Geometry) {
					btn.CB(w, btn, MouseX, MouseY)
				}
			}
		}
	case glfw.MouseButtonRight:
		BtnState = byte('R')
	}
}
