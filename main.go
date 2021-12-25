
package main

import (
_ "image/png"
"log"
"runtime"

"github.com/go-gl/mathgl/mgl32"

"github.com/huangxiaobo/toy-engine/engine"
"github.com/huangxiaobo/toy-engine/engine/mesh"
"github.com/huangxiaobo/toy-engine/engine/mesh/ground"
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {

	world := new(engine.World)
	err := world.Init()
	if err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer world.Destroy()

	world.AddRenderObj(&ground.Ground{
		WavefrontObject: mesh.WavefrontObject{
			Name:         "ground",
			ObjFilePath:  "",
			VertFilePath: "./resource/shader/ground/shader.vert",
			FragFilePath: "./resource/shader/ground/shader.frag",
		},
	})

	world.AddRenderObj(&mesh.Axis{
		Name:       "axis",
		VsFilePath: "./resource/shader/axis.vert",
		FsFilePath: "./resource/shader/axis.frag",
	})

	world.AddRenderObj(&mesh.WavefrontObject{
		Name:         "cube",
		ObjFilePath:  "./resource/model/icosphere.obj",
		VertFilePath: "./resource/shader/v2/shader.vert",
		FragFilePath: "./resource/shader/v2/shader.frag",
		Position: mgl32.Vec3{5, 0, -5},
		Scale: mgl32.Vec3{50, 50, 50},
	})

	// world.AddRenderObj(&mesh.WavefrontObject{
	// 	Name:         "icosphere",
	// 	ObjFilePath:  "./resource/model/icosphere.obj",
	// 	VertFilePath: "./resource/shader/v2/shader.vert",
	// 	FragFilePath: "./resource/shader/v2/shader.frag",
	// 	Position: mgl32.Vec3{-5, 0, 5},
	// 	Scale: mgl32.Vec3{4, 4, 4},
	// })

	world.Run()
}