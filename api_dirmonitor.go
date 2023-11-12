package main

import (
	"fmt"

	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib/packagelib"
	"github.com/fsnotify/fsnotify"
)

type dirmonitor struct{
	watcher *fsnotify.Watcher
	stop chan struct{}
	callback *rt.Closure
}

var dirmonitorMetaKey = rt.StringValue("_lavateraDirmonitor")
var dirmonitorLoader = packagelib.Loader{
	Name: "dirmonitor",
	Load: dirmonitorLoad,
}

func dirmonitorLoad(rtm *rt.Runtime) (rt.Value, func()) {
	dirmonitorMethods := rt.NewTable()
	r.SetEnvGoFunc(dirmonitorMethods, "watch", dirmonitorWatch, 2, false)
	r.SetEnvGoFunc(dirmonitorMethods, "unwatch", dirmonitorWatch, 2, false)
	r.SetEnvGoFunc(dirmonitorMethods, "check", dirmonitorWatch, 2, false)
	r.SetEnvGoFunc(dirmonitorMethods, "mode", dirmonitorMode, 1, false)

	dirmonitorMeta := rt.NewTable()
	r.SetEnv(dirmonitorMeta, "__index", rt.TableValue(dirmonitorMethods))
	r.SetRegistry(dirmonitorMetaKey, rt.TableValue(dirmonitorMeta))

	exports := map[string]luaExport{
		"new": {dirmonitorNew, 1, false},
	}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

func dirmonitorArg(c *rt.GoCont, n int) (*dirmonitor, error) {
	re, ok := valueToDirmonitor(c.Arg(n))
	if ok {
		return re, nil
	}
	return nil, fmt.Errorf("#%d must be a dirmonitor", n+1)
}

func valueToDirmonitor(v rt.Value) (re *dirmonitor, ok bool) {
	var u *rt.UserData
	u, ok = v.TryUserData()
	if ok {
		re, ok = u.Value().(*dirmonitor)
	}
	return
}

func dirmonitorNew(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	stop := make(chan struct{})
	monitor := &dirmonitor{watcher, stop, nil}

	go func(d *dirmonitor, sc chan struct{}) {
		select {
			case ev, ok := <-d.watcher.Events:
				if !ok {
					return
				}
				if d.callback != nil {
					rt.Call1(t.Runtime.MainThread(), rt.FunctionValue(d.callback), rt.StringValue(ev.Name))
				}
			case <-sc:
				return
		}
	}(monitor, stop)

	dirmonitorMeta := t.Registry(dirmonitorMetaKey)
	return c.PushingNext1(t.Runtime, t.NewUserDataValue(monitor, dirmonitorMeta.AsTable())), nil
}

func dirmonitorWatch(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	w, err := dirmonitorArg(c, 0)
	if err != nil {
		return nil, err
	}

	dir, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	err = w.watcher.Add(dir)
	if err != nil {
		return c.PushingNext1(t.Runtime, rt.IntValue(-1)), nil
	}

	return c.PushingNext1(t.Runtime, rt.IntValue(int64(len(w.watcher.WatchList())))), nil
}

func dirmonitorUnwatch(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	w, err := dirmonitorArg(c, 0)
	if err != nil {
		return nil, err
	}

	dir, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	err = w.watcher.Remove(dir)
	if err != nil {
		return nil, err
	}

	return c.Next(), nil
}

func dirmonitorCheck(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	w, err := dirmonitorArg(c, 0)
	if err != nil {
		return nil, err
	}

	fun, err := c.ClosureArg(1)
	if err != nil {
		return nil, err
	}

	w.callback = fun

	return c.Next(), nil
}

func dirmonitorMode(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	return c.PushingNext1(t.Runtime, rt.BoolValue(true)), nil
}
