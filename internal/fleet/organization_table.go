package fleet

import (
	"sort"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type organizationTable struct {
	*tview.Table
	app           *tview.Application
	manager       *FleetManager
	onChange      func(*OrganizationInfo)
	organizations []OrganizationInfo
	selected      *OrganizationInfo
	stopChan      chan int
	reloadMutex   sync.Mutex
}

func newOrganizationTable(app *tview.Application, manager *FleetManager) *organizationTable {
	organizationTable := &organizationTable{
		Table:   tview.NewTable(),
		app:     app,
		manager: manager,
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
		organizationTable.selected = &organizationTable.organizations[index]
		if organizationTable.onChange != nil {
			organizationTable.onChange(organizationTable.selected)
		}
	})

	organizationTable.reload(false)

	return organizationTable
}

func (v *organizationTable) name() string {
	return "Organizations"
}

func (v *organizationTable) reload(force bool) {
	if force {
		v.reloadMutex.Lock()
	} else {
		if v.reloadMutex.TryLock() == false {
			return
		}
	}
	defer v.reloadMutex.Unlock()

	spinTitle(v.app, v, "Organizations", func() {
		organizations, _ := v.manager.Organization.ListAll()

		sort.Slice(organizations, func(i, j int) bool {
			return organizations[i].Label < organizations[j].Label
		})

		v.organizations = organizations
		v.redraw()
	})
}

var organizationHeaders = []string{
	"Name",
}

func (v *organizationTable) redraw() {
	go v.app.QueueUpdateDraw(func() {
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
			v.SetCell(i+1, 0, &tview.TableCell{
				Text:        tview.Escape(organization.Label),
				Transparent: true,
				Clicked: func() bool {
					v.handleSelect(i+1, organization)
					return true
				},
			})
		}
	})
}

func (v *organizationTable) handleSelect(row int, organization OrganizationInfo) {
	v.Select(row, 0)
	v.selected = &organization
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
