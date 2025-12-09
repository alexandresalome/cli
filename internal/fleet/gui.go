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
	app              *tview.Application
	organizationList *organizationList
	state            *state
	manager          *FleetManager
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
	projectList := newProjectList(g.app, g.manager)
	projectList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return g.HandleGlobalKeybinding(event)
	})
	projectList.onChange = func(p ProjectInfo) {
		//environmentList.setProject(p)
	}

	organizationList := newOrganizationList(g.app, g.manager)
	organizationList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return g.HandleGlobalKeybinding(event)
	})
	organizationList.onChange = func(o *OrganizationInfo) {
		projectList.setOrganization(o)
		g.app.SetFocus(projectList)
	}

	g.state.panels.panel = append(g.state.panels.panel, organizationList, projectList)

	environmentList := tview.NewList().
		ShowSecondaryText(false).
		AddItem("xxx", "", 0, func() {}).
		AddItem("yyy", "", 0, func() {})
	environmentList.SetTitle("Environments").SetBorder(true)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(organizationList, 10, 1, true).
			AddItem(projectList, 0, 1, false).
			AddItem(environmentList, 10, 1, false),
			40, 1, true).
		AddItem(tview.NewBox().SetBorder(true), 0, 1, false)

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
