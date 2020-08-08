package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	c "github.com/jroimartin/gocui"
	"log"
	"os/exec"
)

func selectQuery(g *c.Gui, v *c.View) error {
	currItem := QueryList.CurrentItem()
	log.Println(currItem)
	IssueList.Clear()
	IssueView.Clear()

	IssueList.Focus(g)
	g.SelFgColor = c.ColorGreen | c.AttrBold
	IssueList.Title = " Fetching ..."
	g.Update(func(g *c.Gui) error {
		query := GetQuery(currItem.(string))
		issues, _, _ := GetClient().Issue.Search(query, nil)
		IssueList.Focus(g)
		data := make([]interface{}, len(issues))
		for i, issue := range issues {
			data[i] = issue
		}
		log.Printf("IssueList type %T", IssueList)
		result := IssueList.SetItems(data)
		log.Printf("Done SetItems")
		return result
	})

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

	fmt.Fprintf(IssueView, "%s", issue.Fields.Description)
	return nil
}

func ListDown(g *c.Gui, v *c.View) error {
	switch v.Name() {

	case SIDE_VIEW:
		if err := QueryList.MoveDown(); err != nil {
			log.Println("Error on SitesList.MoveDown()", err)
			return err
		}
	case LIST_VIEW:
		if err := IssueList.MoveDown(); err != nil {
			log.Println("Error on NewsList.MoveDown()", err)
			return err
		}
		if err := UpdateIssue(); err != nil {
			log.Println("Error on UpdateSummary()", err)
			return err
		}
	}
	return nil
}

func ListPgDown(g *c.Gui, v *c.View) error {
	switch v.Name() {

	case SIDE_VIEW:
		if err := QueryList.MovePgDown(); err != nil {
			log.Println("Error on SitesList.MovePgDown()", err)
			return err
		}
	case LIST_VIEW:
		if err := IssueList.MovePgDown(); err != nil {
			log.Println("Error on NewsList.MovePgDown()", err)
			return err
		}
		if err := UpdateIssue(); err != nil {
			log.Println("Error on UpdateSummary()", err)
			return err
		}
	}
	return nil
}

func ListPgUp(g *c.Gui, v *c.View) error {
	switch v.Name() {

	case SIDE_VIEW:
		if err := QueryList.MovePgUp(); err != nil {
			log.Println("Error on SitesList.MovePgUp()", err)
			return err
		}
	case LIST_VIEW:
		if err := IssueList.MovePgUp(); err != nil {
			log.Println("Error on NewsList.MovePgUp()", err)
			return err
		}
		if err := UpdateIssue(); err != nil {
			log.Println("Error on UpdateSummary()", err)
			return err
		}
	}
	return nil
}
