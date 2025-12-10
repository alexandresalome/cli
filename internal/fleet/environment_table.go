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
	onChange      func(*EnvironmentInfo)
	environments  []EnvironmentInfo
	selected      *EnvironmentInfo
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

	environmentTable.reload(false)

	return environmentTable
}

func (v *environmentTable) setProject(project *ProjectInfo) {
	v.project = project
	if project != nil {
		v.selected = project.ProductionEnvironment
	} else {
		v.selected = nil
	}
	v.environments = make([]EnvironmentInfo, 0)

	go v.app.QueueUpdateDraw(func() {
		v.Clear()
	})

	v.reload(true)
}

func (v *environmentTable) name() string {
	return "Environments"
}

func (v *environmentTable) reload(force bool) {
	if v.project == nil {
		v.environments = make([]EnvironmentInfo, 0)
		v.redraw()

		return
	}

	spinTitle(v.app, v, "Environments", func() {
		environments, _ := v.manager.Environment.List(v.project)

		sort.Slice(environments, func(i, j int) bool {
			return environments[i].Title < environments[j].Title
		})

		v.environments = environments
		if v.selected != nil {
			found := -1
			for i, env := range v.environments {
				if env.ID == v.selected.ID {
					found = i
					break
				}
			}
			if found < 0 {
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
				Text:        tview.Escape(environment.ID),
				Transparent: true,
				Clicked:     handleClick,
			})

			v.SetCell(i+1, 1, &tview.TableCell{
				Text:        tview.Escape(environment.Title),
				Transparent: true,
				Clicked:     handleClick,
				Expansion:   1,
			})

			v.SetCell(i+1, 2, &tview.TableCell{
				Text:        tview.Escape(environment.Status),
				Transparent: true,
				Clicked:     handleClick,
				Expansion:   1,
			})
		}
	})
}

func (v *environmentTable) handleSelect(row int, environment EnvironmentInfo) {
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
				OpenURL("https://console.upsun.com/" + v.project.OrganizationName + "/" + v.project.ProjectID + "/" + v.selected.ID)
			} else {
				OpenURL("https://console.upsun.com/" + v.project.OrganizationName + "/" + v.project.ProjectID)
			}
			return nil
		}
	}

	return event
}
