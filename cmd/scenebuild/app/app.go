package app

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tgreiser/cymapper/cmd/scenebuild/fixture"

	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/application"
	"github.com/g3n/engine/util/logger"
	"github.com/g3n/engine/util/stats"
	"github.com/g3n/engine/window"
)

type App struct {
	*application.Application                    // Embedded standard application object
	log                      *logger.Logger     // Application logger
	labelFPS                 *gui.Label         // header FPS label
	stats                    *stats.Stats       // statistics object
	statsTable               *stats.StatsTable  // statistics table panel
	control                  *gui.ControlFolder // Pointer to gui control panel
	fs                       *FileSelect        // File select dialog
	sceneFS                  *FileSelect
	ed                       *ErrorDialog // Error dialog
	ambLight                 *light.Ambient
	fixtures                 []*fixture.Fixture
	selected                 int // selected fixture
	width                    *gui.Edit
	height                   *gui.Edit
	tlx                      *gui.Edit          // Top left x; x coordinate of top left corner of current fixture
	tly                      *gui.Edit          // Top left y
	brx                      *gui.Edit          // Bottom right x
	bry                      *gui.Edit          // Bottom right y
	sceneWidth               float32
	sceneHeight              float32
    zoom                     *gui.Slider
}

type IScreen interface {
	Initialize(*App) // Called once to initialize the demo
	Render(*App)     // Called at each frame for animations
}

// Command line options
// The standard application object may add other command line options
var (
	oNogui       = flag.Bool("nogui", false, "Do not show the GUI, only the specified demo")
	oHideFPS     = flag.Bool("hidefps", true, "Do now show calculated FPS in the GUI")
	oUpdateFPS   = flag.Uint("updatefps", 1000, "Time interval in milliseconds to update the FPS in the GUI")
	oLogs        = flag.String("logs", "", "Set log levels for packages. Ex: gui:debug,gls:info")
	oStats       = flag.Bool("stats", false, "Shows statistics control panel in the GUI")
	oRenderStats = flag.Bool("renderstats", false, "Shows gui renderer statistics in the console")
)

const (
	progName = "Cyma Scene Builder"
	execName = "cyscene"
	vmajor   = 0 // Major version number
	vminor   = 1 // Minor version number
)

func Create() *App {
	flag.Usage = usage

	// Creates standard application object
	a, err := application.Create(application.Options{
		Title:       progName,
		Width:       800,
		Height:      800,
		Fullscreen:  false,
		LogPrefix:   "CYSCENE",
		LogLevel:    logger.DEBUG,
		TargetFPS:   60,
		EnableFlags: true,
	})
	if err != nil {
		panic(err)
	}
	app := new(App)
	app.sceneWidth = 1280
	app.sceneHeight = 720
    app.selected = -1
	app.Application = a
	app.log = app.Log()
	app.log.Info("%s v%d.%d starting", progName, vmajor, vminor)
	app.stats = stats.NewStats(app.Gl())

	// Apply log levels to engine package loggers
	if *oLogs != "" {
		logs := strings.Split(*oLogs, ",")
		for i := 0; i < len(logs); i++ {
			parts := strings.Split(logs[i], ":")
			if len(parts) != 2 {
				app.log.Error("Invalid logs level string")
				continue
			}
			pack := strings.ToUpper(parts[0])
			level := strings.ToUpper(parts[1])
			path := "G3N/" + pack
			packlog := logger.Find(path)
			if packlog == nil {
				app.log.Error("No logger for package:%s", pack)
				continue
			}
			err := packlog.SetLevelByName(level)
			if err != nil {
				app.log.Error("%s", err)
			}
			app.log.Info("Set log level:%s for package:%s", level, pack)
		}
	}

	// Setup scene
	app.setupScene()

	// Builds user interface
	if *oNogui == false {
		app.buildGui()
	}

	// Subscribe to after render events to update the FPS
	app.Subscribe(application.OnAfterRender, func(evname string, ev interface{}) {
		// Update statistics
		if app.stats.Update(time.Second) {
			if app.statsTable != nil {
				app.statsTable.Update(app.stats)
			}
		}
		// Update render stats
		if *oRenderStats {
			stats := app.Renderer().Stats()
			if stats.Panels > 0 {
				app.log.Debug("render stats:%+v", stats)
			}
		}
		// Update FPS
		app.updateFPS()
	})
	return app
}

// setupScene resets the current scene for executing a new (or first) test
func (app *App) setupScene() {
	// Cancel next events and clear all window subscriptions
	app.Window().CancelDispatch()
	app.Window().ClearSubscriptions()

	// Dispose of all test scene children
	app.Scene().DisposeChildren(true)
	if app.Panel3D() != nil {
		app.Panel3D().GetPanel().DisposeChildren(true)
	}

	// Sets default background color
	app.Gl().ClearColor(0, 0, 0, 1.0)

	// Adds ambient light to the test scene
	app.ambLight = light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.5)
	app.Scene().Add(app.ambLight)

	// Sets perspective camera position
	//width, height := app.Window().Size()
	//aspect := float32(width) / float32(height)
	//camOrtho := camera.NewOrthographic(0, 640, 480, 0, 0.1, 100)


	vx := app.sceneWidth / 2
	vy := app.sceneHeight / 2
	app.CameraOrtho().SetPosition(vx, vy, 99)
	app.CameraOrtho().LookAt(&math32.Vector3{vx, vy, 0})
    if app.zoom != nil {
        app.CameraOrtho().SetZoom(app.zoom.Value() / 100)
    }

	// Default camera is perspective
	app.SetCamera(app.CameraOrtho())
	//app.SetOrbit(control.NewOrbitControl(camOrtho, app.Window()))
	// Adds camera to scene (important for audio demos)
	app.Scene().Add(app.CameraOrtho().GetCamera())


    // Setup for perspective camera

	// centerX := app.sceneWidth / 2
	// centerY := app.sceneHeight / 2

	// // Sets perspective camera position
	// width, height := app.Window().Size()
	// aspect := float32(width) / float32(height)
	// app.CameraPersp().SetPosition(centerX, centerY, 1000)
	// app.CameraPersp().LookAt(&math32.Vector3{centerX, centerY, 0})
	// app.CameraPersp().SetAspect(aspect)

	// // Default camera is perspective
	// app.SetCamera(app.CameraPersp())
	// // Adds camera to scene (important for audio demos)
	// app.Scene().Add(app.Camera().GetCamera())


	// Subscribe to window key events
	app.Window().Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		// ESC terminates program
		if kev.Keycode == window.KeyEscape {
			app.Quit()
			return
		}
		// Alt F11 toggles full screen
		if kev.Keycode == window.KeyF11 && kev.Mods == window.ModAlt {
			app.Window().SetFullScreen(!app.Window().FullScreen())
			return
		}
		// Ctr-Alt-S prints statistics in the console
		if kev.Keycode == window.KeyS && kev.Mods == window.ModControl|window.ModAlt {
			app.logStats()
		}
	})

	// Subscribe to window resize events
	app.Window().Subscribe(window.OnWindowSize, func(evname string, ev interface{}) {
		app.OnWindowResize()
	})

	// Subscribe to mouse button down events
	app.Window().Subscribe(window.OnMouseDown, func(evname string, ev interface{}) {
		app.onMouse(ev)
	})

	// Because all windows events were cleared
	// We need to inform the gui root panel to subscribe again.
	app.Gui().SubscribeWin()

	// If no gui control folder, nothing more to do
	if app.control == nil {
		return
	}

	// Remove all controls and adds default ones
	app.control.Clear()
}

func (app *App) onMouse(ev interface{}) {
	// Convert mouse coordinates to normalized device coordinates
	// mev := ev.(*window.MouseEvent)
	// width, height := app.Window().Size()
	// x := 2*(mev.Xpos/float32(width)) - 1
	// y := -2*(mev.Ypos/float32(height)) + 1
    // app.log.Info("Xpos: %f Ypos: %f", x, y)

	// // Set the raycaster from the current camera and mouse coordinates
	// app.Camera().SetRaycaster(t.rc, x, y)
	// //fmt.Printf("rc:%+v\n", t.rc.Ray)

	// // Checks intersection with all objects in the scene
	// intersects := t.rc.IntersectObjects(app.Scene().Children(), true)
	// //fmt.Printf("intersects:%+v\n", intersects)

}

// GuiPanel returns the current gui panel for demos to add elements to.
func (app *App) GuiPanel() *gui.Panel {

	if *oNogui {
		return &app.Gui().Panel
	}
	return app.Panel3D().GetPanel()
}

// ControlFolder returns the application control folder
func (app *App) ControlFolder() *gui.ControlFolder {

	return app.control
}

// AmbLight returns the default scene ambient light
func (app *App) AmbLight() *light.Ambient {

	return app.ambLight
}

func (app *App) CurrentFixture() *fixture.Fixture {
    return app.fixtures[app.selected]
}

// UpdateFPS updates the fps value in the window title or header label
func (app *App) updateFPS() {

	if *oHideFPS {
		return
	}

	// Get the FPS and potential FPS from the frameRater
	fps, pfps, ok := app.FrameRater().FPS(time.Duration(*oUpdateFPS) * time.Millisecond)
	if !ok {
		return
	}

	// Shows the values in the window title or header label
	msg := fmt.Sprintf("%3.1f / %3.1f", fps, pfps)
	if *oNogui {
		app.Window().SetTitle(msg)
	} else {
		app.labelFPS.SetText(msg)
	}
}

// logStats generate log with current statistics
func (app *App) logStats() {

	const statsFormat = `
         Shaders: %d
            Vaos: %d
         Buffers: %d
        Textures: %d
  Uniforms/frame: %d
Draw calls/frame: %d
 CGO calls/frame: %d
`
	app.log.Info(statsFormat,
		app.stats.Glstats.Shaders,
		app.stats.Glstats.Vaos,
		app.stats.Glstats.Buffers,
		app.stats.Glstats.Textures,
		app.stats.Unisets,
		app.stats.Drawcalls,
		app.stats.Cgocalls,
	)
}

// usage shows the application usage
func usage() {

	fmt.Fprintf(os.Stderr, "%s v%d.%d\n", progName, vmajor, vminor)
	fmt.Fprintf(os.Stderr, "usage: %s [options] [<test>] \n", execName)
	flag.PrintDefaults()
	os.Exit(2)
}
