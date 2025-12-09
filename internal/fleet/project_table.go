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
		project := &projectTable.projects[index]
		if projectTable.onChange != nil {
			projectTable.onChange(project)
		}
		projectTable.selected = project
	})

	projectTable.reload(false)

	return projectTable
}

func (v *projectTable) setOrganization(organization *OrganizationInfo) {
	v.organization = organization
	v.projects = make([]ProjectInfo, 0)

	go v.app.QueueUpdateDraw(func() {
		v.Clear()
	})

	if v.loadingCancel != nil {
		v.loadingCancel()
		v.loadingCancel = nil
	}
	v.reload(true)
}

func (v *projectTable) name() string {
	return "Projects"
}

func (v *projectTable) reload(force bool) {
	if v.loadingCancel != nil {
		v.loadingCancel()
		v.loadingCancel = nil
	}

	if v.organization == nil {
		v.projects = make([]ProjectInfo, 0)
		v.redraw()

		return
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	v.loadingCancel = cancelCtx
	spinTitle(v.app, v, "Projects", func() {
		projectChan := v.manager.Project.Subscribe(v.organization)
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
		v.loadingCancel = nil
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
			v.SetCell(i+1, 0, &tview.TableCell{
				Text:        tview.Escape(project.ProjectID),
				Transparent: true,
				Clicked: func() bool {
					v.handleSelect(i+1, project)
					return true
				},
			})

			v.SetCell(i+1, 1, &tview.TableCell{
				Text:        tview.Escape(project.ProjectTitle),
				Transparent: true,
				Clicked: func() bool {
					v.handleSelect(i+1, project)
					return true
				},
				Expansion: 1,
			})

			defaultEnvironment := project.ProductionEnvironment
			if defaultEnvironment == nil {
				v.SetCell(i+1, 2, &tview.TableCell{
					Text:        tview.Escape("<unknown>"),
					Transparent: true,
					Clicked: func() bool {
						v.handleSelect(i+1, project)
						return true
					},
				})
			} else {

				v.SetCell(i+1, 2, &tview.TableCell{
					Text:        tview.Escape(defaultEnvironment.Title),
					Transparent: true,
					Clicked: func() bool {
						v.handleSelect(i+1, project)
						return true
					},
				})
				v.SetCell(i+1, 3, &tview.TableCell{
					Text:        tview.Escape(defaultEnvironment.Status),
					Transparent: true,
					Clicked: func() bool {
						v.handleSelect(i+1, project)
						return true
					},
				})
			}
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
			defaultEnvironment := v.selected.ProductionEnvironment
			if defaultEnvironment != nil {
				OpenURL("https://console.upsun.com/" + v.selected.OrganizationName + "/" + v.selected.ProjectID + "/" + defaultEnvironment.Title)
			} else {
				OpenURL("https://console.upsun.com/" + v.selected.OrganizationName + "/" + v.selected.ProjectID)
			}
			return nil
		}
	}

	return event
}
