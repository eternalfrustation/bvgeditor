package main

import (
	"encoding/binary"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sqweek/dialog"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"unsafe"
)

const (
	W         = 500
	H         = 500
	fps       = time.Second / 60
	pi        = 3.1415926535897932384626433832795028841971693993751058209749445923078164062862089986280348253421170679
	viewRange = 1000
	// The first point of the array of the Vectors in the ray struct
	// is used as initial point for subsequent rays, for eg.
	// Consider for following array : [{0, 1}, {13, 23}, {23, 24}, {1, 23}]
	// There will be a total of 3 rays constructed from the array and
	// they will be intersecting [{0, 1}, {13, 23}], [{0, 1}, {23, 24}]
	// and [{0, 1}, {1, 23}] respectively
	RAY_TYPE_CENTERED = 0x0
	// The initial point in the array of vectors in the ray struct
	// is every other vector, or the index has the index 2n where
	// n is the number of ray being considered, for eg.

	// Consider for following array : [{0, 1}, {13, 23}, {23, 24}, {1, 23}]
	// There will be a total of 2 rays constructed from the array and
	// they will be intersecting [{0, 1}, {13, 23}] and [{23, 24}, {1, 23}]
	// respectively
	RAY_TYPE_STRIP = 0x1
	pointByteSize  = int32(13 * 4)
)

var (
	viewMat        mgl32.Mat4
	projMat        mgl32.Mat4
	defaultViewMat mgl32.Mat4
	AddState       byte
	program        uint32
	MouseX         float64
	MouseY         float64
	CurrPoint      mgl32.Vec2
	Btns           []*Button
	BtnState       = byte('C')
	eyePos         mgl32.Vec3
	LookAt         mgl32.Vec3
	MouseRay       *Ray
	framesDrawn    int
	endianness     binary.ByteOrder
	Ident          = mgl32.Ident4()
)
var i int

func main() {
	var i int32 = 0x1
	bs := (*[4]byte)(unsafe.Pointer(&i))
	if bs[0] == 0 {
		endianness = binary.BigEndian
	} else {
		endianness = binary.LittleEndian
	}
	//	bvgPath, err := dialog.File().Load()
	//	orDie(err)
	//	fmt.Println(LoadBvg(bvgPath), bvgPath )

	dialog.File().Title("Select the BVG file").Load()
	runtime.LockOSThread()
	orDie(glfw.Init())
	//lotsOfPoints := DecodeTanishqsWierdFormat("lorenz_output3.txt")
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
		panic(fmt.Sprintf("%d, %d, %d, %d, %d, %s \n", source, gltype, severity, id, length, message))

	}, nil)
	// Create an OpenGL "Program" and link it for current drawing
	prog, err := newProg(string(vertexShader), string(fragmentShader))
	orDie(err)
	// Check for the version
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL Version", version)
	// Main draw loop

	// Set the refresh function for the window
	// Use this program
	gl.UseProgram(prog)
	// Calculate the projection matrix
	projMat = mgl32.Ident4()
	// set the value of Projection matrix
	UpdateUniformMat4fv("projection", program, &projMat[0])
	// Set the value of view matrix
	UpdateView(
		mgl32.Vec3{0, 0, -1},
		mgl32.Vec3{0, 0, 1},
	)
	program = prog
	// GLFW Initialization
	CurrPoint = mgl32.Vec2{0, 0}
	eyePos = mgl32.Vec3{0, 0, 1}

	// Get a font
	fnt := NewFont("font/space.ttf", "1234567890,./';[]{}|:\"<>?!@#$%^&*()_+-=qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM~`ā ", IntTo26_6(20))
	fnt.GlyphMap['ā'].SetTypes(gl.LINES)

	var bvgShapes []Drawable
	b := NewButton(-1, -1, 0, 0, window, "Click!", func(w *glfw.Window, mx, my float64, h []*mgl32.Vec3, v [][3]*mgl32.Vec3) {

		ex, err := os.Executable()
		orDie(err)
		exPath := filepath.Dir(ex)
		fmt.Println("Got the current path")
		bvgPath, err := dialog.File().Title("Select the BVG file").SetStartDir(exPath).Load()
		fmt.Println("Got the file path")
		if err != nil {
			fmt.Println("Fk, there was an err")
			dialog.Message("No valid file", "You need to provide a bvg file").Title("Select Bvg").Error()
			w.SetShouldClose(true)
			w.Destroy()
			os.Exit(0)
		}
		fmt.Println("There was no err")
		fmt.Println(bvgPath)
		bvgStruct := LoadBvg(bvgPath)
		fmt.Println("Loaded the bvg")
		bvgShapes = BvgS(bvgStruct)
		fmt.Println("Converted to shapes")
		for i := 0; i < 10; i++ {
			bvgShapes[i].GenVao()
		}
		fmt.Println("Genereated Vao")
		if len(bvgShapes) != 0 {
			for i := 0; i < 10; i++ {
				circ, _ := bvgShapes[i].(*Circle)
				fmt.Println(circ.Center.P)
			}
		}
	}, fnt)
	b.Geometry.Triangulate()
	modelMat := mgl32.Ident4()
	UpdateUniformMat4fv("model", program, &modelMat[0])
	window.SetKeyCallback(HandleKeys)
	window.SetCursorPosCallback(HandleMouseMovement)
	window.SetMouseButtonCallback(HandleMouseButton)
	window.SetRefreshCallback(Refresh)
	go func() {
		for {
			time.Sleep(time.Second)
			//	fmt.Printf("FPS: %d \r", framesDrawn)
			framesDrawn = 0
		}
	}()
	Btns = append(Btns, b)
	b.GenVao()
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	//lotsOfPoints.GenVao()
	for !window.ShouldClose() {
		time.Sleep(fps)
		// Clear everything that was drawn previously
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// Actually draw something
		//		b.Draw()
		if len(bvgShapes) != 0 {
			for i := 0; i < 10; i++ {
				bvgShapes[i].Draw()
			}
		}
		framesDrawn++
		b.Draw()
		//		fnt.GlyphMap['e'].Draw()
		// display everything that was drawn
		window.SwapBuffers()
		// check for any events
		glfw.PollEvents()
	}
}
