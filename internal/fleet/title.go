package fleet

import (
	"time"

	"github.com/rivo/tview"
)

type SetTitleable interface {
	SetTitle(title string) *tview.Box
}

func spinTitle(app *tview.Application, box SetTitleable, title string, action func()) {
	done := make(chan bool)

	//action
	go func() {
		action()
		done <- true
		close(done)
	}()

	// spinner
	go func() {
		spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		var i int
		for {
			select {
			case _ = <-done:
				go app.QueueUpdateDraw(func() {
					box.SetTitle(title)
				})
				return
			case <-time.After(100 * time.Millisecond):
				spin := i % len(spinners)
				go app.QueueUpdateDraw(func() {
					box.SetTitle(title + " " + spinners[spin])
				})
				i++
			}
		}
	}()
}
