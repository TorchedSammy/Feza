package main

import (
	"log"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/sdlcanvas"
	rt "github.com/arnodel/golua/runtime"
)

var r *rt.Runtime
var wnd *sdlcanvas.Window
var cv *canvas.Canvas

func main() {
	var err error
	w, h := 1280, 720
	wnd, cv, err = sdlcanvas.CreateWindow(w, h, "")
	if err != nil {
		log.Println(err)
		return
	}
	defer wnd.Destroy()

	initLua()
}

