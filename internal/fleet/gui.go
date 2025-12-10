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
	infoBox := newInfoBox(g.app, g.manager)
	infoBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		event = g.HandleKeybinding(event)
		if event == nil {
			return nil
		}
		return infoBox.HandleKeybinding(event)
	})

	environmentTable := newEnvironmentTable(g.app, g.manager)
	environmentTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		event = g.HandleKeybinding(event)
		if event == nil {
			return nil
		}
		return environmentTable.HandleKeybinding(event)
	})
	environmentTable.onChange = func(e *Environment) {
		infoBox.setEnvironment(e)
	}

	projectTable := newProjectTable(g.app, g.manager)
	projectTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		event = g.HandleKeybinding(event)
		if event == nil {
			return nil
		}
		return projectTable.HandleKeybinding(event)
	})
	projectTable.onChange = func(p *ProjectInfo) {
		environmentTable.setProject(p)
	}

	organizationTable := newOrganizationTable(g.app, g.manager)
	organizationTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		event = g.HandleKeybinding(event)
		if event == nil {
			return nil
		}
		return organizationTable.HandleKeybinding(event)
	})
	organizationTable.onChange = func(o *OrganizationInfo) {
		environmentTable.setProject(nil)
		projectTable.setOrganization(o)
		g.app.SetFocus(projectTable)
	}

	g.state.panels.panel = append(g.state.panels.panel, organizationTable, projectTable, environmentTable)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(organizationTable, 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(projectTable, 0, 1, true).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(environmentTable, 0, 1, true).
				AddItem(infoBox, 0, 4, false), 0, 1, false),
			0, 4, false)

	g.app.SetRoot(flex, true).EnableMouse(true)
}

func (g *Gui) Start() error {
	g.init()
	if err := g.app.Run(); err != nil {
		g.app.Stop()

		return err
	}

	return nil
}

func (g *Gui) Stop() error {
	g.app.Stop()

	return nil
}
