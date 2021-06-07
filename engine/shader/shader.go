package shader

import (
	"fmt"
	"io/ioutil"
	"strings"
	"toy/engine/logger"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Shader struct {
	VsFilePath string
	FsFilePath string
	Program    uint32
}

func (s *Shader) Init() error {
	if s.VsFilePath == "" {
		s.VsFilePath = "./resource/cube.vs"
	}
	if s.FsFilePath == "" {
		s.FsFilePath = "./resource/cube.fs"
	}

	vsData, err := ioutil.ReadFile(s.VsFilePath)
	if err != nil {
		fmt.Println(err)
	}
	logger.Info(string(vsData))
	fsData, err := ioutil.ReadFile(s.FsFilePath)
	if err != nil {
		fmt.Println(err)
	}
	logger.Info(string(fsData))

	s.Program, err = s.NewProgram(string(vsData), string(fsData))
	if err != nil {
		panic(err)
	}
	return nil
}

func (s *Shader) NewProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	// 加载并编译shader
	vertexShader, err := s.CompileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := s.CompileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	// program
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func (s *Shader) CompileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csource, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csource, nil)
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

func (s *Shader) Use() uint32 {
	gl.UseProgram(s.Program)
	return s.Program
}
