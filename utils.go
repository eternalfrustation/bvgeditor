package main

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype/truetype"

	//	"github.com/go-gl/mathgl/mgl32"
	"strings"
)

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func newProg(vertShad, fragShad string) (uint32, error) {
	vertexShader, err := compileShader(vertShad, gl.VERTEX_SHADER)
	orDie(err)
	fragmentShader, err := compileShader(fragShad, gl.FRAGMENT_SHADER)
	orDie(err)
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(prog, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link prog: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return prog, nil
}

func UpdateUniformMat4fv(name string, prog uint32, value *float32) {
	UniformLocation := gl.GetUniformLocation(prog, gl.Str(name+"\x00"))
	gl.UniformMatrix4fv(UniformLocation, 1, false, value)

}

func Refresh(w *glfw.Window) {
	width, height := w.GetFramebufferSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	projectionMat := mgl32.Perspective(mgl32.DegToRad(120), float32(width)/float32(height), 0.001, 200)
	UpdateUniformMat4fv("projection", program, &projectionMat[0])
	fmt.Println(float32(width) / float32(height))
}

// This Algorithm was taken from http://www.jeffreythompson.org/collision-detection/poly-point.php
// aka idk how this works go on their website to find out
func PtPolyCollision(pt *Point, poly *Shape) bool {
	collision := false
	next := 0
	for i := 0; i < len(poly.Pts); i++ {
		next = i + 1
		if next == len(poly.Pts) {
			next = 0
		}
		Vc := poly.Pts[i]
		Vn := poly.Pts[next]
		if (Vc.Y() > pt.Y()) != (Vn.Y() > pt.Y()) && pt.X() < (Vn.X()-Vc.X())*(pt.Y()-Vc.Y())/(Vn.Y()-Vc.Y())+Vc.X() {
			collision = !collision
		}
	}
	return collision
}

func TextToShape(f *Font, s string) *Shape {
	text := NewShape(mgl32.Ident4(), program)
	offset := P(0, 0, 0)
	var prevI truetype.Index
	for _, r := range s {
		// Ultimate jank returns, Getting the Points from the glyph map,
		// then offsetting them as required, then converting them to
		// line Segments to be able to use them with lines
		fmt.Println(f.GlyphMap[r].Pts)
		text.Pts = append(text.Pts, offset.MassOffset(f.GlyphMap[r].Pts...)...)
		I := f.TtfFont.Index(r)
		// Scale the offset from truetype coords to opengl coords
		offX := float32(f.TtfFont.Kern(f.OgScale, prevI, I).Round())
		offX /= float32(f.TtfFont.Bounds(f.OgScale).Max.X.Round())
		prevI = I
		// Apply the offset
		offset = offset.SetP(offset.X()-1, 0, 0)
	}
	// Using LINES instead of LINE_LOOP to prevent lines joining between letters
	text.SetTypes(gl.LINES)
	return text
}

func orDie(err error) {
	if err != nil {
		panic(err)
	}
}

// Converts a given line loop to line segments
func LineLoopToSeg(pts ...*Point) []*Point {
	// Initialize the array
	ps := make([]*Point, 2*len(pts))
	// First and last element would be equal to the first element
	// Found experimentally
	ps[0] = pts[0]
	ps[len(ps)-1] = pts[len(pts)-1]
	for i := 1; i < len(pts); i++ {
		// One object of pts slice maps to two objects in
		// ps array
		ps[2*i-1], ps[2*i] = pts[i], pts[i]
	}
	return ps
}
