package fleet

import (
	"github.com/gdamore/tcell/v2"
)

func (g *Gui) SetGlobalKeybinding(event *tcell.EventKey) {
	switch event.Rune() {
	case 'q':
		g.Stop()
	}

	switch event.Key() {
	case tcell.KeyTab:
		g.nextPanel()
	case tcell.KeyBacktab:
		g.prevPanel()
	case tcell.KeyRight:
		g.nextPanel()
	case tcell.KeyLeft:
		g.prevPanel()
	}
}
