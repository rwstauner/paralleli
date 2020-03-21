package main

import (
	"fmt"
	"time"

	"github.com/awesome-gocui/gocui"
)

// Manager uses a Commander to build the gui layout.
type Manager struct {
	commander *Commander
	started   bool
}

// Layout gets called for updates to the gui.
func (m *Manager) Layout(g *gocui.Gui) error {
	if !m.started {
		m.started = true
		go func() {
			// TODO: make this configurable
			for range time.Tick(250 * time.Millisecond) {
				// TODO: stop when all commands are done
				// Signal a redraw, the layout func will do it.
				g.Update(func(g *gocui.Gui) error { return nil })
			}
		}()
	}

	// total := len(m.commander.Commands)
	for i, cmd := range m.commander.Commands {
		viewName := "x-" + cmd.Tag
		width, _ := g.Size()
		height := 5 // TODO: maxHeight / total, min of 5 (arg)
		top := i * (height + 2)

		// FIXME: why not width-2 ?
		v, err := g.SetView(viewName, 0, top, width-1, top+6, 0)
		if err != nil {
			if !gocui.IsUnknownView(err) {
				return err
			}

			v.Autoscroll = true
			v.Frame = true
			v.Title = cmd.Tag
		}

		decor := "running"
		if cmd.ExitCode > -1 {
			decor = fmt.Sprintf("exited %d", cmd.ExitCode)
		}
		v.Title = fmt.Sprintf("%s| %s ", decor, cmd.Tag)
		v.Clear()
		fmt.Fprintf(v, "%s", cmd.Bytes)

		if _, err := g.SetCurrentView(viewName); err != nil {
			return err
		}
	}

	return nil
}
