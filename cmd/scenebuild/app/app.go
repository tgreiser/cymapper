package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/application"
	"github.com/g3n/engine/util/logger"
	"github.com/g3n/engine/util/stats"
	"github.com/g3n/engine/window"
	"github.com/kardianos/osext"
)

type App struct {
	*application.Application                    // Embedded standard application object
	log                      *logger.Logger     // Application logger
	currentScreen            IScreen            // Current screen being rendered
	dirData                  string             // full path of data directory
	labelFPS                 *gui.Label         // header FPS label
	stats                    *stats.Stats       // statistics object
	statsTable               *stats.StatsTable  // statistics table panel
	control                  *gui.ControlFolder // Pointer to gui control panel
	fs                       *FileSelect        // File select dialog
	ed                       *ErrorDialog       // Error dialog
	ambLight                 *light.Ambient
	finalizers               []func() // List of demo finalizers functions
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
	vmajor   = 0
	vminor   = 1
)

func Create() *App {
	flag.Usage = usage

	// Creates standard application object
	a, err := application.Create(application.Options{
		Title:       progName,
		Width:       800,
		Height:      600,
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

	// Check for data directory and aborts if not found
	app.dirData = app.checkDirData("data")
	app.log.Info("Using data directory:%s", app.dirData)

	// Builds user interface
	if *oNogui == false {
		app.buildGui()
	}

	// Setup scene
	app.setupScene()

	// Subscribe to before render events to call current test Render method
	app.Subscribe(application.OnBeforeRender, func(evname string, ev interface{}) {
		if app.currentScreen != nil {
			app.currentScreen.Render(app)
		}
	})

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
	// Execute demo finalizers functions and clear finalizers list
	for i := 0; i < len(app.finalizers); i++ {
		app.finalizers[i]()
	}
	app.finalizers = app.finalizers[0:0]

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
	app.CameraOrtho().SetPosition(320, 240, 99)
	app.CameraOrtho().LookAt(&math32.Vector3{320, 240, 0})
	app.CameraOrtho().SetZoom(0.005)

	// Default camera is perspective
	app.SetCamera(app.CameraOrtho())
	//app.SetOrbit(control.NewOrbitControl(camOrtho, app.Window()))
	// Adds camera to scene (important for audio demos)
	app.Scene().Add(app.CameraOrtho().GetCamera())

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

// GuiPanel returns the current gui panel for demos to add elements to.
func (app *App) GuiPanel() *gui.Panel {

	if *oNogui {
		return &app.Gui().Panel
	}
	return app.Panel3D().GetPanel()
}

// DirData returns the base directory for data
func (app *App) DirData() string {

	return app.dirData
}

// ControlFolder returns the application control folder
func (app *App) ControlFolder() *gui.ControlFolder {

	return app.control
}

// AmbLight returns the default scene ambient light
func (app *App) AmbLight() *light.Ambient {

	return app.ambLight
}

// AddFinalizer adds a function which will be executed before another demo is started
func (app *App) AddFinalizer(f func()) {

	app.finalizers = append(app.finalizers, f)
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

// checkDirData try to find and return the complete data directory path.
// Aborts if not found
func (app *App) checkDirData(dirDataName string) string {

	// Checks first if data directory is in the current directory
	if _, err := os.Stat(dirDataName); err == nil {
		dirData, err := filepath.Abs(dirDataName)
		if err != nil {
			panic(err)
		}
		return dirData
	}

	// Get the executable path
	execPath, err := osext.Executable()
	if err != nil {
		panic(err)
	}

	// Checks if data directory is in the executable directory
	execDir := filepath.Dir(execPath)
	path := filepath.Join(execDir, dirDataName)
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Assumes the executable is in $GOPATH/bin
	goPath := filepath.Dir(execDir)
	path = filepath.Join(goPath, "src", "github.com", "g3n", "g3nd", dirDataName)
	// Checks data path
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Shows error message and aborts
	app.log.Fatal("Data directory NOT FOUND")
	return ""
}

// usage shows the application usage
func usage() {

	fmt.Fprintf(os.Stderr, "%s v%d.%d\n", progName, vmajor, vminor)
	fmt.Fprintf(os.Stderr, "usage: %s [options] [<test>] \n", execName)
	flag.PrintDefaults()
	os.Exit(2)
}
