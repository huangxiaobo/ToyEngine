package engine

import (
	"fmt"
	_ "image/png"
	"os"
	"reflect"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/huangxiaobo/toy-engine/engine/model"
	"github.com/huangxiaobo/toy-engine/engine/platforms"
	"github.com/huangxiaobo/toy-engine/engine/text"
	"github.com/huangxiaobo/toy-engine/engine/window"
	"github.com/inkyblackness/imgui-go/v4"

	"github.com/huangxiaobo/toy-engine/engine/camera"
	"github.com/huangxiaobo/toy-engine/engine/config"
	"github.com/huangxiaobo/toy-engine/engine/light"
	"github.com/huangxiaobo/toy-engine/engine/logger"
)

const (
	millisPerSecond = 1000
	sleepDuration   = time.Millisecond * 25
)

type World struct {
	context  *imgui.Context
	platform *platforms.SDL
	imguiio  imgui.IO
	renderer *platforms.OpenGL3

	Light      *light.PointLight
	renderObjs []model.RenderObj
	Camera     *camera.Camera
	Text       *text.Text

	bRun bool
}

func (w *World) initSDL() {
	var err error

	windowWidth := config.Config.WindowWidth
	windowHeight := config.Config.WindowHeight
	w.platform, err = platforms.NewSDL(w.imguiio, platforms.SDLClientAPIOpenGL4, windowWidth, windowHeight)
	if err != nil {
		panic(err)
	}

	w.renderer, err = platforms.NewOpenGL3(w.imguiio)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}

}

func (w *World) initGL() {
	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	logger.Info("OpenGL version", version)

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)

	gl.Enable(gl.SAMPLES)

	// 只显示正面 , 不显示背面
	// gl.Enable(gl.CULL_FACE)

	// 设置顺时针方向 CW : Clock Wind 顺时针方向
	// 默认是 GL_CCW : Counter Clock Wind 逆时针方向
	gl.FrontFace(gl.CCW)

	// 设置线框模式
	// 设置了该模式后 , 之后的所有图形都会变成线
	//gl.PolygonMode(gl.FRONT, gl.LINE)

	// 设置点模式
	// 设置了该模式后 , 之后的所有图形都会变成点
	// glPolygonMode(GL_FRONT, GL_POINT);

}

func (w *World) Init() error {

	w.context = imgui.CreateContext(nil)

	w.imguiio = imgui.CurrentIO()

	w.initSDL()
	//w.initGL()

	// 初始化摄像机
	w.Camera = new(camera.Camera)
	w.Camera.Init(mgl32.Vec3{0.0, 50.0, 50.0}, mgl32.Vec3{-0.0, -0.0, -0.0})

	// 初始化灯光
	w.Light = &light.PointLight{Position: mgl32.Vec4{0, 50.0, 0, 0.0}, Color: mgl32.Vec3{1.0, 1.0, 1.0}}
	w.Light.Init()

	// Text
	w.Text, _ = text.NewText("Toy引擎", 32)

	w.bRun = true
	return nil
}

func (w *World) Destroy() {
	w.renderer.Dispose()
	w.context.Destroy()
	w.platform.Dispose()

}

// Platform covers mouse/keyboard/gamepad inputs, cursor shape, timing, windowing.
type Platform interface {
	// ShouldStop is regularly called as the abort condition for the program loop.
	ShouldStop() bool
	// ProcessEvents is called once per render loop to dispatch any pending events.
	ProcessEvents()
	// DisplaySize returns the dimension of the display.
	DisplaySize() [2]float32
	// FramebufferSize returns the dimension of the framebuffer.
	FramebufferSize() [2]float32
	// NewFrame marks the begin of a render pass. It must update the imgui IO state according to user input (mouse, keyboard, ...)
	NewFrame()
	// PostRender marks the completion of one render pass. Typically this causes the display buffer to be swapped.
	PostRender()
	// ClipboardText returns the current text of the clipboard, if available.
	ClipboardText() (string, error)
	// SetClipboardText sets the text as the current text of the clipboard.
	SetClipboardText(text string)
}

type clipboard struct {
	platform Platform
}

func (board clipboard) Text() (string, error) {
	return board.platform.ClipboardText()
}

func (board clipboard) SetText(text string) {
	board.platform.SetClipboardText(text)
}

func (w *World) Run() {
	imgui.CurrentIO().SetClipboard(clipboard{platform: w.platform})

	mainWindow := window.NewWindowMain()
	for _, renderObj := range w.renderObjs {
		name := reflect.ValueOf(renderObj).Elem().FieldByName("Name").String()
		id := reflect.ValueOf(renderObj).Elem().FieldByName("Id").String()

		fmt.Printf("name: %s, id: %s\n", name, id)
		mainWindow.AddModelItem(window.ModelItem{Name: name, Id: id})
	}

	mainWindow.NotifyModelItemChange(func(item window.ModelItem) {
		window.InitWindowMaterial()

		var targetRenderObj model.RenderObj = nil
		for _, renderObj := range w.renderObjs {
			name := reflect.ValueOf(renderObj).Elem().FieldByName("Name").String()
			if name == item.Name {
				targetRenderObj = renderObj
				break
			}
		}
		if targetRenderObj == nil {
			return
		}
		value := reflect.ValueOf(targetRenderObj).Elem().FieldByName("Position")
		window.AddMaterialAttr("Position", value.Interface().(mgl32.Vec3))
		value = reflect.ValueOf(targetRenderObj).Elem().FieldByName("Scale")
		window.AddMaterialAttr("Scale", value.Interface().(mgl32.Vec3))
	})

	for !w.platform.ShouldStop() {
		w.platform.ProcessEvents()

		// Signal start of a new frame
		w.platform.NewFrame()
		imgui.NewFrame()

		mainWindow.ShowWindowMain()
		window.ShowWindowMaterial()

		// Rendering
		imgui.Render() // This call only creates the draw data list. Actual rendering to framebuffer is done below.

		w.renderer.PreRender([3]float32{0.8, 0.85, 0.85})

		projection := mgl32.Perspective(
			mgl32.DegToRad(w.Camera.Zoom),
			float32(config.Config.WindowHeight/config.Config.WindowHeight),
			0.1,
			100.0,
		)
		view := w.Camera.GetViewMatrix()
		model := mgl32.Ident4()
		//mvp := projection.Mul4(view).Mul4(model)

		//w.DrawAxis()
		w.DrawLight()
		// Update
		elapsed := 0.01

		for _, renderObj := range w.renderObjs {
			renderObj.Update(elapsed)
			renderObj.PreRender()
			renderObj.Render(projection, model, view, &w.Camera.Position, w.Light)
			renderObj.PostRender()
		}

		// Logo
		w.Text.Render(0, 50, mgl32.Vec4{1.0, 1.0, 1.0, 1.0})

		// Maintenance
		w.renderer.Render(w.platform.DisplaySize(), w.platform.FramebufferSize(), imgui.RenderedDrawData())
		w.platform.PostRender()

		// sleep to avoid 100% CPU usage for this demo
		<-time.After(sleepDuration)
	}
}

func (w *World) DrawAxis() {
	// logger.Info("DrawAxis...")
}

func (w *World) DrawLight() {
	// RenderObj
	width := float32(config.Config.WindowWidth)
	height := float32(config.Config.WindowHeight)
	projection := mgl32.Perspective(
		mgl32.DegToRad(w.Camera.Zoom),
		width/height,
		0.1,
		100.0,
	)
	view := w.Camera.GetViewMatrix()
	model := mgl32.Ident4().Mul4(mgl32.Scale3D(1, 1, 1))

	position := w.Light.Position
	model = model.Add(mgl32.Translate3D(position.X(), position.Y(), position.Z()))

	w.Light.Render(projection, view, model)

}

func (w *World) AddRenderObj(renderObj model.RenderObj) {
	w.renderObjs = append(w.renderObjs, renderObj)
}
