package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"io/ioutil"
)

var (
	ProjMat    = mgl32.Ident4()
	VeiwMat    = mgl32.Ident4()
	ProjMatVao uint32
	VeiwMatVao uint32
)

type Point struct {
	// Position Vectors
	P mgl32.Vec3
	// Color Vectors
	C mgl32.Vec4
	// Normal Vectors
	N mgl32.Vec3
}

func (p *Point) X() float32 {
	return p.P[0]
}

func (p *Point) Y() float32 {
	return p.P[1]
}

func (p *Point) Z() float32 {
	return p.P[2]
}

/* Returns a point with x, y, z as its position with white color and normal in the
positive z axis */
func P(x, y, z float32) *Point {
	return &Point{P: mgl32.Vec3{x, y, z},
		C: mgl32.Vec4{1, 1, 1, 1},
		N: mgl32.Vec3{0, 0, 1},
	}
}

/* Returns a point with x, y, z as its position,  r,g,b,a as red, green,
blue and alpha respectively and normal in the positive z axis direction */
func PC(x, y, z, r, g, b, a float32) *Point {
	return &Point{P: mgl32.Vec3{x, y, z},
		C: mgl32.Vec4{r, g, b, a},
		N: mgl32.Vec3{0, 0, 1},
	}
}

/* Returns a point with x, y, z as its position,  r,g,b,a as red, green,
blue and alpha respectively and normal in the direction of normal of i,j,k */
func PCN(x, y, z, r, g, b, a, i, j, k float32) *Point {
	return &Point{P: mgl32.Vec3{x, y, z},
		C: mgl32.Vec4{r, g, b, a},
		N: mgl32.Vec3{i, j, k}.Normalize(),
	}
}

/* NOTE: This function returns a new Point with the given position */
func (p *Point) SetP(x, y, z float32) *Point {
	return &Point{P: mgl32.Vec3{x, y, z},
		C: p.C,
		N: p.N,
	}
}

/* NOTE: This function returns a new Point with the given Color */
func (p *Point) SetC(r, g, b, a float32) *Point {
	return &Point{P: p.P,
		C: mgl32.Vec4{r, g, b, a},
		N: p.N,
	}
}

/* NOTE: This function returns a new Point with the given Normal */
func (p *Point) SetN(i, j, k float32) *Point {
	return &Point{P: p.P,
		C: p.C,
		N: mgl32.Vec3{i, j, k},
	}
}

/* Offsets all of the given points with the positional coords of
the parent point
NOTE: This function returns the new points
*/

func (p *Point) MassOffset(pts ...*Point) []*Point {
	Offseted := make([]*Point, len(pts))
	for i, val := range pts {
		Offseted[i] = P(0, 0, 0).SetP(val.X()+p.X(), val.Y()+p.Y(), val.Z()+p.Y())
		Offseted[i].C, Offseted[i].N = val.C, val.N
	}
	return Offseted
}

type Circle struct {
	// Center point determines the center of the circle
	// And the color of the center of the circle
	Center   *Point
	Vao      uint32
	Vbo      uint32
	IsFilled bool
	ModelMat *mgl32.Mat4
	// Edge point determines the radius and the
	// color gradient of the circle
	Edge *Point
}

func (s *Circle) PointData() []byte {
	return []byte{}
}

func (s *Circle) GenVao() {
	data := make([]byte, 40*2+1)
	data[0] = byte(0)
//	for i, f := range s.PointData() {

//	}
	var vbo uint32
	// Generate the buffer for the Vertex data
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	// Fill the buffer with the Points data in our shape
	gl.BufferData(gl.ARRAY_BUFFER, 40*2+1, gl.Ptr(data), gl.STATIC_DRAW)
	var vao uint32
	// Generate our Vertex Array
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	// At index 0, Put the type of the curve
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 1, gl.BYTE, false, 82, nil)
	// At index 0, Put all the Position data
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 40, gl.PtrOffset(1))
	// At index 1, Put all the Color data
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 40, gl.PtrOffset(12+1))
	// At index 2, Put all the Normal's data
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(3, 3, gl.FLOAT, false, 40, gl.PtrOffset(28+1))
	// store the Vao and Vbo representatives in the shape
	s.Vbo = vbo
	s.Vao = vao

}

type Ray struct {
	Pts  []*mgl32.Vec3
	Type uint8
}

func NewRay(RayType uint8, modelMat mgl32.Mat4, points ...mgl32.Vec3) *Ray {
	transformedPoints := make([]*mgl32.Vec3, len(points))
	for i, val := range points {
		transformedPoint := mgl32.TransformCoordinate(val, modelMat)
		transformedPoints[i] = &transformedPoint
	}
	return &Ray{
		Pts:  transformedPoints,
		Type: RayType,
	}
}

// Takes a shape and check for collison the the ray r, if there is collision
// IsColliding is true, CollidingAt is where the collision happend and
// s can only be of type TRIANGLES, TRIANGLE_STRIP, TRIANGLE_FAN
func (r *Ray) PolyCollide(s *Shape) (IsColliding bool, CollidingAt []*mgl32.Vec3, CollTri [][3]*mgl32.Vec3) {
	triang := make([]mgl32.Vec3, len(s.Triangulated))
	for i, v := range s.Triangulated {
		triang[i] = mgl32.TransformCoordinate(*v, s.ModelMat)
	}
	switch r.Type {
	case RAY_TYPE_CENTERED:
		InitVec := r.Pts[0]
		for i := 1; i < len(r.Pts); i++ {
			for j := 0; j < len(triang)/3; j++ {
				IsItColling, WhereIsIt := RayTriangleCollision([2]*mgl32.Vec3{InitVec, r.Pts[i]},
					[3]*mgl32.Vec3{&triang[3*j], &triang[3*j+1], &triang[3*j+2]},
				)
				if !IsColliding {
					IsColliding = IsItColling
					CollidingAt = append(CollidingAt, &WhereIsIt)
					CollTri = append(CollTri, [3]*mgl32.Vec3{&triang[3*j], &triang[3*j+1], &triang[3*j+2]})
				}
			}

		}
	}
	return IsColliding, CollidingAt, CollTri
}

type Shape struct {
	// Points making up the shape
	Pts          []*Point
	ModelMat     mgl32.Mat4
	Vao          uint32
	Vbo          uint32
	Prog         uint32
	Type         uint32
	Primitives   int32
	Triangulated []*mgl32.Vec3
}

func NewShape(mat mgl32.Mat4, prog uint32, pts ...*Point) *Shape {
	return &Shape{
		Pts:      pts,
		ModelMat: mat,
		Prog:     prog,
	}
}

func (s *Shape) Triangulate() {
	var triang []*mgl32.Vec3
	switch s.Type {
	case gl.TRIANGLES:
		triang = make([]*mgl32.Vec3, len(s.Pts))
		for i, v := range s.Pts {
			triang[i] = &v.P
		}
	case gl.TRIANGLE_FAN:
		triang = make([]*mgl32.Vec3, (len(s.Pts)-2)*3)
		InitVec := s.Pts[0].P
		n := 1
		for i := 0; i < len(triang)/3; i++ {
			triang[3*i] = &InitVec
			triang[3*i+1] = &s.Pts[n].P
			n++
			triang[3*i+2] = &s.Pts[n].P
		}
	case gl.TRIANGLE_STRIP:
		triang = make([]*mgl32.Vec3, (len(s.Pts)-2)*3)
		var prevV, prevPrevV *mgl32.Vec3
		prevPrevV = &s.Pts[0].P
		prevV = &s.Pts[1].P
		for i := 2; i < len(s.Pts); i++ {
			triang[(i-2)*3] = prevPrevV
			triang[(i-2)*3+1] = prevV
			triang[(i-2)*3+2] = &s.Pts[i].P
			prevPrevV = prevV
			prevV = &s.Pts[i].P

		}
	}
	s.Triangulated = triang
}

func (p *Point) Arr() []float32 {
	return []float32{
		p.P[0], p.P[1], p.P[2],
		p.C[0], p.C[1], p.C[2], p.C[3],
		p.N[0], p.N[1], p.N[2],
	}
}

// Do not use this function frequently,
// Instead use ModelMat to transform the shapes
func (p *Point) ReScale(x, y, z float32) *Point {
	return &Point{
		P: mgl32.Vec3{p.X() * x, p.Y() * y, p.Z() * z},
		C: p.C,
		N: p.N,
	}
}

// Do not use this function frequently,
// Instead use ModelMat to transform the shapes
func (s *Shape) ReScale(x, y, z float32) *Shape {
	S := NewShape(mgl32.Ident4(), program)
	ps := make([]*Point, len(s.Pts))
	for i, p := range s.Pts {
		ps[i] = p.ReScale(x, y, z)
	}
	S.Pts = ps
	return S
}

func (s *Shape) PointData() []float32 {
	var data []float32
	for _, p := range s.Pts {
		data = append(data, p.Arr()...)
	}
	return data
}

func (s *Shape) TransformData() []float32 {
	var data []float32

	for i, val := range s.ModelMat {
		data[i] = val
	}
	return data

}

func (s *Shape) GenVao() {
	pointSlice := s.PointData()
	floatBytes := make([]byte, len(pointSlice)*4+1)
	floatBytes[0] = byte(0)
	for i, f := range Float32SlicetoBytes(pointSlice) {
		floatBytes[i+1] = f
	}
	var vbo uint32
	// Generate the buffer for the Vertex data
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	// Fill the buffer with the Points data in our shape
	gl.BufferData(gl.ARRAY_BUFFER, 40*len(s.Pts), gl.Ptr(floatBytes), gl.STATIC_DRAW)
	var vao uint32
	// Generate our Vertex Array
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	// At index 0, Put the type of the shape
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 1, gl.INT, false, 41*int32(len(s.Pts)), nil)
	// At index 0, Put all the Position data
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 40, gl.PtrOffset(1))
	// At index 1, Put all the Color data
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 4, gl.FLOAT, false, 40, gl.PtrOffset(12+1))
	// At index 2, Put all the Normal's data
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(3, 3, gl.FLOAT, false, 40, gl.PtrOffset(28+1))
	// store the Vao and Vbo representatives in the shape
	s.Vbo = vbo
	s.Vao = vao

}

func (s *Shape) SetTypes(mode uint32) {
	s.Type = mode
	s.Primitives = int32(len(s.Pts))
}

func (s *Shape) Free() {
	gl.DeleteBuffers(1, &s.Vao)
	gl.DeleteVertexArrays(1, &s.Vao)
}

func (s *Shape) Draw() {
	UpdateUniformMat4fv("model", program, &s.ModelMat[0])
	gl.BindVertexArray(s.Vao)
	gl.DrawArrays(s.Type, 0, s.Primitives)
}

type Button struct {
	Win       *glfw.Window
	Geometry  *Shape
	Text      string
	TextShape *Shape
	CB        Callback
}

type Callback func(w *glfw.Window, MX, MY float64, click3D []*mgl32.Vec3, NearTri [][3]*mgl32.Vec3)

type Font struct {
	GlyphMap map[rune]*Shape
	TtfFont  *sfnt.Font
	OgScale  fixed.Int26_6
}

func NewButton(x1, y1, x2, y2 float32, w *glfw.Window, text string, cb Callback, font *Font) *Button {
	b := new(Button)
	b.Geometry = NewShape(mgl32.Ident4(), program,
		PC(x1, y1, 1, 1, 0, 1, 1),
		PC(x1, y2, 1, 1, 0, 1, 1),
		PC(y2, x1, 1, 1, 0, 1, 1),
		PC(x2, y2, 1, 1, 0, 1, 1),
	)
	b.Geometry.SetTypes(gl.TRIANGLE_STRIP)
	b.Win = w
	b.Text = text
	b.CB = cb
	b.TextShape = TextToShape(font, text)
	b.TextShape.ModelMat = mgl32.Translate3D(x1-x2, (y1-y2)/2, 0)
	ShapePrint(b.Geometry)
	return b
}

func (b *Button) Draw() {
	b.Geometry.Draw()
	b.TextShape.Draw()
}

func (b *Button) GenVao() {
	b.Geometry.GenVao()
	b.TextShape.GenVao()
}

// This function creates a new Font to be used by TextToShape function
// Supply the characters to load in runes
// NOTE: This function is not very memory efficient, donot call this in loop
func NewFont(path string, runes string, OgScale fixed.Int26_6) *Font {
	// Inittialize a new Font struct
	f := new(Font)
	f.OgScale = OgScale
	f.GlyphMap = make(map[rune]*Shape)
	// Read and parse the file provided
	fontFile, err := ioutil.ReadFile(path)
	orDie(err)
	ttFont, err := sfnt.Parse(fontFile)
	orDie(err)
	f.TtfFont = ttFont
	// If Default scale is 0, set it to a default value
	if f.OgScale == 0 {
		f.OgScale = fixed.I(64)
	}
	boundR, err := ttFont.Bounds(nil, f.OgScale, font.HintingNone)
	orDie(err)
	bound := boundR.Max.Sub(boundR.Min)
	maxX, maxY := bound.X.Round(), bound.Y.Round()

	// Get the glyphs from rune 0 to 512 and create shapes out of them
	// and store them in the Font struct
	for _, i := range runes {
		// Initialize a new glyph for rune i, with the provided scale and no hinting
		glyph := &sfnt.Buffer{}
		I, err := ttFont.GlyphIndex(glyph, rune(i))
		orDie(err)
		segs, err := ttFont.LoadGlyph(glyph, I, f.OgScale, nil)
		// Add the glyph to Font if needed elesewhere
		f.GlyphMap[rune(i)] = NewShape(mgl32.Ident4(), program)
		// If the given rune has no shape in it, then give it a line
		// This happens in case of space, escape codes and invalid characters
		if len(segs) == 0 {
			f.GlyphMap[rune(i)].Pts = make([]*Point, 2)
			f.GlyphMap[rune(i)].Pts[0] = P(-1, -1, 1)
			f.GlyphMap[rune(i)].Pts[1] = P(1, -1, 1)
		} else {
			// Get the bounds of the glyph
			// Make a point to store the coords of SegmentOpMoveTo
			prevP := P(0, 0, 0)
			for _, val := range segs {
				// Scale its coords to -1 to 1
				x0, y0 := -float32(val.Args[0].X.Round())/float32(maxX), -float32(val.Args[0].Y.Round())/float32(maxY)
				x1, y1 := -float32(val.Args[1].X.Round())/float32(maxX), -float32(val.Args[1].Y.Round())/float32(maxY)
				x2, y2 := -float32(val.Args[2].X.Round())/float32(maxX), -float32(val.Args[2].Y.Round())/float32(maxY)
				//fmt.Println(x1, y1)
				switch val.Op {

				case sfnt.SegmentOpMoveTo:
					prevP = P(x0, y0, 1)
				case sfnt.SegmentOpLineTo:
					f.GlyphMap[rune(i)].Pts = append(f.GlyphMap[rune(i)].Pts,
						P(prevP.X(), prevP.Y(), 1),
						P(x0, y0, 1))
					prevP = P(x0, y0, 1)
				case sfnt.SegmentOpQuadTo:
					f.GlyphMap[rune(i)].Pts = append(f.GlyphMap[rune(i)].Pts,
						LineStripToSeg(BezCurve(8/float32(f.OgScale),
							P(prevP.X(), prevP.Y(), 1),
							P(x0, y0, 1),
							P(x1, y1, 1))...)...)

					prevP = P(x1, y1, 1)

				case sfnt.SegmentOpCubeTo:
					f.GlyphMap[rune(i)].Pts = append(f.GlyphMap[rune(i)].Pts,
						LineStripToSeg(CubicBezCurve(8/float32(f.OgScale),
							P(prevP.X(), prevP.Y(), 1),
							P(x0, y0, 1),
							P(x1, y1, 1),
							P(x2, y2, 1))...)...)

					prevP = P(x2, y2, 1)
				}
			}
		}

		f.GlyphMap[rune(i)].SetTypes(gl.LINES)
		//	f.GlyphMap[rune(i)].GenVao()
		orDie(err)
	}
	return f
}
