package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/andygrunwald/go-jira"
	c "github.com/jroimartin/gocui"
)

var currentSearch string

func resetQuery(g *c.Gui, v *c.View) error {
	currentSearch = ""
	selectQuery(g, v)
	return nil
}

var done = make(chan bool)
var abort = make(chan bool)

func selectQuery(g *c.Gui, v *c.View) error {
	select {
	case abort <- true:
		break
	default:
		break
	}

	currItem := QueryList.CurrentItem()
	log.Println(currItem)
	IssueList.Reset()
	IssueList.Clear()
	IssueView.Clear()

	IssueList.Focus(g)
	g.SelFgColor = c.ColorGreen | c.AttrBold
	IssueList.Title = " Fetching ..."
	//	g.Update(func(g *c.Gui) error {
	query := GetQuery(currItem.(string))
	if len(currentSearch) > 0 {
		query = fmt.Sprintf("(summary ~ \"%s\" OR description ~ \"%s\") AND %s", currentSearch, currentSearch, query)
	}
	log.Printf("query = %s", query)
	// ct := 0
	cevent := make(chan interface{})
	go func() {
		// ct := 0
		for {
			select {
			case <-done:
				g.Update(func(g *c.Gui) error {
					IssueList.SetTitle("Done")
					return nil
				})
				return
			case iss := <-cevent:
				g.Update(func(g *c.Gui) error {
					IssueList.AddItem(g, iss)
					return nil
				})

			}
		}

	}()
	cb := func(issue jira.Issue) error {
		cevent <- issue
		select {
		case <-abort:
			return errors.New("abort")
		default:
			break
		}
		return nil
	}

	go func() {
		defer func() {
			done <- true
		}()

		GetClient().Issue.SearchPages(query, nil, cb)
		IssueList.Focus(g)
		log.Printf("IssueList type %T", IssueList)
		log.Printf("Done SetItems")

	}()
	return nil
}

func OpenBrowser(g *c.Gui, v *c.View) error {
	currItem := IssueList.CurrentItem()
	if currItem == nil {
		return nil
	}
	url := fmt.Sprintf("https://appen.atlassian.net/browse/%s", currItem.(jira.Issue).Key)
	if v.Name() == LIST_VIEW {
		cmd := exec.Command("open", url)

		if err := cmd.Run(); err != nil {
			log.Println("Error on opening browser", err)
			return err
		}
	}
	return nil
}

func ListUp(g *c.Gui, v *c.View) error {
	switch v.Name() {

	case SIDE_VIEW:
		if err := QueryList.MoveUp(); err != nil {
			log.Println("Error on SitesList.MoveUp()", err)
			return err
		}
	case LIST_VIEW:
		if err := IssueList.MoveUp(); err != nil {
			log.Println("Error on NewsList.MoveUp()", err)
			return err
		}
		if err := UpdateIssue(); err != nil {
			log.Println("Error on UpdateIssue()", err)
			return err
		}
	}
	return nil
}

func UpdateIssue() error {
	IssueView.Clear()

	currItem := IssueList.CurrentItem()
	if currItem == nil {
		return nil
	}
	issue := currItem.(jira.Issue)

	description := strings.Replace(issue.Fields.Description, "\r\n", "\n", -1)
	fmt.Fprintf(IssueView, "%s", strings.TrimSpace(description))
	return nil
}

func ListDown(g *c.Gui, v *c.View) error {
	log.Println("current view", v.Name())
	switch v.Name() {

	case SIDE_VIEW:
		if err := QueryList.MoveDown(); err != nil {
			log.Println("Error on QueryList.MoveDown()", err)
			return err
		}
	case LIST_VIEW:
		if err := IssueList.MoveDown(); err != nil {
			log.Println("Error on IssueList.MoveDown()", err)
			return err
		}
		if err := UpdateIssue(); err != nil {
			log.Println("Error on UpdateIssue()", err)
			return err
		}
	}
	return nil
}

func ListPgDown(g *c.Gui, v *c.View) error {
	switch v.Name() {

	case SIDE_VIEW:
		if err := QueryList.MovePgDown(); err != nil {
			log.Println("Error on QueryList.MovePgDown()", err)
			return err
		}
	case LIST_VIEW:
		if err := IssueList.MovePgDown(); err != nil {
			log.Println("Error on IssueList.MovePgDown()", err)
			return err
		}
		if err := UpdateIssue(); err != nil {
			log.Println("Error on UpdateIssue()", err)
			return err
		}
	}
	return nil
}

func closeDetailView(g *c.Gui, v *c.View) error {
	g.Cursor = false
	g.DeleteView(DETAIL_VIEW)
	IssueList.Focus(g)
	return nil
}

func ListPgUp(g *c.Gui, v *c.View) error {
	switch v.Name() {

	case SIDE_VIEW:
		if err := QueryList.MovePgUp(); err != nil {
			log.Println("Error on QueryList.MovePgUp()", err)
			return err
		}
	case LIST_VIEW:
		if err := IssueList.MovePgUp(); err != nil {
			log.Println("Error on IssueList.MovePgUp()", err)
			return err
		}
		if err := UpdateIssue(); err != nil {
			log.Println("Error on UpdateIssue()", err)
			return err
		}
	}
	return nil
}

func createPromptView(g *c.Gui, title string) error {
	tw, th := g.Size()
	v, err := g.SetView(PROMPT_VIEW, tw/6, (th/2)-1, (tw*5)/6, (th/2)+1)
	if err != nil && err != c.ErrUnknownView {
		return err
	}
	v.Editable = true
	setTopWindowTitle(g, PROMPT_VIEW, title)

	g.Cursor = true
	_, err = g.SetCurrentView(PROMPT_VIEW)

	return err
}

// createContentView creates a view where the contents of thecurrently selected
// event will be displayed
func createDetailView(g *c.Gui) error {
	tw, th := g.Size()
	v, err := g.SetView(DETAIL_VIEW, tw/8, th/8, (tw*7)/8, (th*7)/8)
	v.Wrap = true
	v.Autoscroll = true
	if err != nil && err != c.ErrUnknownView {
		return err
	}
	curr_issue := IssueList.CurrentItem().(jira.Issue)
	log.Println(curr_issue.Fields.Description)
	myText := strings.Replace(curr_issue.Fields.Description, "\r\n", "\n", -1)
	v.Write([]byte(myText))
	setTopWindowTitle(g, DETAIL_VIEW, "")
	_, err = g.SetCurrentView(DETAIL_VIEW)

	return err
}

// deletePromptView deletes the current prompt view
func deletePromptView(g *c.Gui) error {
	g.Cursor = false
	return g.DeleteView(PROMPT_VIEW)
}
func setTopWindowTitle(g *c.Gui, view_name, title string) {
	v, err := g.View(view_name)
	if err != nil {
		log.Println("Error on setTopWindowTitle", err)
		return
	}
	v.Title = fmt.Sprintf("%v (Ctrl-q to close)", title)
}

func Search(g *c.Gui, v *c.View) error {
	if err := createPromptView(g, "Search with multiple terms:"); err != nil {
		log.Println("Error on createPromptView", err)
		return err
	}

	return nil
}

func ShowDetailView(g *c.Gui, v *c.View) error {
	if err := createDetailView(g); err != nil {
		log.Println("Error on createPromptView", err)
		return err
	}

	return nil
}

func PerformSearch(g *c.Gui, v *c.View) error {
	IssueList.Reset()
	IssueList.Clear()
	IssueList.Focus(g)
	// QueryList.Unfocus()
	IssueList.Title = " Searching ... "
	deletePromptView(g)
	currentSearch = strings.TrimSpace(v.ViewBuffer())
	selectQuery(g, v)
	g.SetCurrentView(LIST_VIEW)
	// QueryList.Focus(g)
	return nil
}
