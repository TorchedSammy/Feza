package main

import (
	"fmt"
	"regexp"
	"strings"

	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib/packagelib"
)

const (
	REGEX_ANCHORED = iota + 1 << 2
	REGEX_ENDANCHORED
)

var regexMetaKey = rt.StringValue("_lavateraRegexp")
var regexLoader = packagelib.Loader{
	Name: "regex",
	Load: regexLoad,
}

func regexLoad(rtm *rt.Runtime) (rt.Value, func()) {
	regexMethods := rt.NewTable()
	r.SetEnvGoFunc(regexMethods, "cmatch", regexMatch, 2, false)

	regexMeta := rt.NewTable()
	r.SetEnv(regexMeta, "__index", rt.TableValue(regexMethods))
	r.SetRegistry(regexMetaKey, rt.TableValue(regexMeta))

	exports := map[string]luaExport{
		"compile": {regexCompile, 1, true},
	}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	r.SetEnv(mod, "ANCHORED", rt.IntValue(REGEX_ANCHORED))
	r.SetEnv(mod, "ENDANCHORED", rt.IntValue(REGEX_ENDANCHORED))

	return rt.TableValue(mod), nil
}

func regexArg(c *rt.GoCont, n int) (*regexp.Regexp, error) {
	re, ok := valueToRegex(c.Arg(n))
	if ok {
		return re, nil
	}
	return nil, fmt.Errorf("#%d must be a regex", n+1)
}

func valueToRegex(v rt.Value) (re *regexp.Regexp, ok bool) {
	var u *rt.UserData
	u, ok = v.TryUserData()
	if ok {
		re, ok = u.Value().(*regexp.Regexp)
	}
	return
}

func regexCompile(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	pat, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	var flags string
	if len(c.Etc()) != 0 {
		options, ok := c.Etc()[0].TryString()
		if !ok {
			return nil, fmt.Errorf("??")
		}

		if strings.Contains(options, "i") {
			flags += "i"
		}
		if strings.Contains(options, "s") {
			flags += "s"
		}
		if strings.Contains(options, "m") {
			flags += "s"
		}
	}

	if flags != "" {
		pat = "(?" + flags + ")" + pat
	}

	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}

	regexMeta := t.Registry(regexMetaKey)
	return c.PushingNext1(t.Runtime, t.NewUserDataValue(re, regexMeta.AsTable())), nil
}

func regexMatch(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	re, err := regexArg(c, 0)
	if err != nil {
		return nil, err
	}

	subject, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	var indexes [][]int
	if len(c.Etc()) != 0 {
		offset, ok := c.Etc()[0].TryInt()
		if !ok {
			return nil, fmt.Errorf("??")
		}
		var opts int64
		if len(c.Etc()) > 1 {
			var ok bool
			opts, ok = c.Etc()[0].TryInt()
			if !ok {
				return nil, fmt.Errorf("??")
			}
		}

		match := true
		anchored := opts & REGEX_ANCHORED == REGEX_ANCHORED
		endAnchored := opts & REGEX_ENDANCHORED == REGEX_ENDANCHORED;

		ln := re.FindStringIndex(subject)
		if anchored {
			start := ln[0]
			if int64(start) + offset != offset {
				match = false
			}
		}

		if endAnchored {
			end := ln[1]
			if end != len(subject) {
				match = false
			}
		}

		if match {
			indexes = append(indexes, ln)
		}
	}

	luaIndexes := rt.NewTable()
	if len(indexes) == 0 && len(c.Etc()) == 0 {
		indexes = re.FindAllStringIndex(subject, -1)
	}

	for v, idxSlice := range indexes {
		luaIdxSlice := rt.NewTable()
		for vv, i := range idxSlice {
			luaIdxSlice.Set(rt.IntValue(int64(vv)), rt.IntValue(int64(i)))
		}

		luaIndexes.Set(rt.IntValue(int64(v)), rt.TableValue(luaIdxSlice))
	}

	return c.PushingNext1(t.Runtime, rt.TableValue(luaIndexes)), nil
}
