package main

import (
	"os"
	"syscall"

	"github.com/awesome-gocui/gocui"
)

func main() {
	if err := frd(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

func frd(args []string) error {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return err
	}
	defer g.Close()
	g.Highlight = true
	g.Cursor = true
	g.Mouse = true
	cmder := NewCommanderFromArgs(args)
	cmder.Start()
	man := &Manager{
		commander: cmder,
	}
	g.SetManager(man)

	ctrlc := func(g *gocui.Gui, v *gocui.View) error {
		// Wait for cleanup.
		cmder.Signal(syscall.SIGTERM)
		go func() {
			cmder.Wait()
			g.Update(quit)
		}()
		return nil
	}

	// Ctrl-C to quit.
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, ctrlc); err != nil {
		return err
	}

	// TODO: a keys to clear gui and print command output more concisely (with --tag if requested).

	if err := g.MainLoop(); err != nil && !gocui.IsQuit(err) {
		return err
	}

	return nil
}

func quit(g *gocui.Gui) error {
	return gocui.ErrQuit
}
