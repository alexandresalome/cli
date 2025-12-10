package fleet

import (
	"context"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type environmentTable struct {
	*tview.Table
	app           *tview.Application
	manager       *FleetManager
	onChange      func(*Environment)
	environments  []Environment
	selected      *Environment
	stopChan      chan int
	project       *ProjectInfo
	loadingCancel context.CancelFunc
}

func newEnvironmentTable(app *tview.Application, manager *FleetManager) *environmentTable {
	environmentTable := &environmentTable{
		Table:   tview.NewTable(),
		app:     app,
		manager: manager,
	}

	environmentTable.SetTitle("Environments").SetTitleAlign(tview.AlignLeft)
	environmentTable.SetBorder(true)
	environmentTable.SetSelectable(true, false)
	environmentTable.SetSelectedStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorBlack))
	environmentTable.SetSelectionChangedFunc(func(row, column int) {
		index := row - 1
		if index < 0 || index >= len(environmentTable.environments) {
			return
		}
		environment := &environmentTable.environments[index]
		if environmentTable.onChange != nil {
			environmentTable.onChange(environment)
		}
		environmentTable.selected = environment
	})

	environmentTable.reload()

	return environmentTable
}

func (v *environmentTable) setProject(project *ProjectInfo) {
	v.project = project
	if project != nil {
		v.selected = project.DefaultEnvironment
	} else {
		v.selected = nil
	}
	v.environments = make([]Environment, 0)

	go v.app.QueueUpdateDraw(func() {
		v.Clear()
	})

	v.reload()
}

func (v *environmentTable) name() string {
	return "Environments"
}

func (v *environmentTable) reload() {
	if v.loadingCancel != nil {
		v.loadingCancel()
	}

	if v.project == nil {
		v.environments = make([]Environment, 0)
		v.redraw()

		return
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	v.loadingCancel = cancelCtx
	spinTitle(v.app, v, "Environments", func() {
		environments, _ := v.manager.Environment.List(v.project, ctx)

		sort.Slice(environments, func(i, j int) bool {

			return environments[i].Info.Title < environments[j].Info.Title
		})

		v.environments = environments
		if v.selected != nil {
			found := -1
			for i, env := range v.environments {
				if env.Ref == v.selected.Ref {
					found = i
					break
				}
			}
			if found < 0 {
				v.selected = nil
			} else {
				v.Select(found, 0)
			}
		}
		v.redraw()
	})
}

var environmentHeaders = []string{
	"Id",
	"Name",
	"Status",
}

func (v *environmentTable) redraw() {
	go v.app.QueueUpdateDraw(func() {
		v.Clear()
		if v.project == nil {
			v.SetCell(0, 0, &tview.TableCell{
				Text:          "<select a project>",
				NotSelectable: true,
				Align:         tview.AlignLeft,
				Expansion:     1,
			})

			return
		}
		for i, header := range environmentHeaders {
			v.SetCell(0, i, &tview.TableCell{
				Text:            header,
				NotSelectable:   true,
				Align:           tview.AlignLeft,
				Color:           tcell.ColorWhite,
				BackgroundColor: tcell.ColorDefault,
				Attributes:      tcell.AttrBold,
			})
		}
		if len(v.environments) == 0 {
			return
		}

		for i, environment := range v.environments {
			handleClick := func() bool {
				v.handleSelect(i+1, environment)
				return true
			}
			v.SetCell(i+1, 0, &tview.TableCell{
				Text:        tview.Escape(environment.Ref),
				Transparent: true,
				Clicked:     handleClick,
			})

			v.SetCell(i+1, 1, &tview.TableCell{
				Text:        tview.Escape(environment.Info.Title),
				Transparent: true,
				Clicked:     handleClick,
				Expansion:   1,
			})

			v.SetCell(i+1, 2, &tview.TableCell{
				Text:        tview.Escape(environment.Info.Status),
				Transparent: true,
				Clicked:     handleClick,
				Expansion:   1,
			})
		}
	})
}

func (v *environmentTable) handleSelect(row int, environment Environment) {
	v.Select(row, 0)
	v.selected = &environment
	if v.onChange != nil {
		v.onChange(v.selected)
	}
}

func (v *environmentTable) HandleKeybinding(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'o':
		if v.project != nil {
			if v.selected != nil {
				OpenURL("https://console.upsun.com/" + v.project.OrganizationName + "/" + v.project.ProjectID + "/" + v.selected.Ref)
			} else {
				OpenURL("https://console.upsun.com/" + v.project.OrganizationName + "/" + v.project.ProjectID)
			}
			return nil
		}
	}

	return event
}
