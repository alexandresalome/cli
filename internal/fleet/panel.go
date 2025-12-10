package fleet

import "github.com/rivo/tview"

type panel interface {
	tview.Primitive
	reload()
}

type panels struct {
	panel []panel
}

func (g *Gui) reload() {
	for _, panel := range g.state.panels.panel {
		panel.reload()
	}
}

func (g *Gui) currentFocus() int {
	item := g.app.GetFocus()
	x, y, width, height := item.GetRect()
	for i, panel := range g.state.panels.panel {
		px, py, pwidth, pheight := panel.GetRect()
		if x >= px && y >= py && x+width <= px+pwidth && y+height <= py+pheight {

			return i
		}
	}

	return 0
}

func (g *Gui) nextPanel() {
	current := g.currentFocus()
	next := (current + 1) % len(g.state.panels.panel)

	g.app.SetFocus(g.state.panels.panel[next])
}

func (g *Gui) prevPanel() {
	current := g.currentFocus()
	next := current - 1

	if next < 0 {
		next = len(g.state.panels.panel) - 1
	}

	g.app.SetFocus(g.state.panels.panel[next])
}
