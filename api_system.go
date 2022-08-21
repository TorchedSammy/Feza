package main

import (
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
		"absolute_path": {systemAbsolutePath, 1, false},
		"get_time": {systemGetTime, 0, false},
		"get_file_info": {systemGetFileInfo, 1, false},
		"show_fatal_error": {systemShowFatalError, 2, false},
		"chdir": {systemChdir, 1, false},
		"list_dir": {systemListDir, 1, false},
		"mkdir": {systemMkdir, 1, false},
		"get_fs_type": {systemFsType, 1, false},
		"set_window_bordered": {systemSetBordered, 1, false},
		"set_window_hit_test": {systemWindowHitTest, 1, false},
		"set_window_title": {systemSetTitle, 1, false},
		"sleep": {systemSleep, 1, false},
	}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

func systemPollEvent(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	event := sdl.PollEvent()

	var eventName string
	var eventArgs []interface{}
	switch e := event.(type) {
		case *sdl.QuitEvent:
			eventName = "quit"
		case *sdl.MouseButtonEvent:
			eventName = "mousereleased"
			if e.Type == sdl.MOUSEBUTTONDOWN {
				eventName = "mousepressed"
			}
			var buttonName string
			switch e.Button {
				case sdl.BUTTON_LEFT: buttonName = "left"
				case sdl.BUTTON_MIDDLE: buttonName = "middle"
				case sdl.BUTTON_RIGHT: buttonName = "right"
				case sdl.BUTTON_X1: buttonName = "x"
				case sdl.BUTTON_X2: buttonName = "y"
			}
			eventArgs = []interface{}{buttonName, int(e.X), int(e.Y), int(e.Clicks)}
		case *sdl.MouseMotionEvent:
			eventName = "mousemoved"
			eventArgs = []interface{}{int(e.X), int(e.Y), int(e.XRel), int(e.YRel)}
		case *sdl.MouseWheelEvent:
			eventName = "mousewheel"
			eventArgs = []interface{}{int(e.Y)}
	}

	luaEventArgs := rt.NewTable()
	for i, arg := range eventArgs {
		if num, ok := arg.(int); ok {
			luaEventArgs.Set(rt.IntValue(int64(i + 1)), rt.IntValue(int64(num)))
			continue
		}
		luaEventArgs.Set(rt.IntValue(int64(i + 1)), rt.AsValue(arg))
	}

	return c.PushingNext(t.Runtime, rt.StringValue(eventName), rt.TableValue(luaEventArgs)), nil
}

func systemGetTime(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	// "Get amount of iterations since the application was launched, also known as SDL_GetPerformanceCounter() / SDL_GetPerformanceFrequency()"
	// what is this going to be when we use just gl ...
	time := sdl.GetPerformanceCounter() / sdl.GetPerformanceFrequency()

	return c.PushingNext1(t.Runtime, rt.IntValue(int64(time))), nil
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

func systemFsType(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	_, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	fsType := "ext4" // TODO: get actual fs type

	return c.PushingNext1(t.Runtime, rt.StringValue(fsType)), nil
}

func systemWindowHitTest(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	// TODO
	return c.Next(), nil
}

func systemSetBordered(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	_, err := c.BoolArg(0)
	if err != nil {
		return nil, err
	}

	// TODO: set (non) bordered

	return c.Next(), nil
}

func systemSleep(r *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	secs, err := c.IntArg(0)
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Duration(secs) * time.Second)
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
