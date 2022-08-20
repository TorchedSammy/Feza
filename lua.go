package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"runtime"

	"github.com/arnodel/golua/lib"
	"github.com/arnodel/golua/lib/debuglib"
	rt "github.com/arnodel/golua/runtime"
)

func initLua() {
	r = rt.New(os.Stdout)
	r.PushContext(rt.RuntimeContextDef{
		MessageHandler: debuglib.Traceback,
	})
	lib.LoadAll(r)

	// Core native libraries
	lib.LoadLibs(r, rendererLoader)
	lib.LoadLibs(r, systemLoader)
	lib.LoadLibs(r, processLoader)
	lib.LoadLibs(r, regexLoader)

	env := r.GlobalEnv()
	renderer, _ := rendererLoader.Load(r)
	system, _ := systemLoader.Load(r)
	process, _ := processLoader.Load(r)
	regex, _ := regexLoader.Load(r)

	r.SetEnv(env, "renderer", renderer)
	r.SetEnv(env, "system", system)
	r.SetEnv(env, "process", process)
	r.SetEnv(env, "regex", regex)

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

	err := doFile(r, "init.lua")
	fmt.Println(err)
}

func doFile(rtm *rt.Runtime, path string) error {
	f, err := os.Open(path)
	defer f.Close()

	if err != nil {
		return err
	}
	
	reader := bufio.NewReader(f)
	c, err := reader.ReadByte()
	if err != nil && err != io.EOF {
		return err
	}

	// unread so a char won't be missing
	err = reader.UnreadByte()
	if err != nil {
		return err
	}

	var buf []byte
	if c == byte('#') {
		// shebang - skip that line
		_, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		buf = []byte{'\n'}
	}

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		
		buf = append(buf, line...)
	}

	clos, err := rtm.LoadFromSourceOrCode(path, buf, "bt", rt.TableValue(rtm.GlobalEnv()), false)
	if clos != nil {
		_, err = rt.Call1(rtm.MainThread(), rt.FunctionValue(clos))
	}

	return err
}
