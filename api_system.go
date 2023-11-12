package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib/packagelib"
	"github.com/veandco/go-sdl2/sdl"
)

var systemLoader = packagelib.Loader{
	Name: "system",
	Load: systemLoad,
}

func systemLoad(rtm *rt.Runtime) (rt.Value, func()) {
	exports := map[string]luaExport{
		"poll_event": {systemPollEvent, 0, false},
		"wait_event": {systemWaitEvent, 1, false},
		"set_cursor": {systemSetCursor, 1, false}, // TODO
		"set_window_title": {systemSetTitle, 1, false},
		"set_window_mode": {systemSetWindowMode, 1, false}, // TODO
		"get_window_mode": {systemGetWindowMode, 0, false},
		"set_window_bordered": {systemSetBordered, 1, false}, // TODO
		"set_window_hit_test": {systemWindowHitTest, 1, false}, // TODO
		"get_window_size": {systemGetWindowSize, 0, false},
		"set_window_size": {systemSetWindowSize, 4, false},
		"window_has_focus": {systemWindowHasFocus, 0, false},
		"show_fatal_error": {systemShowFatalError, 2, false},
		"chdir": {systemChdir, 1, false},
		"mkdir": {systemMkdir, 1, false},
		"list_dir": {systemListDir, 1, false},
		"absolute_path": {systemAbsolutePath, 1, false},
		"get_file_info": {systemGetFileInfo, 1, false},
		// get_clipboard
		// set_clipboard
		// get_process_id
		"get_time": {systemGetTime, 0, false},
		"sleep": {systemSleep, 1, false},
		// exec (don't return, non blocking)
		// fuzzy_match
		"set_window_opacity": {systemSetWindowOpacity, 1, false},
		// these below aren't documented in docs/system.lua of lite xl upstream
		"get_fs_type": {systemFsType, 1, false},
		// rmdir
		// path_compare
		"path_compare": {systemPathCompare, 4, false},
		// load_native_plugn - we can't actually implement for (hopefully obvious) reasons; we're in go land
	}
	makeStubs([]string{
		"text_input",
	}, exports)
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

func makeStubs(names []string, exports map[string]luaExport) {
	for _, name := range names {
		exports[name] = luaExport{luaStub, 0, true}
	}
}

func luaStub(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	return c.Next(), nil
}

func itv(num int64) rt.Value {
	return rt.IntValue(num)
}

func stv(str string) rt.Value {
	return rt.StringValue(str)
}

var i = 0
func systemPollEvent(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	n := c.Next()
	fmt.Println("poll called. waiting")
poll:
	event := sdl.PollEvent()
	fmt.Println("past poll call", i)
	i++

	switch e := event.(type) {
		case *sdl.QuitEvent:
			n.Push(t.Runtime, stv("quit"))
		// Window Events
		case *sdl.WindowEvent:
			switch e.Event {
				case sdl.WINDOWEVENT_EXPOSED: n.Push(t.Runtime, stv("exposed"))
				case sdl.WINDOWEVENT_RESTORED: n.Push(t.Runtime, stv("restored"))
				case sdl.WINDOWEVENT_RESIZED:
					w := e.Data1
					h := e.Data2
					n.Push(t.Runtime, stv("resized"))
					n.Push(t.Runtime, itv(int64(w)))
					n.Push(t.Runtime, itv(int64(h)))
				default:
					goto poll
			}
		// Mouse Events
		case *sdl.MouseButtonEvent:
			typ := "mousereleased"
			if e.Type == sdl.MOUSEBUTTONDOWN {
				typ = "mousepressed"
			}
			n.Push(t.Runtime, stv(typ))
	
			var buttonName string
			switch e.Button {
				case sdl.BUTTON_LEFT: buttonName = "left"
				case sdl.BUTTON_MIDDLE: buttonName = "middle"
				case sdl.BUTTON_RIGHT: buttonName = "right"
				case sdl.BUTTON_X1: buttonName = "x"
				case sdl.BUTTON_X2: buttonName = "y"
			}
			n.Push(t.Runtime, stv(buttonName))
			n.Push(t.Runtime, itv(int64(e.X)))
			n.Push(t.Runtime, itv(int64(e.Y)))
			n.Push(t.Runtime, itv(int64(e.Clicks)))
		case *sdl.MouseMotionEvent:
			n.Push(t.Runtime, stv("mousemoved"))
			n.Push(t.Runtime, itv(int64(e.X)))
			n.Push(t.Runtime, itv(int64(e.Y)))
			n.Push(t.Runtime, itv(int64(e.XRel)))
			n.Push(t.Runtime, itv(int64(e.YRel)))
		case *sdl.MouseWheelEvent:
			n.Push(t.Runtime, stv("mousewheel"))
			n.Push(t.Runtime, itv(int64(e.Y)))
		default:
			goto poll
	}

	return n, nil
}

func systemWaitEvent(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	wait := func(t int) sdl.Event {
		return sdl.WaitEvent()
	}

	var timeout float64
	if err := c.Check1Arg(); err == nil {
		fmt.Println("not indefinite wait")
		wait = sdl.WaitEventTimeout
		timeout, err = c.FloatArg(0)
		if err != nil {
			return nil, err
		}
	}

	ev := wait(int(time.Duration(timeout) * time.Second))
	return c.PushingNext1(t.Runtime, rt.BoolValue(ev != nil)), nil
}

func systemSetCursor(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	// TODO: set passed cursor
	curs := sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_ARROW)
	sdl.SetCursor(curs)
	return c.Next(), nil
}

func systemSetTitle(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	title, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	wnd.Window.SetTitle(title)
	return c.Next(), nil
}

func systemSetWindowMode(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	// TODO: set window mode
	return c.Next(), nil
}

func systemGetWindowMode(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	flags := wnd.Window.GetFlags()
	var mode string
	switch {
		case flags & sdl.WINDOW_FULLSCREEN_DESKTOP == 0:
			mode = "fullscreen"
		case flags & sdl.WINDOW_MINIMIZED == 0:
			mode = "minimized"
		case flags & sdl.WINDOW_MAXIMIZED == 0:
			mode = "maximized"
		default:
			mode = "normal"
	}

	return c.PushingNext1(t.Runtime, rt.StringValue(mode)), nil
}

func systemSetBordered(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	_, err := c.BoolArg(0)
	if err != nil {
		return nil, err
	}

	// TODO: set (non) bordered

	return c.Next(), nil
}

func systemWindowHitTest(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	// TODO
	return c.Next(), nil
}

func systemGetWindowSize(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	w, h := wnd.Window.GetSize()
	x, y := wnd.Window.GetPosition()

	return c.PushingNext(t.Runtime, rt.IntValue(int64(w)), rt.IntValue(int64(h)), rt.IntValue(int64(x)), rt.IntValue(int64(y))), nil
}

func systemSetWindowSize(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	w, err := c.IntArg(0)
	if err != nil {
		return nil, err
	}
	h, err := c.IntArg(1)
	if err != nil {
		return nil, err
	}
	x, err := c.IntArg(2)
	if err != nil {
		return nil, err
	}
	y, err := c.IntArg(3)
	if err != nil {
		return nil, err
	}

	wnd.Window.SetSize(int32(w), int32(h))
	wnd.Window.SetPosition(int32(x), int32(y))

	return c.Next(), nil
}

func systemWindowHasFocus(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	// TODO: get focus status
	focus := true

	return c.PushingNext1(t.Runtime, rt.BoolValue(focus)), nil
}

func systemShowFatalError(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}

	title, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	message, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	// really: what are we gonna do when we move to just gl ...
	if err := sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_ERROR, title, message, wnd.Window); err != nil {
		return nil, err
	}

	return c.Next(), nil
}

func systemChdir(t *rt.Thread, c *rt.GoCont) (rt.Cont, error){
	path, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	if err := os.Chdir(path); err != nil {
		return nil, err
	}

	return c.Next(), nil
}

func systemMkdir(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(1); err != nil {
		return nil, err
	}
	path, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	err = os.Mkdir(path, 0744)
	if err != nil {
		return c.PushingNext1(t.Runtime, rt.BoolValue(false)), err
	}

	return c.PushingNext1(t.Runtime, rt.BoolValue(true)), err
}

func systemListDir(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	dir, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	names := rt.NewTable()

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return c.PushingNext(t.Runtime, rt.NilValue, rt.StringValue(err.Error())), nil
	}
	for i, entry := range dirEntries {
		names.Set(rt.IntValue(int64(i + 1)), rt.StringValue(entry.Name()))
	}

	return c.PushingNext1(t.Runtime, rt.TableValue(names)), nil
}

func systemAbsolutePath(t *rt.Thread, c *rt.GoCont) (rt.Cont, error){
	path, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	return c.PushingNext1(t.Runtime, rt.StringValue(abspath)), nil
}

func systemGetFileInfo(t *rt.Thread, c *rt.GoCont) (rt.Cont, error){
	path, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	pathinfo, err := os.Stat(path)
	if err != nil {
		return c.PushingNext1(t.Runtime, rt.StringValue(err.Error())), nil
	}
	statTbl := rt.NewTable()
	if pathinfo.IsDir() {
		r.SetEnv(statTbl, "type", rt.StringValue("dir"))
	} else {
		r.SetEnv(statTbl, "type", rt.StringValue("file"))
	}
	r.SetEnv(statTbl, "modified", rt.IntValue(int64(pathinfo.ModTime().Second())))
	r.SetEnv(statTbl, "size", rt.IntValue(int64(pathinfo.Size())))

	return c.PushingNext1(t.Runtime, rt.TableValue(statTbl)), nil
}

// get_clipboard

// set_clipboard

// get_process_id

func systemGetTime(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	// "Get amount of iterations since the application was launched, also known as SDL_GetPerformanceCounter() / SDL_GetPerformanceFrequency()"
	// what is this going to be when we use just gl ...
	time := sdl.GetPerformanceCounter() / sdl.GetPerformanceFrequency()

	return c.PushingNext1(t.Runtime, rt.IntValue(int64(time))), nil
}

func systemSleep(r *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	secs, err := c.IntArg(0)
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Duration(secs) * time.Second)
	return c.Next(), nil
}

// exec

// fuzzy_match

// set_window_opacity
func systemSetWindowOpacity(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	opacity, err := c.FloatArg(0)
	if err != nil {
		return nil, err
	}

	wnd.Window.SetWindowOpacity(float32(opacity))

	return c.Next(), nil
}

func systemFsType(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	_, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	fsType := "ext4" // TODO: get actual fs type

	return c.PushingNext1(t.Runtime, rt.StringValue(fsType)), nil
}

// rmdir

// path_compare
func systemPathCompare(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	return c.PushingNext1(t.Runtime, rt.BoolValue(true)), nil
}

// load_native_plugin
