package fleet

import (
	"context"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type environmentTable struct {
	*tview.Table
	app           *tview.Application
	manager       *FleetManager
	logger        *logrus.Entry
	onChange      func(*Environment)
	environments  []Environment
	selected      *Environment
	project       *ProjectInfo
	loadingCancel context.CancelFunc
}

func newEnvironmentTable(app *tview.Application, manager *FleetManager, logger *logrus.Entry) *environmentTable {
	environmentTable := &environmentTable{
		Table:   tview.NewTable(),
		app:     app,
		manager: manager,
		logger:  logger.WithField("component", "EnvironmentTable"),
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

		environmentTable.setSelected(&environmentTable.environments[index])
	})

	environmentTable.reload()

	return environmentTable
}

func (v *environmentTable) setProject(project *ProjectInfo) {
	v.project = project
	v.setSelected(nil)
	v.environments = make([]Environment, 0)
	v.redraw()

	v.reload()
}

func (v *environmentTable) reload() {
	v.logger.Debug("reloading")
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

		envMap := make(map[string]Environment)
		for _, env := range environments {
			envMap[env.Ref] = env
		}

		v.environments = environments
		// select best environment
		found := -1
		if v.selected != nil {
			for i, env := range v.environments {
				if env.Ref == v.selected.Ref {
					found = i
					break
				}
			}
		}
		if found < 0 {
			for i, e := range v.environments {
				if e.IsDefault() {
					found = i
					break
				}
			}
		}
		if found < 0 && len(v.environments) > 0 {
			found = 0
		}
		if found < 0 {
			v.setSelected(nil)
		} else {
			v.setSelected(&v.environments[found])
			v.Select(found+1, 0)
		}
		v.redraw()
		v.ScrollToBeginning()

		envChan := v.manager.Environment.Subscribe(v.project, ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case env, ok := <-envChan:
				if !ok {
					return
				}

				envMap[env.Ref] = *env
				newEnvs := make([]Environment, 0, len(envMap))
				for _, e := range envMap {
					newEnvs = append(newEnvs, e)
				}
				sort.Slice(newEnvs, func(i, j int) bool {
					return newEnvs[i].Info.Title < newEnvs[j].Info.Title
				})
				v.environments = newEnvs
				// reselect
				if v.selected != nil {
					for i, e := range v.environments {
						if e.Ref == v.selected.Ref {
							found = i
							break
						}
					}
				}
				if found < 0 {
					for i, e := range v.environments {
						if e.IsDefault() {
							found = i
							break
						}
					}
				}
				if found < 0 && len(v.environments) > 0 {
					found = 0
				}
				if found < 0 {
					v.setSelected(nil)
				} else {
					v.setSelected(&v.environments[found])
					v.Select(found+1, 0)
				}
				v.redraw()
			}
		}
	})
}

func (v *environmentTable) drawHeader(col int, text string, minWidth int, expansion int) {
	if len(text) < minWidth {
		for len(text) < minWidth {
			text += " "
		}
	}
	v.SetCell(0, col, &tview.TableCell{
		Text:            text,
		NotSelectable:   true,
		Align:           tview.AlignLeft,
		Color:           tcell.ColorWhite,
		BackgroundColor: tcell.ColorDefault,
		Attributes:      tcell.AttrBold,
		Expansion:       expansion,
	})
}

func (v *environmentTable) redraw() {
	go v.app.QueueUpdateDraw(func() {
		v.logger.Debug("redrawing")
		v.Clear()
		if v.project == nil {
			v.ScrollToBeginning()
			v.SetCell(0, 0, &tview.TableCell{
				Text:          "<select a project>",
				NotSelectable: true,
				Align:         tview.AlignLeft,
				Expansion:     1,
			})

			return
		}
		v.drawHeader(1, "Id", 10, 0)
		v.drawHeader(0, "Status", 0, 1)
		if len(v.environments) == 0 {
			return
		}

		for i, environment := range v.environments {
			handleClick := func() bool {
				v.Select(i+1, 0)
				v.setSelected(&environment)
				return true
			}
			deployStatus := ""
			if environment.IsDefault() {
				deployStatus += "⭐ "
			}
			if environment.Details != nil {
				if environment.Details.LastDeploymentAt == "" {
					deployStatus += "⏳ "
				} else if environment.Details.LastDeploymentSuccessful {
					deployStatus += "✅ "
				} else {
					deployStatus += "🚩 "
				}
			}
			deployStatus += environment.Info.Status

			v.SetCell(i+1, 0, &tview.TableCell{
				Text:        tview.Escape(deployStatus),
				Transparent: true,
				Clicked:     handleClick,
			})

			v.SetCell(i+1, 1, &tview.TableCell{
				Text:        tview.Escape(environment.Ref),
				Transparent: true,
				Clicked:     handleClick,
				Expansion:   1,
			})
		}
	})
}

func (v *environmentTable) setSelected(environment *Environment) {
	if v.selected == environment {
		return
	}
	v.logger.WithField("environment", environment).Debug("selecting")

	v.selected = environment
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
