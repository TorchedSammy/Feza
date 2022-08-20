package main

import (
	"fmt"

	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib/packagelib"
	"github.com/tfriedel6/canvas"
)

var rendererLoader = packagelib.Loader{
	Name: "renderer",
	Load: rendererLoad,
}

func rendererLoad(rtm *rt.Runtime) (rt.Value, func()) {
	exports := map[string]luaExport{
		"begin_frame": {rendererBeginFrame, 0, false},
		"end_frame": {rendererEndFrame, 0, false},
		"draw_rect": {rendererDrawRect, 5, false},
		"draw_text": {rendererDrawText, 5, false},
	}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	r.SetEnv(mod, "font", rFontLoader(rtm))

	return rt.TableValue(mod), nil
}

func rendererBeginFrame(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	wnd.StartFrame()
	return c.Next(), nil
}

func rendererEndFrame(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	wnd.FinishFrame()
	return c.Next(), nil
}

func rendererDrawRect(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(5); err != nil {
		return nil, err
	}

	x, err := c.FloatArg(0)
	if err != nil {
		return nil, err
	}
	y, err := c.FloatArg(1)
	if err != nil {
		return nil, err
	}
	w, err := c.FloatArg(2)
	if err != nil {
		return nil, err
	}
	h, err := c.FloatArg(3)
	if err != nil {
		return nil, err
	}
	color, err := c.TableArg(4)
	if err != nil {
		return nil, err
	}

	r := int(color.Get(rt.IntValue(1)).AsInt())
	g := int(color.Get(rt.IntValue(2)).AsInt())
	b := int(color.Get(rt.IntValue(3)).AsInt())
	a := int(color.Get(rt.IntValue(4)).AsInt())

	cv.SetFillStyle(r, g, b, a)
	cv.FillRect(x, y, w, h)

	return c.Next(), nil
}

func rendererDrawText(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(5); err != nil {
		return nil, err
	}

	_, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	text, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}
	x, err := c.FloatArg(2)
	if err != nil {
		return nil, err
	}
	y, err := c.FloatArg(3)
	if err != nil {
		return nil, err
	}
	color, err := c.TableArg(4)
	if err != nil {
		return nil, err
	}

	r := int(color.Get(rt.IntValue(1)).AsInt())
	g := int(color.Get(rt.IntValue(2)).AsInt())
	b := int(color.Get(rt.IntValue(3)).AsInt())
	a := int(color.Get(rt.IntValue(4)).AsInt())

	f, _ := cv.LoadFont("/usr/share/fonts/dejavu-sans-fonts/DejaVuSans.ttf")
	cv.SetFont(f, 36)
	cv.SetFillStyle(r, g, b, a)
	cv.FillText(text, x, y)

	return c.Next(), nil
}

var fontMetaKey = rt.StringValue("_lavateraFont")
type font struct{
	f *canvas.Font
	size int
}

func rFontLoader(rtm *rt.Runtime) rt.Value {
	fontMethods := rt.NewTable()
	r.SetEnvGoFunc(fontMethods, "copy", rFontCopy, 1, true)

	fontMeta := rt.NewTable()
	r.SetEnv(fontMeta, "__index", rt.TableValue(fontMethods))
	r.SetRegistry(fontMetaKey, rt.TableValue(fontMeta))

	exports := map[string]luaExport{
		"load": {rFontLoad, 2, true},
	}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	return rt.TableValue(mod)
}

func fontArg(c *rt.GoCont, n int) (*font, error) {
	re, ok := valueToFont(c.Arg(n))
	if ok {
		return re, nil
	}
	return nil, fmt.Errorf("#%d must be a font", n+1)
}

func valueToFont(v rt.Value) (f *font, ok bool) {
	var u *rt.UserData
	u, ok = v.TryUserData()
	if ok {
		f, ok = u.Value().(*font)
	}
	return
}

func rFontLoad(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}

	path, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	size, err := c.IntArg(1)
	if err != nil {
		return nil, err
	}

	fnt, err := cv.LoadFont(path)
	if err != nil {
		return nil, err
	}
	f := &font{fnt, int(size)}

	fontMeta := t.Registry(fontMetaKey)
	return c.PushingNext1(t.Runtime, t.NewUserDataValue(f, fontMeta.AsTable())), nil
}

func rFontCopy(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	fnt, err := fontArg(c, 0)
	if err != nil {
		return nil, err
	}

	size := fnt.size
	if len(c.Etc()) != 0 {
		sz, ok := c.Etc()[0].TryInt()
		if !ok {
			return nil, fmt.Errorf("???")
		}
		size = int(sz)
	}

	f := &font{fnt.f, size}
	fontMeta := t.Registry(fontMetaKey)
	return c.PushingNext1(t.Runtime, t.NewUserDataValue(f, fontMeta.AsTable())), nil
}
