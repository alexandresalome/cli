package fleet

import (
	"sort"
	"sync"
	"time"

	"github.com/rivo/tview"
)

type projectList struct {
	*tview.List
	app          *tview.Application
	manager      *FleetManager
	onChange     func(ProjectInfo)
	projects     []ProjectInfo
	stopChan     chan int
	reloadMutex  sync.Mutex
	organization *OrganizationInfo
}

func newProjectList(app *tview.Application, manager *FleetManager) *projectList {
	projectList := &projectList{
		List:    tview.NewList().ShowSecondaryText(false),
		app:     app,
		manager: manager,
	}

	projectList.SetTitle("Projects").SetTitleAlign(tview.AlignLeft)
	projectList.SetBorder(true)
	projectList.SetHighlightFullLine(true)

	projectList.reload(false)

	return projectList
}

func (v *projectList) setOrganization(organization *OrganizationInfo) {
	v.organization = organization
	v.projects = make([]ProjectInfo, 0)

	go v.app.QueueUpdateDraw(func() {
		v.Clear()
	})
	v.reload(true)
}

func (v *projectList) name() string {
	return "Projects"
}

func (v *projectList) reload(force bool) {
	if force {
		v.reloadMutex.Lock()
	} else {
		if v.reloadMutex.TryLock() == false {
			return
		}
	}
	defer v.reloadMutex.Unlock()

	if v.organization == nil {
		v.projects = make([]ProjectInfo, 0)
		v.redraw()

		return
	}

	spinTitle(v.app, v, "Projects", func() {
		projects, _ := v.manager.Project.List(v.organization)

		keys := make([]string, 0, len(projects))
		tmpMap := make(map[string]ProjectInfo)

		for _, p := range projects {
			tmpMap[p.ProjectTitle] = p

			keys = append(keys, p.ProjectTitle)
		}

		v.projects = make([]ProjectInfo, 0)
		sort.Strings(keys)

		for _, key := range keys {
			v.projects = append(v.projects, tmpMap[key])
		}
		v.redraw()
	})

}

func (v *projectList) redraw() {
	go v.app.QueueUpdateDraw(func() {
		currentItem := v.GetCurrentItem()
		list := v.Clear()

		if v.organization == nil {
			list.AddItem("<select an org>", "", 0, func() {})

			return
		}

		if len(v.projects) == 0 {
			list.AddItem("<empty>", "", 0, func() {})

			return
		}

		for _, project := range v.projects {
			list.AddItem(tview.Escape(project.ProjectTitle), "", 0, func() {
				v.onChange(project)
			})
		}
		v.SetCurrentItem(currentItem)
	})
}

func (v *projectList) stopMonitoring() {
	v.stopChan <- 1
}

func (v *projectList) startMonitoring() {
	stop := make(chan int, 1)
	v.stopChan = stop
	ticker := time.NewTicker(60 * time.Second)

LOOP:
	for {
		select {
		case <-ticker.C:
			v.reload(false)
		case <-v.stopChan:
			ticker.Stop()
			break LOOP
		}
	}
}
