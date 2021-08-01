package mesh

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"toy/engine"
	"toy/engine/config"
	"toy/engine/loader"
	"toy/engine/logger"
	"toy/engine/shader"
)

type WavefrontObject struct {
	Name         string
	ObjFilePath  string
	VertFilePath string
	FragFilePath string

	Position [3]float32
	Scale    [3]float32
	model    mgl32.Mat4

	meshData          *engine.MeshData
	shader            *shader.Shader
	projectionUniform int32
	viewUniform       int32
	modelUniform      int32
	wvpUniform               int32

	vao          uint32
	vbo          uint32
	ebo          uint32
	nbo          uint32
	tbo          uint32
	vertAttrib   uint32
	normalAttrib uint32

	// Draw
	DrawMode uint32
}

func (wfo *WavefrontObject) Init(w *engine.World) {
	wfo.meshData = &engine.MeshData{}
	switch wfo.Name {
	case "ground":
		wfo.meshData = GenGroundMeshData()
	default:
		if err := loader.LoadWavefrontObj(wfo.ObjFilePath, wfo.meshData); err != nil {
			logger.Error(err)
			return
		}
	}

	wfo.shader = &shader.Shader{VertFilePath: wfo.VertFilePath, FragFilePath: wfo.FragFilePath}
	if err := wfo.shader.Init(); err != nil {
		logger.Error(err)
		return
	}

	program := wfo.shader.Use()

	// Shader
	wfo.projectionUniform = gl.GetUniformLocation(program, gl.Str("projection\x00"))
	wfo.viewUniform = gl.GetUniformLocation(program, gl.Str("view\x00"))
	wfo.modelUniform = gl.GetUniformLocation(program, gl.Str("model\x00"))
	wfo.wvpUniform = gl.GetUniformLocation(program, gl.Str("gWVP\x00"))

	gl.BindFragDataLocation(program, 0, gl.Str("color\x00"))

	// Configure the vertex data
	gl.GenVertexArrays(1, &wfo.vao)
	gl.BindVertexArray(wfo.vao)

	// vert buff
	gl.GenBuffers(1, &wfo.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, wfo.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(wfo.meshData.Vertices)*4, gl.Ptr(&wfo.meshData.Vertices[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// normal buff
	gl.GenBuffers(1, &wfo.nbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, wfo.nbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(wfo.meshData.Normals)*4, gl.Ptr(&wfo.meshData.Normals[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// uv buff
	gl.GenBuffers(1, &wfo.tbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, wfo.tbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(wfo.meshData.Uvs)*2, gl.Ptr(&wfo.meshData.Uvs[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// index buff
	gl.GenBuffers(1, &wfo.ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, wfo.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(wfo.meshData.VertexIndices)*2, gl.Ptr(&wfo.meshData.VertexIndices[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// // Get an index for the attribute from the shader
	wfo.vertAttrib = uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
	wfo.normalAttrib = uint32(gl.GetAttribLocation(program, gl.Str("normal\x00")))

	// Unbind the buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.EnableVertexAttribArray(0)
	gl.BindVertexArray(0)

	wfo.model = mgl32.Ident4()
	wfo.model = wfo.model.Mul4(mgl32.Scale3D(wfo.Scale[0], wfo.Scale[1], wfo.Scale[2]))
	wfo.model = wfo.model.Add(mgl32.Translate3D(wfo.Position[0], wfo.Position[1], wfo.Position[2]))

	wfo.DrawMode = gl.TRIANGLES
}

func (wfo *WavefrontObject) Update(elapsed float64) {
	wfo.model = wfo.model.Mul4(mgl32.HomogRotate3DY(float32(elapsed)))
}

func (wfo *WavefrontObject) Render(w *engine.World) {
	// Render
	width := float32(config.Config.WindowWidth)
	height := float32(config.Config.WindowHeight)
	projection := mgl32.Perspective(
		mgl32.DegToRad(w.Camera.Zoom),
		width/height,
		0.1,
		100.0,
	)
	view := w.Camera.GetViewMatrix()
	mvp := projection.Mul4(view).Mul4(mgl32.Ident4())

	program := wfo.shader.Program
	// Shader
	gl.UseProgram(program)

	gl.UniformMatrix4fv(wfo.projectionUniform, 1, false, &projection[0])
	gl.UniformMatrix4fv(wfo.viewUniform, 1, false, &view[0])
	gl.UniformMatrix4fv(wfo.modelUniform, 1, false, &wfo.model[0])
	gl.UniformMatrix4fv(wfo.wvpUniform, 1, false, &mvp[0])

	gl.BindFragDataLocation(program, 0, gl.Str("color\x00"))
	lightPosAttrib := gl.GetUniformLocation(program, gl.Str("gLightPos\x00"))
	gl.Uniform4fv(lightPosAttrib, 1, &w.Light.Position[0])
	lightColorAttrib := gl.GetUniformLocation(program, gl.Str("gLightColor\x00"))
	gl.Uniform3fv(lightColorAttrib, 1, &w.Light.Color[0])
	viewPosAttrib := gl.GetUniformLocation(program, gl.Str("gViewPos\x00"))
	gl.Uniform3fv(viewPosAttrib, 1, &w.Camera.Position[0])

	// 开启顶点数组
	gl.BindVertexArray(wfo.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, wfo.vbo)
	gl.EnableVertexAttribArray(wfo.vertAttrib)
	gl.VertexAttribPointer(
		wfo.vertAttrib, // attribute index
		3,              // number of elements per vertex, here (x,y,z)
		gl.FLOAT,       // the type of each element
		false,          // take our values as-is
		0,              // no extra data between each position
		nil,            // offset of first element
	)

	gl.BindBuffer(gl.ARRAY_BUFFER, wfo.nbo)
	gl.EnableVertexAttribArray(wfo.normalAttrib)
	gl.VertexAttribPointer(
		wfo.normalAttrib, // attribute index
		3,                // number of elements per vertex, here (x,y,z)
		gl.FLOAT,         // the type of each element
		false,            // take our values as-is
		0,                // no extra data between each position
		nil,              // offset of first element
	)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, wfo.ebo)
	gl.DrawElements(
		wfo.DrawMode,                           // mode
		int32(len(wfo.meshData.VertexIndices)), // count
		gl.UNSIGNED_SHORT,                      // type
		nil,                                    // element array buffer offset
	)

	// gl.PointSize(10)
	// gl.DrawElements(
	// 	gl.POINTS,                              // mode
	// 	int32(len(wfo.meshData.VertexIndices)), // count
	// 	gl.UNSIGNED_SHORT,                      // type
	// 	nil,                                    // element array buffer offset
	// )
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.EnableVertexAttribArray(0)
	gl.BindVertexArray(0)
}
