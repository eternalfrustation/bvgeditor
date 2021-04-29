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
	"math"
)

func init() {
	runtime.LockOSThread()
}

const (
	W   = 500
	H   = 500
	fps = 30
	pi = 3.1415926535897932384626433832795028841971693993751058209749445923078164062862089986280348253421170679
	viewRange = 1000
)

var (
	viewMat mgl32.Mat4
	projMat mgl32.Mat4
	program uint32
)

func main() {
	// GLFW Initialization
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	// Close glfw when main exits
	defer glfw.Terminate()
	// Window Properties
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	// Create the window with the above hints
	window, err := glfw.CreateWindow(W, H, "Game", nil, nil)
	if err != nil {
		panic(err)
	}
	window.SetKeyCallback(HandleKeys)
	window.SetCursorPosCallback(HandleMouseMovement)
	window.SetRefreshCallback(Refresh)
	// Load the icon file
	icoFile, err := os.Open("ico.png")
	if err != nil {
		panic(err)
	}
	// decode the file to an image.Image
	ico, err := png.Decode(icoFile)
	if err != nil {
		panic(err)
	}
	fmt.Println(ico.ColorModel())
	window.SetIcon([]image.Image{ico})
	window.MakeContextCurrent()
	// OpenGL Initialization
	// Check for the version
	//version := gl.GoStr(gl.GetString(gl.VERSION))
	//	fmt.Println("OpenGL Version", version)
	// Read the vertex and fragment shader files
	vertexShader, err := ioutil.ReadFile("vertex.vert")
	if err != nil {
		panic(err)
	}
	vertexShader = append(vertexShader, []byte("\x00")...)
	fragmentShader, err := ioutil.ReadFile("frag.frag")
	if err != nil {
		panic(err)
	}
	fragmentShader = append(fragmentShader, []byte("\x00")...)

	err = gl.Init()
	if err != nil {
		panic(err)
	}
	// Set the function for handling errors
	gl.DebugMessageCallback(func(source, gltype, id, severity uint32, length int32, message string, userParam unsafe.Pointer) {
		fmt.Println(source, gltype, severity, id, length, message, userParam)
	}, nil)
	// Create an OpenGL "Program" and link it for current drawing
	program, err = newProg(string(vertexShader), string(fragmentShader))
	if err != nil {
		panic(err)
	}
	// Check for the version
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL Version", version)
	// Main draw loop
	// Draw a Shape
	shape := NewShape(mgl32.Ident4(), program)
	for i := float64(0); i < 2*pi; i += pi/10 {
		shape.Pts = append(shape.Pts, P(0.7*float32(math.Cos(i)), 0.7*float32(math.Sin(i)), 1))
	} 
	// Generate the Vao for the shape
	shape.GenVao()
	shape.SetTypes(gl.LINE_LOOP)
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
	for !window.ShouldClose() {
		time.Sleep(time.Second / fps)
		// Clear everything that was drawn previously
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// Actually draw something
	//	shape.Draw()
		// display everything that was drawn
		window.SwapBuffers()
		// check for any events
		glfw.PollEvents()
	}
}
