package fleet

import (
	"context"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type organizationTable struct {
	*tview.Table
	app           *tview.Application
	manager       *FleetManager
	logger        *logrus.Entry
	onChange      func(*OrganizationInfo)
	organizations []OrganizationInfo
	selected      *OrganizationInfo
	stopChan      chan int
	loadingCancel context.CancelFunc
}

func newOrganizationTable(app *tview.Application, manager *FleetManager, logger *logrus.Entry) *organizationTable {
	organizationTable := &organizationTable{
		Table:   tview.NewTable(),
		app:     app,
		manager: manager,
		logger:  logger.WithField("component", "OorganizationTable"),
	}

	organizationTable.SetTitle("Organizations").SetTitleAlign(tview.AlignLeft)
	organizationTable.SetBorder(true)
	organizationTable.SetSelectable(true, false)
	organizationTable.SetSelectedStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorBlack))
	organizationTable.SetSelectionChangedFunc(func(row, column int) {
		index := row - 1
		if index < 0 || index >= len(organizationTable.organizations) {
			return
		}
		organizationTable.selected = &organizationTable.organizations[index]
	})
	organizationTable.SetSelectedFunc(func(row, column int) {
		index := row - 1
		if index < 0 || index >= len(organizationTable.organizations) {
			return
		}
		organizationTable.setSelected(&organizationTable.organizations[index])
	})

	organizationTable.reload()

	return organizationTable
}

func (v *organizationTable) reload() {
	v.logger.Debug("reloading")
	if v.loadingCancel != nil {
		v.loadingCancel()
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	v.loadingCancel = cancelCtx
	spinTitle(v.app, v, "Organizations", func() {
		organizations, _ := v.manager.Organization.ListAll(ctx)

		sort.Slice(organizations, func(i, j int) bool {
			return organizations[i].Label < organizations[j].Label
		})

		v.organizations = organizations
		v.redraw()
		v.ScrollToBeginning()
	})
}

var organizationHeaders = []string{
	"Name",
}

func (v *organizationTable) redraw() {
	go v.app.QueueUpdateDraw(func() {
		v.logger.Debug("redrawing")
		v.Clear()
		for i, header := range organizationHeaders {
			v.SetCell(0, i, &tview.TableCell{
				Text:            header,
				NotSelectable:   true,
				Color:           tcell.ColorWhite,
				BackgroundColor: tcell.ColorDefault,
				Attributes:      tcell.AttrBold,
				Expansion:       1,
			})
		}
		if len(v.organizations) == 0 {
			return
		}

		for i, organization := range v.organizations {
			handleClick := func() bool {
				v.Select(i+1, 0)
				v.setSelected(&organization)
				return true
			}

			v.SetCell(i+1, 0, &tview.TableCell{
				Text:        tview.Escape(organization.Label),
				Transparent: true,
				Clicked:     handleClick,
			})
		}
	})
}

func (v *organizationTable) setSelected(organization *OrganizationInfo) {
	v.logger.WithField("organization", organization).Debug("selecting")
	v.selected = organization
	if v.onChange != nil {
		v.onChange(v.selected)
	}
}

func (v *organizationTable) HandleKeybinding(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'o':
		if v.selected != nil {
			OpenURL("https://console.upsun.com/" + v.selected.Name)
			return nil
		}
	}

	return event
}
