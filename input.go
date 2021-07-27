package main

import (
	"fmt"
	"os"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func HandleKeys(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	LookAt = LookAt.Sub(eyePos)
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

		eyePos[0] -= 0.05

	case glfw.KeyLeft:

		eyePos[0] += 0.05

	case glfw.KeySpace:

		eyePos[1] -= 0.05

	case glfw.KeyZ:

		eyePos[1] += 0.05
	}
	//	fmt.Println(program)
	LookAt = eyePos.Add(LookAt)
	UpdateView(
		LookAt,
		eyePos,
	)
}

func HandleMouseMovement(w *glfw.Window, xpos, ypos float64) {
	width, height := w.GetFramebufferSize()
	CurrPoint[0] = float32(2*xpos/float64(width) - 1)
	CurrPoint[1] = -float32(2*ypos/float64(height) - 1)
	switch BtnState {
	case byte('P'):

	case byte('C'):
		LookAt = mgl32.Rotate3DX(CurrPoint[1]).Mul3(mgl32.Rotate3DY(CurrPoint[0])).Mul3x1(mgl32.Vec3{0, 0, -1}).Normalize().Add(eyePos)
		UpdateView(
			LookAt,
			eyePos,
		)
	}

}

func HandleMouseButton(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch button {
	case glfw.MouseButtonLeft:

		switch AddState {
		case byte('l'):
			// TODO: Add CurrPoint to an array of Lines in Bvg
		default:
			if action == glfw.Press {
				for _, btn := range Btns {

					MouseRay = NewRay(RAY_TYPE_CENTERED,
						UnProject(viewMat, projMat),
						mgl32.Vec3{CurrPoint[0], CurrPoint[1], 0},
						mgl32.Vec3{CurrPoint[0], CurrPoint[1], 1},
					)
					iscolliding, collidingAt, collFaces := MouseRay.PolyCollide(btn.Geometry)
					if iscolliding {
						btn.CB(w, MouseX, MouseY, collidingAt, collFaces)
					}
				}
			}
		}
	case glfw.MouseButtonRight:
		fmt.Println(string(BtnState))

		if action == glfw.Press {
			switch BtnState {
			case byte('P'):
				w.SetInputMode(glfw.RawMouseMotion, glfw.True)
				w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
				BtnState = byte('C')
			case byte('C'):

				w.SetInputMode(glfw.RawMouseMotion, glfw.False)
				w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
				BtnState = byte('P')

			}
		}
	}
}
