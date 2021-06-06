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
		viewMat = mgl32.Translate3D(0, 0, -eyePos[2]).Mul4(viewMat)
		eyePos[2] += 0.05

		viewMat = mgl32.Translate3D(0, 0, eyePos[2]).Mul4(viewMat)

	case glfw.KeyDown:
		viewMat = mgl32.Translate3D(0, 0, -eyePos[2]).Mul4(viewMat)

		eyePos[2] -= 0.05
		viewMat = mgl32.Translate3D(0, 0, eyePos[2]).Mul4(viewMat)

	case glfw.KeyRight:
		viewMat = mgl32.Translate3D(-eyePos[0], 0, 0).Mul4(viewMat)

		eyePos[0] -= 0.05
		viewMat = mgl32.Translate3D(eyePos[0], 0, 0).Mul4(viewMat)

	case glfw.KeyLeft:
		viewMat = mgl32.Translate3D(-eyePos[0], 0, 0).Mul4(viewMat)

		eyePos[0] += 0.05
		viewMat = mgl32.Translate3D(eyePos[0], 0, 0).Mul4(viewMat)

	case glfw.KeySpace:
		viewMat = mgl32.Translate3D(0, -eyePos[1], 0).Mul4(viewMat)

		eyePos[1] -= 0.05
		viewMat = mgl32.Translate3D(0, eyePos[1], 0).Mul4(viewMat)

	case glfw.KeyZ:
		viewMat = mgl32.Translate3D(0, -eyePos[1], 0).Mul4(viewMat)

		eyePos[1] += 0.05
		viewMat = mgl32.Translate3D(0, eyePos[1], 0).Mul4(viewMat)
	}
	//	fmt.Println(program)
	UpdateUniformMat4fv("view", program, &viewMat[0])
	prevEyePos[0], prevEyePos[1], prevEyePos[2] = eyePos[0], eyePos[1], eyePos[2]
	//fmt.Println(eyePos)
}

func HandleMouseMovement(w *glfw.Window, xpos, ypos float64) {
	switch BtnState {
	case byte('R'):

	default:
		width, height := w.GetFramebufferSize()
		viewMat = defaultViewMat
		viewMat = mgl32.Rotate3DY(float32(2*xpos/float64(width)-1)).Mat4().Mul4(viewMat)

		viewMat = mgl32.Rotate3DX(float32(2*ypos/float64(height)-1)).Mat4().Mul4(viewMat)
		viewMat = mgl32.Translate3D(eyePos[0], eyePos[1], eyePos[2]).Mul4(viewMat)

		CurrPoint.P[0] = float32(2*xpos/float64(width) - 1)
		CurrPoint.P[1] = float32(2*ypos/float64(height) - 1)

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
