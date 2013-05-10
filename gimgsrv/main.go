package main

import (
	"fmt"
	"github.com/ziutek/gdk"
	"github.com/ziutek/gtk"
)

// this happens when the X is clicked (to close the app)
func cbDelete(w *gtk.Widget, ev *gdk.Event) bool {
	fmt.Println("Delete")
	defer gtk.MainQuit()
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

	vb := gtk.NewVBox(false, 0)

	f := gtk.NewFileChooserButton("Select File", gtk.FILE_CHOOSER_ACTION_OPEN)
	vb.PackEnd(f.AsWidget(), false, false, 0)
	f.Show()

	a := Action{"Hello World!\n"}
	b := gtk.NewButtonWithLabel("Quit")
	b.Connect("clicked", (*Action).Quit, &a)
	b.ConnectNoi("clicked", (*gtk.Widget).Destroy, w.AsWidget())
	vb.PackEnd(b.AsWidget(), false, false, 0)
	b.Show()

	w.Add(vb.AsWidget())

	gtk.Main()
}

type Action struct {
	s string
}

func (a *Action) Quit(w *gtk.Widget) {
	fmt.Printf(a.s)
}
