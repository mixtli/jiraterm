package main

import (
	"fmt"
	c "github.com/jroimartin/gocui"
	"log"
	"time"
)

const (
	SIDE_VIEW       = "side"
	LIST_VIEW       = "list"
	ISSUE_VIEW      = "issue"
	EDIT_ISSUE_VIEW = "edit_issue"
)

var (
	QueryList *List
	IssueList *List
	IssueView *c.View
	curW      int
	curH      int
)

func RunUI() {
	g, err := c.NewGui(c.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	// some basic configuration
	g.SelFgColor = c.ColorGreen | c.AttrBold
	g.BgColor = c.ColorDefault
	g.Highlight = true
	g.Mouse = true

	curW, curH = g.Size()
	rw, rh := relSize(g)

	v, err := g.SetView(SIDE_VIEW, 0, 0, rw, curH-1)
	if err != nil && err != c.ErrUnknownView {
		log.Fatal("Failed to create search list:", err)
	}
	QueryList = CreateList(v, true)
	QueryList.Focus(g)

	// it loads the existing sites if any at the beginning
	g.Update(func(g *c.Gui) error {
		if err := LoadQueries(); err != nil {
			log.Fatal("Error while loading searches", err)
		}
		log.Println("Loaded searches")
		return nil
	})

	// Issues List
	v, err = g.SetView(LIST_VIEW, rw+1, 0, curW-1, rh)
	if err != nil && err != c.ErrUnknownView {
		log.Fatal(" Failed to create issues list:", err)
	}
	IssueList = CreateList(v, true)
	IssueList.SetTitle("Issues")

	IssueView, err = g.SetView(ISSUE_VIEW, rw+1, rh+1, curW-1, curH-1)
	if err != nil && err != c.ErrUnknownView {
		log.Fatal("Failed to create Issue view:", err)
	}
	IssueView.Title = " Issue Summary "

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	g.Update(func(g *c.Gui) error {
		selectQuery(g, QueryList.View)
		g.Update(func(g *c.Gui) error {
			UpdateIssue()
			return nil
		})
		return nil
	})

	// Periodically update the issue list
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			<-ticker.C
			g.Update(func(g *c.Gui) error {
				selectQuery(g, QueryList.View)
				g.Update(func(g *c.Gui) error {
					UpdateIssue()
					return nil
				})
				return nil
			})
			log.Println("ticker")
		}
	}()

	if err := g.MainLoop(); err != nil && err != c.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *c.Gui) error {
	tw, th := g.Size()
	rw, rh := relSize(g)

	if v, err := g.SetView(SIDE_VIEW, 0, 0, rw, th-1); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.SelBgColor = c.ColorGreen
		v.SelFgColor = c.ColorBlack

		searches := GetQueries()
		for i := 0; i < len(searches); i++ {
			fmt.Fprintln(v, searches[i])
		}
	}

	if v, err := g.SetView(LIST_VIEW, rw+1, 0, tw-1, rh); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
		v.Wrap = true
		if _, err := g.SetCurrentView(LIST_VIEW); err != nil {
			return err
		}
	}

	if v, err := g.SetView(ISSUE_VIEW, rw+1, rh+1, tw-1, th-1); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.SelBgColor = c.ColorGreen
		v.SelFgColor = c.ColorBlack
	}

	if _, err := g.View(EDIT_ISSUE_VIEW); err == nil {
		_, err = g.SetView(EDIT_ISSUE_VIEW, tw/8, th/8, (tw*7)/8, (th*7)/8)
		if err != nil && err != c.ErrUnknownView {
			return err
		}
	}

	if curW != tw || curH != th {
		QueryList.ResetPages()
		QueryList.Draw()
		IssueList.ResetPages()
		IssueList.Draw()
		curW = tw
		curH = th
	}

	return nil
}

func keybindings(g *c.Gui) error {
	if err := g.SetKeybinding(SIDE_VIEW, c.KeyTab, c.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(LIST_VIEW, c.KeyTab, c.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(LIST_VIEW, c.KeyCtrlO, c.ModNone, OpenBrowser); err != nil {
		return err
	}
	if err := g.SetKeybinding(ISSUE_VIEW, c.KeyTab, c.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("", c.KeyCtrlC, c.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding(LIST_VIEW, c.KeyArrowUp, c.ModNone, ListUp); err != nil {
		return err
	}

	if err := g.SetKeybinding(LIST_VIEW, c.KeyArrowDown, c.ModNone, ListDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(SIDE_VIEW, c.KeyArrowUp, c.ModNone, ListUp); err != nil {
		return err
	}

	if err := g.SetKeybinding(SIDE_VIEW, c.KeyArrowDown, c.ModNone, ListDown); err != nil {
		return err
	}

	if err := g.SetKeybinding(LIST_VIEW, c.KeyPgup, c.ModNone, ListPgUp); err != nil {
		log.Fatal("Failed to set keybindings")
	}

	if err := g.SetKeybinding(LIST_VIEW, c.KeyPgdn, c.ModNone, ListPgDown); err != nil {
		log.Fatal("Failed to set keybindings")
	}

	// if err := g.SetKeybinding("", c.KeyArrowDown, c.ModNone, cursorDown); err != nil {
	//   return err
	// }
	// if err := g.SetKeybinding("", c.KeyArrowUp, c.ModNone, cursorUp); err != nil {
	//   return err
	// }

	if err := g.SetKeybinding("", c.KeyEnter, c.ModNone, selectQuery); err != nil {
		return err
	}
	return nil
}

func nextView(g *c.Gui, v *c.View) error {
	if v == nil || v.Name() == SIDE_VIEW {
		_, err := g.SetCurrentView(LIST_VIEW)
		return err
	}
	_, err := g.SetCurrentView(SIDE_VIEW)
	return err
}

func quit(g *c.Gui, v *c.View) error {
	return c.ErrQuit
}

func cursorDown(g *c.Gui, v *c.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *c.Gui, v *c.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

// relSize calculates the  sizes of the sites view width
// and the news view height in relation to the current terminal size
func relSize(g *c.Gui) (int, int) {
	tw, th := g.Size()

	return (tw * 3) / 10, (th * 70) / 100
}

func LoadQueries() error {
	QueryList.SetTitle("Queries")

	searches := GetQueries()
	log.Println("number of searches", len(searches))
	if len(searches) == 0 {
		QueryList.SetTitle("No searches in config file")
		QueryList.Reset()
		IssueList.Reset()
		IssueList.SetTitle("No issues...")
		return nil
	}
	data := make([]interface{}, len(searches))
	for i, rr := range searches {
		data[i] = rr
	}

	return QueryList.SetItems(data)
}
