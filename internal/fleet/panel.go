package fleet

import "github.com/rivo/tview"

type panel interface {
	tview.Primitive
	name() string
	reload(force bool)
}

type panels struct {
	currentPanel int
	panel        []panel
}

func (g *Gui) reload(force bool) {
	for _, panel := range g.state.panels.panel {
		panel.reload(force)
	}
}

func (g *Gui) nextPanel() {
	idx := (g.state.panels.currentPanel + 1) % len(g.state.panels.panel)
	g.switchPanel(g.state.panels.panel[idx].name())
}

func (g *Gui) prevPanel() {
	g.state.panels.currentPanel--

	if g.state.panels.currentPanel < 0 {
		g.state.panels.currentPanel = len(g.state.panels.panel) - 1
	}

	idx := (g.state.panels.currentPanel) % len(g.state.panels.panel)
	g.switchPanel(g.state.panels.panel[idx].name())
}

func (g *Gui) switchPanel(panelName string) {
	for i, panel := range g.state.panels.panel {
		if panel.name() == panelName {
			g.app.SetFocus(panel)
			g.state.panels.currentPanel = i
		}
	}
}

func (g *Gui) closeAndSwitchPanel(removePanel, switchPanel string) {
	//g.pages.RemovePage(removePanel).ShowPage("main")
	g.switchPanel(switchPanel)
}
