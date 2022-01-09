package model

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/huangxiaobo/toy-engine/engine/logger"
	"github.com/huangxiaobo/toy-engine/engine/texture"
	"strconv"
	"sync"
	"unsafe"
)

type Vertex struct {
	Position  mgl32.Vec3
	Normal    mgl32.Vec3
	TexCoords mgl32.Vec2
	Tangent   mgl32.Vec3
	Bitangent mgl32.Vec3
}

type Mesh struct {
	Name string

	wg sync.WaitGroup

	Vertices []Vertex
	Indices  []uint32
	Textures []texture.Texture

	vao uint32
	vbo uint32
	ebo uint32

	BasePath string
}

func NewMesh(v []Vertex, i []uint32, t []texture.Texture) Mesh {
	m := Mesh{
		Vertices: v,
		Indices:  i,
		Textures: t,
	}
	//m.setup()
	return m
}

func (m *Mesh) setup() {
	// size of the Vertex struct
	dummy := m.Vertices[0]
	structSize := int(unsafe.Sizeof(dummy))
	structSize32 := int32(structSize)
	logger.Error("structSize: ", structSize, " structSize32: ", structSize32)

	// Configure the vertex data
	gl.GenVertexArrays(1, &m.vao)
	gl.GenBuffers(1, &m.vbo)
	gl.GenBuffers(1, &m.ebo)

	gl.BindVertexArray(m.vao)

	// vert buff 复制顶点数组到缓冲中供OpenGL使用
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(m.Vertices)*structSize, gl.Ptr(m.Vertices), gl.STATIC_DRAW)

	// indic buff, 复制索引数组到缓冲中供OpenGL使用
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(m.Indices)*GL_FLOAT32_SIZE, gl.Ptr(m.Indices), gl.STATIC_DRAW)

	// Set the vertex attribute pointers
	// Vertex Positions
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, structSize32, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Vertex Normals
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, structSize32, unsafe.Pointer(unsafe.Offsetof(dummy.Normal)))
	gl.EnableVertexAttribArray(1)

	// Vertex Texture Coords
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, structSize32, unsafe.Pointer(unsafe.Offsetof(dummy.TexCoords)))
	gl.EnableVertexAttribArray(2)

	// Vertex Tangent
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(3, 3, gl.FLOAT, false, structSize32, unsafe.Pointer(unsafe.Offsetof(dummy.Tangent)))
	// Vertex Bitangent
	gl.EnableVertexAttribArray(4)
	gl.VertexAttribPointer(4, 3, gl.FLOAT, false, structSize32, unsafe.Pointer(unsafe.Offsetof(dummy.Bitangent)))

	// Unbind the buffer
	gl.BindVertexArray(0)
}

func (m *Mesh) draw(program uint32) {
	// Bind appropriate textures
	var (
		diffuseNr  uint64
		specularNr uint64
		normalNr   uint64
		heightNr   uint64
		i          uint32
	)
	diffuseNr = 1
	specularNr = 1
	normalNr = 1
	heightNr = 1
	i = 0
	for i = 0; i < uint32(len(m.Textures)); i++ {
		gl.ActiveTexture(gl.TEXTURE0 + i) // Active proper texture unit before binding

		// Retrieve texture number (the N in diffuse_textureN)
		ss := ""
		switch m.Textures[i].TextureType {
		case "texture_diffuse":
			ss = ss + strconv.FormatUint(diffuseNr, 10) // Transfer GLuint to stream
			diffuseNr++
		case "texture_specular":
			ss = ss + strconv.FormatUint(specularNr, 10) // Transfer GLuint to stream
			specularNr++
		case "texture_normal":
			ss = ss + strconv.FormatUint(normalNr, 10) // Transfer GLuint to stream
			normalNr++
		case "texture_height":
			ss = ss + strconv.FormatUint(heightNr, 10) // Transfer GLuint to stream
			heightNr++
		}

		// Now set the sampler to the correct texture unit
		tu := m.Textures[i].TextureType + ss + "\x00"

		gl.Uniform1i(gl.GetUniformLocation(program, gl.Str(tu)), int32(i))
		// And finally bind the texture
		gl.BindTexture(gl.TEXTURE_2D, m.Textures[i].Id)
	}

	// Draw mesh
	gl.BindVertexArray(m.vao)
	gl.DrawElements(gl.TRIANGLES, int32(len(m.Indices)), gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)

	// Always good practice to set everything back to defaults once configured.
	for i = 0; i < uint32(len(m.Textures)); i++ {
		gl.ActiveTexture(gl.TEXTURE0 + i)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
}