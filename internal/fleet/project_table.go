package fleet

import (
	"context"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type projectTable struct {
	*tview.Table
	app           *tview.Application
	manager       *FleetManager
	onChange      func(*ProjectInfo)
	projects      []ProjectInfo
	selected      *ProjectInfo
	stopChan      chan int
	organization  *OrganizationInfo
	loadingCancel context.CancelFunc
}

func newProjectTable(app *tview.Application, manager *FleetManager) *projectTable {
	projectTable := &projectTable{
		Table:   tview.NewTable(),
		app:     app,
		manager: manager,
	}

	projectTable.SetTitle("Projects").SetTitleAlign(tview.AlignLeft)
	projectTable.SetBorder(true)
	projectTable.SetSelectable(true, false)
	projectTable.SetSelectedStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorBlack))
	projectTable.SetSelectionChangedFunc(func(row, column int) {
		index := row - 1
		if index < 0 || index >= len(projectTable.projects) {
			return
		}
		projectTable.selected = &projectTable.projects[index]
	})
	projectTable.SetSelectedFunc(func(row, column int) {
		index := row - 1
		if index < 0 || index >= len(projectTable.projects) {
			return
		}
		projectTable.selected = &projectTable.projects[index]
		if projectTable.onChange != nil {
			projectTable.onChange(projectTable.selected)
		}
	})

	projectTable.reload()

	return projectTable
}

func (v *projectTable) setOrganization(organization *OrganizationInfo) {
	v.organization = organization
	v.selected = nil
	v.projects = make([]ProjectInfo, 0)

	go v.app.QueueUpdateDraw(func() {
		v.Clear()
	})

	v.reload()
}

func (v *projectTable) name() string {
	return "Projects"
}

func (v *projectTable) reload() {
	if v.loadingCancel != nil {
		v.loadingCancel()
	}

	if v.organization == nil {
		v.projects = make([]ProjectInfo, 0)
		v.redraw()

		return
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	v.loadingCancel = cancelCtx
	spinTitle(v.app, v, "Projects", func() {
		projectChan := v.manager.Project.Subscribe(v.organization, ctx)
		projectMap := make(map[string]ProjectInfo)

		for {
			select {
			case <-ctx.Done():
				return
			case project, ok := <-projectChan:
				if !ok {
					return
				}

				projectMap[project.ProjectID] = project

				projects := []ProjectInfo{}
				for _, p := range projectMap {
					projects = append(projects, p)
				}
				sort.Slice(projects, func(i, j int) bool {
					return projects[i].ProjectTitle < projects[j].ProjectTitle
				})
				v.projects = projects
				v.redraw()
			}
		}
	})
}

var projectHeaders = []string{
	"Id",
	"Name",
	"Environment",
	"Status",
}

func (v *projectTable) redraw() {
	go v.app.QueueUpdateDraw(func() {
		v.Clear()
		if v.organization == nil {
			v.SetCell(0, 0, &tview.TableCell{
				Text:          "<select an org>",
				NotSelectable: true,
				Align:         tview.AlignLeft,
				Expansion:     1,
			})

			return
		}
		for i, header := range projectHeaders {
			v.SetCell(0, i, &tview.TableCell{
				Text:            header,
				NotSelectable:   true,
				Align:           tview.AlignLeft,
				Color:           tcell.ColorWhite,
				BackgroundColor: tcell.ColorDefault,
				Attributes:      tcell.AttrBold,
			})
		}
		if len(v.projects) == 0 {
			return
		}

		for i, project := range v.projects {
			handleClick := func() bool {
				v.handleSelect(i+1, project)
				return true
			}
			v.SetCell(i+1, 0, &tview.TableCell{
				Text:        tview.Escape(project.ProjectID),
				Transparent: true,
				Clicked:     handleClick,
			})

			v.SetCell(i+1, 1, &tview.TableCell{
				Text:        tview.Escape(project.ProjectTitle),
				Transparent: true,
				Clicked:     handleClick,
				Expansion:   1,
			})

			defaultEnvironment := project.DefaultEnvironment
			titleText := ""
			statusText := ""

			if defaultEnvironment != nil && defaultEnvironment.Info != nil {
				titleText = defaultEnvironment.Info.Title
				statusText = defaultEnvironment.Info.Status
			}

			v.SetCell(i+1, 2, &tview.TableCell{
				Text:        tview.Escape(titleText),
				Transparent: true,
				Clicked:     handleClick,
			})
			v.SetCell(i+1, 3, &tview.TableCell{
				Text:        tview.Escape(statusText),
				Transparent: true,
				Clicked:     handleClick,
			})
		}
	})
}

func (v *projectTable) handleSelect(row int, project ProjectInfo) {
	v.Select(row, 0)
	v.selected = &project
	if v.onChange != nil {
		v.onChange(v.selected)
	}
}

func (v *projectTable) HandleKeybinding(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'o':
		if v.selected != nil {
			defaultEnvironment := v.selected.DefaultEnvironment
			if defaultEnvironment != nil && defaultEnvironment.Info != nil {
				OpenURL("https://console.upsun.com/" + v.selected.OrganizationName + "/" + v.selected.ProjectID + "/" + defaultEnvironment.Info.Title)
			} else {
				OpenURL("https://console.upsun.com/" + v.selected.OrganizationName + "/" + v.selected.ProjectID)
			}
			return nil
		}
	}

	return event
}
