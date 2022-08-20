package main

import (
	"os"
	"path/filepath"

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
		//"poll_event": {systemPollEvent, 0, false},
		"absolute_path": {systemAbsolutePath, 1, false},
		"get_time": {systemGetTime, 0, false},
		"get_file_info": {systemGetFileInfo, 1, false},
	}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

func systemPollEvent(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	event := sdl.PollEvent()

	switch event.(type) {
		//case 
	}
	return c.Next(), nil
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
	r.SetEnv(statTbl, "modified", rt.IntValue(int64(pathinfo.Size())))

	return c.PushingNext1(t.Runtime, rt.TableValue(statTbl)), nil
}
