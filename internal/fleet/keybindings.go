package fleet

import (
	"github.com/gdamore/tcell/v2"
)

func (g *Gui) HandleKeybinding(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q':
		g.Stop()
		return nil
	}

	switch event.Key() {
	case tcell.KeyCtrlR:
		g.reload(true)
		return nil
	case tcell.KeyTab:
		g.nextPanel()
		return nil
	case tcell.KeyBacktab:
		g.prevPanel()
		return nil
	}

	return event
}
