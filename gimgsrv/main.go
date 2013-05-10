package main

import (
	"fmt"
	"github.com/ziutek/gdk"
	"github.com/ziutek/gtk"
)

func cbDelete(w *gtk.Widget, ev *gdk.Event) bool {
	fmt.Println("Delete")
	return true
}

func cbDestroy(w *gtk.Widget) {
	fmt.Println("Destroy")
	gtk.MainQuit()
}

func main() {
	w := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	w.Connect("delete-event", cbDelete, nil)
	w.Connect("destroy", cbDestroy, nil)
	w.SetBorderWidth(10)
	w.Show()

	a := A{"Hello World!\n"}

	b := gtk.NewButtonWithLabel("Hello World")
	b.Connect("clicked", (*A).Hello, &a)
	b.ConnectNoi("clicked", (*gtk.Widget).Destroy, w.AsWidget())
	w.Add(b.AsWidget())
	b.Show()

	gtk.Main()
}

type A struct {
	s string
}

func (a *A) Hello(w *gtk.Widget) {
	fmt.Printf(a.s)
}
