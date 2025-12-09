package fleet

import (
	"sort"
	"sync"
	"time"

	"github.com/rivo/tview"
)

type organizationList struct {
	*tview.List
	app           *tview.Application
	manager       *FleetManager
	onChange      func(*OrganizationInfo)
	organizations []OrganizationInfo
	stopChan      chan int
	reloadMutex   sync.Mutex
}

func newOrganizationList(app *tview.Application, manager *FleetManager) *organizationList {
	organizationList := &organizationList{
		List:    tview.NewList().ShowSecondaryText(false),
		manager: manager,
		app:     app,
	}

	organizationList.SetTitle("Organizations").SetTitleAlign(tview.AlignLeft)
	organizationList.SetBorder(true)
	organizationList.SetHighlightFullLine(true)
	organizationList.SetSelectedFocusOnly(true)

	organizationList.reload()

	return organizationList
}

func (v *organizationList) name() string {
	return "Organizations"
}

func (v *organizationList) reload() {
	if v.reloadMutex.TryLock() == false {
		return
	}
	defer v.reloadMutex.Unlock()

	spinTitle(v.app, v, "Organizations", func() {
		organizations, _ := v.manager.Organization.List()

		keys := make([]string, 0, len(organizations))
		tmpMap := make(map[string]OrganizationInfo)

		for _, o := range organizations {
			tmpMap[o.Label] = o

			keys = append(keys, o.Label)
		}

		v.organizations = make([]OrganizationInfo, 0)
		sort.Strings(keys)

		for _, key := range keys {
			v.organizations = append(v.organizations, tmpMap[key])
		}
		v.redraw()
	})

}

func (v *organizationList) redraw() {
	go v.app.QueueUpdateDraw(func() {
		currentItem := v.GetCurrentItem()
		list := v.Clear()

		for _, organization := range v.organizations {
			list.AddItem(tview.Escape(organization.Label), organization.ID, 0, func() {
				v.onChange(&organization)
			})
		}
		v.SetCurrentItem(currentItem)
	})
}

func (v *organizationList) stopMonitoring() {
	v.stopChan <- 1
}

func (v *organizationList) startMonitoring() {
	stop := make(chan int, 1)
	v.stopChan = stop
	ticker := time.NewTicker(60 * time.Second)

LOOP:
	for {
		select {
		case <-ticker.C:
			v.reload()
		case <-v.stopChan:
			ticker.Stop()
			break LOOP
		}
	}
}
