package fleet

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type infoBox struct {
	*tview.Table
	app         *tview.Application
	manager     *FleetManager
	environment *Environment
}

func newInfoBox(app *tview.Application, manager *FleetManager) *infoBox {
	infoBox := &infoBox{
		Table:   tview.NewTable(),
		app:     app,
		manager: manager,
	}

	infoBox.SetBorder(true)

	return infoBox
}

func (v *infoBox) setEnvironment(environment *Environment) {
	v.environment = environment
	v.reload()
}

func (v *infoBox) reload() {
	v.redraw()
}

func (v *infoBox) setField(row int, label string, value string) {
	v.SetCell(row, 0, tview.NewTableCell(tview.Escape(label)).SetAttributes(tcell.AttrBold).SetSelectable(false))
	v.SetCell(row, 1, tview.NewTableCell(tview.Escape(value)).SetSelectable(false))

}
func (v *infoBox) redraw() {
	go v.app.QueueUpdateDraw(func() {
		v.Clear()
		if v.environment == nil {
			return
		}

		i := 0
		v.setField(0, "Organization", v.environment.Project.OrganizationLabel)
		v.setField(1, "Project", v.environment.Project.ProjectTitle)
		v.setField(2, "Environment", v.environment.Ref)
		if v.environment.Info != nil {
			v.setField(i, "Status", v.environment.Info.Status)
			v.setField(i+1, "Status", v.environment.Info.Status)
			i += 2
		}
		if v.environment.Details != nil {
			v.setField(i, "Domain", v.environment.Details.DefaultDomain)
			v.setField(i+1, "Last Deploy At", v.environment.Details.LastDeploymentAt)
			deploySuccess := "🚩"
			if v.environment.Details.LastDeploymentSuccessful {
				deploySuccess = "✅"
			}
			v.setField(i+2, "Last Deploy Success", deploySuccess)
			i += 3
		}
	})
}
func (v *infoBox) HandleKeybinding(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'o':
		if v.environment != nil {
			OpenURL("https://console.upsun.com/" + v.environment.Project.OrganizationName + "/" + v.environment.Project.ProjectID + "/" + v.environment.Ref)
			return nil
		}
	}

	return event
}
