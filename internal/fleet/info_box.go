package fleet

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type infoBox struct {
	*tview.Table
	app         *tview.Application
	manager     *FleetManager
	logger      *logrus.Entry
	environment *Environment
}

func newInfoBox(app *tview.Application, manager *FleetManager, logger *logrus.Entry) *infoBox {
	infoBox := &infoBox{
		Table:   tview.NewTable(),
		app:     app,
		manager: manager,
		logger:  logger.WithField("component", "InfoBox"),
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
		v.logger.Debug("redrawing")
		v.Clear()
		if v.environment == nil {
			return
		}

		i := 0
		v.setField(0, "Organization", v.environment.Project.OrganizationLabel)
		v.setField(1, "Project", v.environment.Project.ProjectTitle)
		v.setField(2, "Environment", v.environment.Ref)
		i += 3
		if v.environment.Info != nil {
			v.setField(i, "Status", v.environment.Info.Status)
			i += 1
		}
		if v.environment.Details != nil {
			v.setField(i, "Domain", v.environment.Details.DefaultDomain)

			deploySuccess := ""
			if v.environment.Details.LastDeploymentAt == "" {
				deploySuccess = "⏳ Not deployed"
			} else {
				if v.environment.Details.LastDeploymentSuccessful {
					deploySuccess = "✅ Success"
				} else {
					deploySuccess = "🚩 Failed"
				}
			}

			v.setField(i+1, "Last Deploy", deploySuccess)
			v.setField(i+2, "Last Deploy At", v.environment.Details.LastDeploymentAt)
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
