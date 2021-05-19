package main

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
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

type Shape struct {
	// Points making up the shape
	Pts        []*Point
	ModelMat   mgl32.Mat4
	Vao        uint32
	Vbo        uint32
	Prog       uint32
	Type       uint32
	Primitives int32
}

func NewShape(mat mgl32.Mat4, prog uint32, pts ...*Point) *Shape {
	return &Shape{
		Pts:      pts,
		ModelMat: mat,
		Prog:     prog,
	}
}

func (p *Point) Arr() []float32 {
	return []float32{
		p.P[0], p.P[1], p.P[2],
		p.C[0], p.C[1], p.C[2], p.C[3],
		p.N[0], p.N[1], p.N[2],
	}
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
	var vbo uint32
	// Generate the buffer for the Vertex data
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	// Fill the buffer with the Points data in our shape
	gl.BufferData(gl.ARRAY_BUFFER, 40*len(s.Pts), gl.Ptr(s.PointData()), gl.STATIC_DRAW)
	var vao uint32
	// Generate our Vertex Array
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	// At index 0, Put all the Position data
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 40, nil)
	// At index 1, Put all the Color data
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 40, gl.PtrOffset(12))
	// At index 2, Put all the Normal's data
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 40, gl.PtrOffset(28))
	// store the Vao and Vbo representatives in the shape
	s.Vbo = vbo
	s.Vao = vao
	// Initialize the model matrix
	s.ModelMat = mgl32.Ident4()

}

func (s *Shape) SetTypes(mode uint32) {
	s.Type = mode
	s.Primitives = int32(len(s.Pts))
}

func (s *Shape) Draw() {
	gl.BindVertexArray(s.Vao)
	gl.DrawArrays(s.Type, 0, s.Primitives)
}

type Button struct {
	Win      *glfw.Window
	Geometry *Shape
	Text     string
	CB       Callback
}

type Callback func(w *glfw.Window, btn *Button, MX, MY float64)

type Font struct {
	GlyphMap map[rune]*Shape
	GlyphArr []*truetype.GlyphBuf
	TtfFont  *truetype.Font
	OgScale  fixed.Int26_6
}

func NewButton(x1, y1, x2, y2 float32, w *glfw.Window, text string, cb Callback) *Button {
	b := new(Button)
	b.Geometry = NewShape(mgl32.Ident4(), gl.TRIANGLE_STRIP,
		P(x1, y1, 1),
		P(x2, y1, 1),
		P(x2, y2, 1),
		P(x1, y2, 1),
	)
	b.Win = w
	b.Text = text
	b.CB = cb
	return b
}

// This function creates a new Font to be used by TextToShape function
// NOTE: This function is not very memory efficient, donot call this in loop
func NewFont(path string) *Font {
	// Inittialize a new Font struct
	f := new(Font)
	f.GlyphMap = make(map[rune]*Shape)
	f.GlyphArr = make([]*truetype.GlyphBuf, 512)
	// Read and parse the file provided
	fontFile, err := ioutil.ReadFile(path)
	orDie(err)
	ttFont, err := truetype.Parse(fontFile)
	orDie(err)
	f.TtfFont = ttFont
	// If Default scale is 0, set it
	if f.OgScale == 0 {
		f.OgScale = fixed.I(64)
	}
	// Get the glyphs from rune 0 to 256 and create shapes out of them
	// and store them in the Font struct
	for i := 0; i < 512; i++ {
		// Initialize a new glyph for rune i, with the provided scale and no hinting
		glyph := &truetype.GlyphBuf{}
		err = glyph.Load(ttFont, f.OgScale, ttFont.Index(rune(i)), font.HintingNone)
		// Add the glyph to Font if needed elesewhere
		f.GlyphArr[i] = glyph
		f.GlyphMap[rune(i)] = NewShape(mgl32.Ident4(), program)
		// If the given rune has no shape in it, then give it a line
		// This happens in case of space, escape codes and invalid characters
		if len(glyph.Points) == 0 {
			f.GlyphMap[rune(i)].Pts = make([]*Point, 2)
			f.GlyphMap[rune(i)].Pts[0] = P(-1, -1, 1)
			f.GlyphMap[rune(i)].Pts[1] = P(1, -1, 1)
		} else {
			// if there is stuff in the glyph, allocate the glyphmap shape with points
			f.GlyphMap[rune(i)].Pts = make([]*Point, 2*len(glyph.Points))
			// Get the bounds of the glyph
			bound := glyph.Bounds.Max.Sub(glyph.Bounds.Min)
			maxX, maxY := bound.X.Round(), bound.Y.Round()
			// Get the First Point from the glyph Points slice
			firstPoint := glyph.Points[0]
			// Scale the font Coords to (-1, 1) system
			x, y := float32(firstPoint.X.Round())/float32(maxX), float32(firstPoint.Y.Round())/float32(maxY)
			// Set the first and last point the same
			f.GlyphMap[rune(i)].Pts[0] = P(x, y, 1)
			f.GlyphMap[rune(i)].Pts[len(f.GlyphMap[rune(i)].Pts)-1] = P(x, y, 1)
			for j := 1; j < len(glyph.Points); j++ {
				// Get the Current point
				val := glyph.Points[j]
				// Scale its coords to -1 to 1
				val.Flags
				x, y = float32(val.X.Round())/float32(maxX), float32(val.Y.Round())/float32(maxY)
				// Set the Corresponding points
				// for point i in glyph's point slice, 
				// 2i and 2i-1 are the corresponding points
				f.GlyphMap[rune(i)].Pts[2*j-1] = P(x, y, 1)
				f.GlyphMap[rune(i)].Pts[2*j] = P(x, y, 1)
			}
		}

		f.GlyphMap[rune(i)].SetTypes(gl.LINES)
		//	f.GlyphMap[rune(i)].GenVao()
		orDie(err)
	}
	return f
}
