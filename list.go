package main

import (
	"bytes"
	"fmt"
	"github.com/andygrunwald/go-jira"
	//"github.com/fatih/color"
	c "github.com/jroimartin/gocui"
	. "github.com/logrusorgru/aurora"
	"log"
)

// Page is used to hold info about a list based view
type Page struct {
	offset, limit int
}

// List overlads the gocui.View by implementing list specific functionalitys
type List struct {
	*c.View
	title       string
	items       []interface{}
	pages       []Page
	currPageIdx int
	ordered     bool
}

type IssueListType struct {
	List
}

// CreateList initializes a List object with an existing View by applying some
// basic configuration
func CreateList(v *c.View, ordered bool) *List {
	list := &List{}
	list.View = v
	list.SelBgColor = c.ColorBlack
	list.SelFgColor = c.ColorWhite | c.AttrBold
	list.Autoscroll = true
	list.ordered = ordered

	return list
}

func CreateIssueList(v *c.View, ordered bool) *IssueListType {
	list := &IssueListType{}
	list.View = v
	list.SelBgColor = c.ColorBlack
	list.SelFgColor = c.ColorWhite | c.AttrBold
	list.Autoscroll = true
	list.ordered = ordered
	return list
}

// IsEMpty indicates whether a list has items or not
func (l *List) IsEmpty() bool {
	return l.length() == 0
}

// Focus hightlights the View of the current List
func (l *List) Focus(g *c.Gui) error {
	l.Highlight = true
	_, err := g.SetCurrentView(l.Name())

	return err
}

// Unfocus is used to remove highlighting from the current list
func (l *List) Unfocus() {
	l.Highlight = false
}

// Reset zeros the list's slices out and clears the underlying View
func (l *List) Reset() {
	l.items = make([]interface{}, 0)
	l.pages = []Page{}
	l.Clear()
	l.ResetCursor()
}

// SetTitle will set the title of the View and display paging information of the
// list if there are more than one pages
func (l *List) SetTitle(title string) {
	l.title = title

	if l.pagesNum() > 1 {
		l.Title = fmt.Sprintf(" %d/%d - %v ", l.currPageNum(), l.pagesNum(), title)
	} else {
		l.Title = fmt.Sprintf(" %v ", title)
	}
}

// SetItems will (re)evaluates the list's items with the given data and redraws
// the View
func (l *List) SetItems(data []interface{}) error {
	log.Printf("SetItems %T", l)
	l.items = data
	l.ResetPages()
	return l.Draw()
}

// AddItem appends a given item to the existing list
func (l *List) AddItem(g *c.Gui, item interface{}) error {
	l.items = append(l.items, item)
	l.ResetPages()
	return l.Draw()
}

func (l *List) UpdateCurrentItem(item interface{}) {
	page := l.currPage()
	data := l.items[page.offset : page.offset+page.limit]

	data[l.currentCursorY()] = item
}

// Draw calculates the pages and draws the first one
func (l *List) Draw() error {
	if l.IsEmpty() {
		return nil
	}
	return l.displayPage(0)
}

// Draw calculates the pages and draws the first one
func (l *List) DrawCurrentPage() error {
	if l.IsEmpty() {
		return nil
	}
	return l.displayPage(l.currPageIdx)
}

// MoveDown moves the cursor to the line below or the next page if any
func (l *List) MoveDown() error {
	if l.IsEmpty() {
		return nil
	}
	log.Println(l.currentCursorY())
	y := l.currentCursorY() + 1
	if l.atBottomOfPage() {
		y = 0
		if l.hasMultiplePages() {
			l.displayPage(l.nextPageIdx())
		}
	}
	return l.SetCursor(0, y)
}

// MoveUp moves the cursor to the line above on the previous page if any
func (l *List) MoveUp() error {
	if l.IsEmpty() {
		return nil
	}
	y := l.currentCursorY() - 1
	if l.atTopOfPage() {
		y = l.pages[l.prevPageIdx()].limit - 1
		if l.hasMultiplePages() {
			l.displayPage(l.prevPageIdx())
		}
	}

	return l.SetCursor(0, y)
}

// MovePgDown displays the next page
func (l *List) MovePgDown() error {
	if l.IsEmpty() {
		return nil
	}
	l.displayPage(l.nextPageIdx())

	return l.SetCursor(0, 0)
}

// MovePgUp displays the previous page
func (l *List) MovePgUp() error {
	if l.IsEmpty() {
		return nil
	}
	l.displayPage(l.prevPageIdx())

	return l.SetCursor(0, 0)
}

// CurrentItem returns the currently selected item of the list no matter what
// page is being displayed
func (l *List) CurrentItem() interface{} {
	if l.IsEmpty() {
		return nil
	}
	page := l.currPage()
	data := l.items[page.offset : page.offset+page.limit]

	return data[l.currentCursorY()]
}

// ResetCursor puts the cirson back at the beginning of the View
func (l *List) ResetCursor() {
	l.SetCursor(0, 0)
}

// ResetPages (re)calculates the pages data based on the current length of the
// list and the current height of the View
func (l *List) ResetPages() {
	l.pages = []Page{}
	for offset := 0; offset < l.length(); offset += l.height() {
		limit := l.height()
		if offset+limit > l.length() {
			limit = l.length() % l.height()
		}
		l.pages = append(l.pages, Page{offset, limit})
	}
}

// currPageNum returns the current page number to display
func (l *List) currPageNum() int {
	if l.IsEmpty() {
		return 0
	}
	return l.currPageIdx + 1
}

// currentCursorY returns the current Y of the cursor
func (l *List) currentCursorY() int {
	_, y := l.Cursor()

	return y
}

// currPage returns the current page being displayd
func (l *List) currPage() Page {
	return l.pages[l.currPageIdx]
}

// height ewturns the current height of the View
func (l *List) height() int {
	_, y := l.Size()

	return y - 1
}

// width ewturns the current width of the View
func (l *List) width() int {
	x, _ := l.Size()

	return x - 1
}

// length returns the length of the list
func (l *List) length() int {
	return len(l.items)
}

// pageNum returns the number of the pages
func (l *List) pagesNum() int {
	return len(l.pages)
}

// nextPageIdx returns the index of the next page to be displayed circularlt
func (l *List) nextPageIdx() int {
	return (l.currPageIdx + 1) % l.pagesNum()
}

// prevPageIdx returns the index of the prev page to be displayed circularlt
func (l *List) prevPageIdx() int {
	pidx := (l.currPageIdx - 1) % l.pagesNum()
	if l.currPageIdx == 0 {
		pidx = l.pagesNum() - 1
	}
	return pidx
}

// sidplayItem displays the text of the item with index i and fills with spaces
// the remaining space until the border of the View
func (l *List) displayItem(i int) string {
	item := l.items[i]
	var text string
	// cyan := color.New(color.FgCyan).SprintfFunc()
	// green := color.New(color.FgGreen).SprintfFunc()
	switch item.(type) {
	case jira.Issue:
		issue := item.(jira.Issue)
		summary := issue.Fields.Summary
		assignee := issue.Fields.Assignee
		var username string
		if assignee != nil {
			username = assignee.DisplayName

		} else {
			username = "unassigned"
		}
		if len(summary) > 70 {
			summary = summary[:70]
		}
		// issueField := cyan("%-9v", issue.Key)
		// statusField := green("%12v", issue.Fields.Status.Name)
		text = fmt.Sprintf("%-8s| %-70s|%12s | %s", Cyan(issue.Key), summary, Green(issue.Fields.Status.Name), username)

		//text = fmt.Sprintf("|%s| %60s | %10s | %20s", cyan("%9v", issue.Key), summary, green.Sprint(issue.Fields.Status.Name), username)
	default:
		text = fmt.Sprint(item)
	}
	sp := spaces(l.width() - len(text) - 3)
	if l.ordered {
		return fmt.Sprintf("%2d. %v%v", i+1, text, sp)
	} else {
		return fmt.Sprintf(" %v%v", text, sp)
	}
}

// displayPage resets the currentPageIdx and displays the requested page
func (l *List) displayPage(p int) error {
	l.Clear()
	l.currPageIdx = p
	page := l.pages[l.currPageIdx]
	for i := page.offset; i < page.offset+page.limit; i++ {
		if _, err := fmt.Fprintln(l.View, l.displayItem(i)); err != nil {
			return err
		}
	}
	l.SetTitle(l.title)

	return nil
}

// atBottomOfPage determines whether the cursor is at the bottom of the current page
func (l *List) atBottomOfPage() bool {
	return l.currentCursorY() == l.currPage().limit-1
}

// atTopOfPage determines whether the cursor is at the top of the current page
func (l *List) atTopOfPage() bool {
	return l.currentCursorY() == 0
}

// hasMultiplePages determines whether there is more than one page to be displayed
func (l *List) hasMultiplePages() bool {
	return l.pagesNum() > 1
}

func spaces(n int) string {
	var s bytes.Buffer
	for i := 0; i < n; i++ {
		s.WriteString(" ")
	}
	return s.String()
}
