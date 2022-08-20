package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/arnodel/golua/lib"
	"github.com/arnodel/golua/lib/debuglib"
	rt "github.com/arnodel/golua/runtime"
)

func initLua() {
	init:
	r = rt.New(os.Stdout)
	r.PushContext(rt.RuntimeContextDef{
		MessageHandler: debuglib.Traceback,
	})
	lib.LoadAll(r)

	setupAPI()

	fn, err := doFile(r, "init.lua")
	fmt.Println(err)
	if b, _ := fn.TryBool(); b {
		// restart
		goto init
	}
}

func doFile(rtm *rt.Runtime, path string) (rt.Value, error) {
	f, err := os.Open(path)
	defer f.Close()

	v := rt.NilValue
	if err != nil {
		return v, err
	}
	
	reader := bufio.NewReader(f)
	c, err := reader.ReadByte()
	if err != nil && err != io.EOF {
		return v, err
	}

	// unread so a char won't be missing
	err = reader.UnreadByte()
	if err != nil {
		return v, err
	}

	var buf []byte
	if c == byte('#') {
		// shebang - skip that line
		_, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return v, err
		}
		buf = []byte{'\n'}
	}

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return v, err
		}
		
		buf = append(buf, line...)
	}

	clos, err := rtm.LoadFromSourceOrCode(path, buf, "bt", rt.TableValue(rtm.GlobalEnv()), false)
	if clos != nil {
		v, err = rt.Call1(rtm.MainThread(), rt.FunctionValue(clos))
	}

	return v, err
}
