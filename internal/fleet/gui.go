package fleet

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type state struct {
	panels panels
}

func newState() *state {
	return &state{}
}

type Gui struct {
	app     *tview.Application
	state   *state
	manager *FleetManager
}

// New create new gui
func New(manager *FleetManager) *Gui {
	return &Gui{
		app:     tview.NewApplication(),
		state:   &state{},
		manager: manager,
	}
}

func (g *Gui) init() {
	projectTable := newProjectTable(g.app, g.manager)
	projectTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		event = g.HandleKeybinding(event)
		if event == nil {
			return nil
		}
		return projectTable.HandleKeybinding(event)
	})

	organizationTable := newOrganizationTable(g.app, g.manager)
	organizationTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		event = g.HandleKeybinding(event)
		if event == nil {
			return nil
		}
		return organizationTable.HandleKeybinding(event)
	})
	organizationTable.onChange = func(o *OrganizationInfo) {
		projectTable.setOrganization(o)
		g.app.SetFocus(projectTable)
	}

	g.state.panels.panel = append(g.state.panels.panel, organizationTable, projectTable)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(organizationTable, 0, 1, true).
		AddItem(projectTable, 0, 4, false)

	g.app.SetRoot(flex, true).EnableMouse(true)
}

func (g *Gui) startMonitoring() {
	for _, panel := range g.state.panels.panel {
		go panel.startMonitoring()
	}
}

func (g *Gui) stopMonitoring() {
	for _, panel := range g.state.panels.panel {
		go panel.stopMonitoring()
	}
}

func (g *Gui) Start() error {
	g.init()
	g.startMonitoring()
	if err := g.app.Run(); err != nil {
		g.app.Stop()

		return err
	}

	return nil
}

func (g *Gui) Stop() error {
	g.stopMonitoring()
	g.app.Stop()

	return nil
}
