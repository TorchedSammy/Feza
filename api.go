package main

import (
	"os"
	"os/user"
	"runtime"

	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib"
)

type luaExport struct {
	Function rt.GoFunctionFunc
	ArgNum int
	Variadic bool
}

func setExports(rtm *rt.Runtime, tbl *rt.Table, exports map[string]luaExport) {
	for name, export := range exports {
		rtm.SetEnvGoFunc(tbl, name, export.Function, export.ArgNum, export.Variadic)
	}
}

func setupAPI() {
	// Core native libraries
	lib.LoadLibs(r, rendererLoader)
	lib.LoadLibs(r, systemLoader)
	lib.LoadLibs(r, processLoader)
	lib.LoadLibs(r, regexLoader)
	lib.LoadLibs(r, dirmonitorLoader)

	env := r.GlobalEnv()
	renderer, _ := rendererLoader.Load(r)
	system, _ := systemLoader.Load(r)
	process, _ := processLoader.Load(r)
	regex, _ := regexLoader.Load(r)
	dirmon, _ := dirmonitorLoader.Load(r) // type is called dirmonitor. doesnt matter in code, but /shrug

	r.SetEnv(env, "renderer", renderer)
	r.SetEnv(env, "system", system)
	r.SetEnv(env, "process", process)
	r.SetEnv(env, "regex", regex)
	r.SetEnv(env, "dirmonitor", dirmon)

	// Globals
	var platform string
	switch runtime.GOOS {
		case "windows": platform = "Windows"
		case "darwin": platform = "Mac OS X"
		case "linux": platform = "Linux"
		// TODO: ios and android (will lavatera even build on those platforms..?)
	}
	exe, _ := os.Executable()
	curuser, _ := user.Current()
	homedir := curuser.HomeDir

	r.SetEnv(env, "ARGS", rt.TableValue(rt.NewTable())) // TODO: convert os.Args to lua
	r.SetEnv(env, "PLATFORM", rt.StringValue(platform))
	r.SetEnv(env, "SCALE", rt.IntValue(1)) // TODO: get dpi
	r.SetEnv(env, "EXEFILE", rt.StringValue(exe))
	r.SetEnv(env, "HOME", rt.StringValue(homedir))
}
