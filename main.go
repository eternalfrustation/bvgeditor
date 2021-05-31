package main

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"runtime"
	"time"
	"unsafe"
)

func init() {
	runtime.LockOSThread()
}

const (
	W         = 500
	H         = 500
	fps       = 30
	pi        = 3.1415926535897932384626433832795028841971693993751058209749445923078164062862089986280348253421170679
	viewRange = 1000
)

var (
	viewMat   mgl32.Mat4
	projMat   mgl32.Mat4
	AddState  byte
	program   uint32
	MouseX    float64
	MouseY    float64
	CurrPoint *Point
	Btns      []*Button
	BtnState  byte
	eyePos    mgl32.Vec3
	LookAt    mgl32.Vec3
)

func main() {
	// GLFW Initialization
	orDie(glfw.Init())
	CurrPoint = P(0, 0, 1)
	eyePos = mgl32.Vec3{0, 0, 1}
	// Close glfw when main exits
	defer glfw.Terminate()
	// Window Properties
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	// Create the window with the above hints
	window, err := glfw.CreateWindow(W, H, "Bvg Editor", nil, nil)
	orDie(err)
	// Load the icon file
	icoFile, err := os.Open("ico.png")
	orDie(err)
	// decode the file to an image.Image
	ico, err := png.Decode(icoFile)
	orDie(err)
	fmt.Println(ico.ColorModel())
	window.SetIcon([]image.Image{ico})
	window.MakeContextCurrent()
	// OpenGL Initialization
	// Check for the version
	//version := gl.GoStr(gl.GetString(gl.VERSION))
	//	fmt.Println("OpenGL Version", version)
	// Read the vertex and fragment shader files
	vertexShader, err := ioutil.ReadFile("vertex.vert")
	orDie(err)
	vertexShader = append(vertexShader, []byte("\x00")...)
	fragmentShader, err := ioutil.ReadFile("frag.frag")
	orDie(err)
	fragmentShader = append(fragmentShader, []byte("\x00")...)

	orDie(gl.Init())

	// Set the function for handling errors
	gl.DebugMessageCallback(func(source, gltype, id, severity uint32, length int32, message string, userParam unsafe.Pointer) {
		panic(fmt.Sprintf("%d, %d, %d, %d, %d, %s", source, gltype, severity, id, length, message))

	}, nil)
	// Create an OpenGL "Program" and link it for current drawing
	program, err = newProg(string(vertexShader), string(fragmentShader))
	orDie(err)
	// Check for the version
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL Version", version)
	// Main draw loop
	// Get a font
	fnt := NewFont("font/space.ttf", "1234567890,./';[]{}|:\"<>?!@#$%^&*()_+-=qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM~`ā ", IntTo26_6(128))
//	fnt.GlyphMap['ā'].SetTypes(gl.LINES)
	HelloWorld := TextToShape(fnt, "H e l l o Z a W a r u d o")

	HelloWorld.GenVao()
	fmt.Println(len(HelloWorld.Pts))
	// Set the refresh function for the window
	// Use this program
	gl.UseProgram(program)
	// Calculate the projection matrix
	projMat = mgl32.Ident4()
	// set the value of Projection matrix
	UpdateUniformMat4fv("projection", program, &projMat[0])
	// Set the value of view matrix
	viewMat = mgl32.Ident4()
	UpdateUniformMat4fv("view", program, &projMat[0])
	modelMat := mgl32.Ident4()
	UpdateUniformMat4fv("model", program, &modelMat[0])
	window.SetKeyCallback(HandleKeys)
	window.SetCursorPosCallback(HandleMouseMovement)
	window.SetMouseButtonCallback(HandleMouseButton)
	window.SetRefreshCallback(Refresh)
	for !window.ShouldClose() {
		time.Sleep(time.Second / fps)
		// Clear everything that was drawn previously
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// Actually draw something
		HelloWorld.Draw()
//		fnt.GlyphMap['e'].Draw()
		// display everything that was drawn
		window.SwapBuffers()
		// check for any events
		glfw.PollEvents()
	}
}
