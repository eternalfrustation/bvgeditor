package main

import (
	"fmt"
	"io/ioutil"
	"math"

	"github.com/eternalfrustation/bvg"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

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
	projMat = mgl32.Perspective(mgl32.DegToRad(120), float32(width)/float32(height), 0.001, 200)
	UpdateUniformMat4fv("projection", program, &projMat[0])
	resVec := mgl32.Vec2{float32(width), float32(height)}
	UniformLocation := gl.GetUniformLocation(program, gl.Str("u_resolution"+"\x00"))
	gl.UniformMatrix4fv(UniformLocation, 1, false, &resVec[0])
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
	offset := P(1, 0, 0)
	var prevI sfnt.GlyphIndex
	for i := len(s) - 1; i > -1; i-- {
		r := rune(s[i])
		switch r {
		case ' ':
			offset = offset.SetP(offset.X()+1, 0, 0)
		default:
			text.Pts = append(text.Pts, offset.MassOffset(f.GlyphMap[r].Pts...)...)
			glyph := &sfnt.Buffer{}
			I, err := f.TtfFont.GlyphIndex(glyph, r)
			orDie(err)
			// Scale the offset from truetype coords to opengl coords
			kerning, err := f.TtfFont.Kern(glyph, prevI, I, f.OgScale, font.HintingNone)
			//	orDie(err)
			bound, err := f.TtfFont.Bounds(glyph, f.OgScale, font.HintingNone)
			//	orDie(err)
			offX := float32(kerning.Round())
			offX /= float32(bound.Max.X.Round())
			prevI = I
			// Apply the offset
			offset = offset.SetP(offset.X()+1, 0, 0)
		}
	}
	reScaleFac := 1 / (offset.X() - 1)
	text = text.ReScale(reScaleFac, reScaleFac, 1)
	// Using LINES instead of LINE_LOOP to prevent lines joining between runes
	text.SetTypes(gl.LINES)
	return text
}

func orDie(err error) {
	if err != nil {
		panic(err)
	}
}

// Converts a given line Strip to line segments
func LineStripToSeg(pts ...*Point) []*Point {
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

func BezCurve(t float32, c0, c1, c2 *Point) (p []*Point) {
	Cs := PointsToMglPos(c0, c1, c2)
	for i := float32(0); i < float32(1.0); i += t {
		p = append(p, MglVecToPoint(mgl32.QuadraticBezierCurve3D(i, Cs[0], Cs[1], Cs[2])))
	}
	return p
}

func CubicBezCurve(t float32, c0, c1, c2, c3 *Point) (p []*Point) {
	Cs := PointsToMglPos(c0, c1, c2, c3)
	for i := float32(0); i < float32(1.0); i += t {
		p = append(p, MglVecToPoint(mgl32.CubicBezierCurve3D(i, Cs[0], Cs[1], Cs[2], Cs[3])))
	}
	return p
}
func MglVecToPoint(v mgl32.Vec3) *Point {
	return P(v[0], v[1], v[2])
}
func MglVecsToPoints(v ...mgl32.Vec3) (p []*Point) {
	for _, val := range v {
		p = append(p, P(val[0], val[1], val[2]))
	}
	return p
}

func PointsToMglPos(p ...*Point) (v []mgl32.Vec3) {
	for _, val := range p {
		v = append(v, val.P)
	}
	return v
}
func ShapePrint(s *Shape) {
	for _, val := range s.Pts {
		fmt.Println(*val)
	}
}

func IntTo26_6(i int) fixed.Int26_6 {
	return fixed.I(i)
}
func UnProject(viewM, projM mgl32.Mat4) mgl32.Mat4 {
	return projM.Mul4(viewM).Inv()
}

func UpdateView(lookingAt, eyePosition mgl32.Vec3) {
	viewMat = mgl32.LookAtV(
		lookingAt,
		eyePosition,
		mgl32.Vec3{0, 1, 0},
	)
	UpdateUniformMat4fv("view", program, &viewMat[0])
}

func RayTriangleCollision(ray [2]*mgl32.Vec3, triangle [3]*mgl32.Vec3) (bool, mgl32.Vec3) {
	Epsl := mgl32.Epsilon
	Null := mgl32.Vec3{mgl32.NaN, mgl32.NaN, mgl32.NaN}
	var edge1, edge2, h, s, q mgl32.Vec3
	var a, f, u, v float32
	edge1 = triangle[1].Sub(*triangle[0])
	edge2 = triangle[2].Sub(*triangle[0])
	h = ray[1].Cross(edge2)
	a = edge1.Dot(h)
	if a > -Epsl && a < Epsl {
		return false, Null
	}
	f = 1 / a
	s = ray[0].Sub(*triangle[0])
	u = f * s.Dot(h)
	if u < 0 || u > 1 {
		return false, Null
	}
	q = s.Cross(edge1)
	v = f * ray[1].Dot(q)
	if v < 0 || u+v > 1 {
		return false, Null
	}
	t := f * edge2.Dot(q)
	if t > Epsl {
		return true, ray[0].Add(ray[1].Mul(t))
	} else {
		return false, Null
	}
}

func LoadBvg(path string) *bvg.Bvg {
	bvgBytes, err := ioutil.ReadFile(path)
	orDie(err)
	LoadedBvg, err := bvg.Decode(bvgBytes)
	orDie(err)
	return LoadedBvg
}

//
func BvgToShapes(b *bvg.Bvg) []*Shape {
	// Sum of number of all the individual shapes in b
	// +2 for lines and points
	index := 0
	shapes := make([]*Shape, len(b.Bezs)+len(b.Circles)+len(b.LineStrips)+len(b.Polys)+2)
	for _, v := range b.Circles {
		var points []*Point
		Verts := 63
		points = make([]*Point, Verts+1)
		points[0] = PC(float32(v.P1.X), float32(v.P1.Y), 1,
			float32(v.P1.R)/255, float32(v.P1.G)/255, float32(v.P1.B)/255, float32(v.P1.A)/255)
		dist := v.P1.Dist(v.P2)
		for i := 0; i < Verts; i++ {
			x := float32(v.P1.X + math.Cos(2*pi*float64(i)/float64(Verts))*dist)
			y := float32(v.P1.Y + math.Sin(2*pi*float64(i)/float64(Verts))*dist)
			points[i+1] = PC(x, y, 1, 
			float32(v.P2.R)/255, float32(v.P2.G)/255, float32(v.P2.B)/255, float32(v.P2.A)/255)
		}
		shapes[index] = NewShape(mgl32.Ident4(), program, points...)
		shapes[index].SetTypes(gl.TRIANGLE_STRIP)
		index++
	}
	return shapes
}
