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
		"set_clip_rect": {rendererClipRect, 4, false},
		"get_size": {rendererGetSize, 0, false},
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

func rendererClipRect(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(4); err != nil {
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

	cv.Rect(x, y, w, h)
	//cv.Clip()

	return c.Next(), nil
}

func rendererGetSize(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	w, h := wnd.Window.GetSize()

	return c.PushingNext(t.Runtime, rt.IntValue(int64(w)), rt.IntValue(int64(h))), nil
}

func rendererDrawText(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(5); err != nil {
		return nil, err
	}

	_, err := fontArg(c, 0)
	if err != nil {
		return nil, err
	}

	text, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}
	x, err := numOrIntArg(c, 2)
	if err != nil {
		return nil, err
	}
	y, err := numOrIntArg(c, 3)
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
	r.SetEnvGoFunc(fontMethods, "get_height", rFontHeight, 1, true)
	r.SetEnvGoFunc(fontMethods, "get_width", rFontWidth, 1, true)
	r.SetEnvGoFunc(fontMethods, "set_tab_size", luaStub, 1, true)

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
	size, err := c.FloatArg(1)
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
		sz, ok := c.Etc()[0].TryFloat()
		if !ok {
			return nil, fmt.Errorf("???")
		}
		size = int(sz)
		fmt.Println(size)
	}

	f := &font{fnt.f, size}
	fontMeta := t.Registry(fontMetaKey)
	return c.PushingNext1(t.Runtime, t.NewUserDataValue(f, fontMeta.AsTable())), nil
}

func rFontHeight(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	_, err := fontArg(c, 0)
	if err != nil {
		return nil, err
	}

	return c.PushingNext1(t.Runtime, rt.IntValue(20)), nil
}

func rFontWidth(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	fnt, err := fontArg(c, 0)
	if err != nil {
		return nil, err
	}

/*
	text, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}
*/
	text := "lol"
	cv.Save()
	cv.SetFont(fnt.f, float64(fnt.size))
	metrics := cv.MeasureText(text)
	cv.Restore()
	
	return c.PushingNext1(t.Runtime, rt.FloatValue(metrics.Width)), nil
}
