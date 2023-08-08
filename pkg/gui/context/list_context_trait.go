package context

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"golang.org/x/exp/slices"
)

type NonModelItem struct {
	Index int    // Where in the model this should be inserted
	Text  string // The text to draw (including styling)
}

type ListContextTrait struct {
	types.Context

	c                 *ContextCommon
	list              types.IList
	getDisplayStrings func(startIdx int, length int) [][]string
	// Alignment for each column. If nil, the default is left alignment
	getColumnAlignments func() []utils.Alignment
	// Function to insert non-model items (e.g. section headers). If nil, no
	// such items are inserted
	// The column positions of the display strings are passed in so that it's
	// possible to align section headers with columns
	getNonModelItems func(columnPositions []int) []NonModelItem
	// Some contexts, like the commit context, will highlight the path from the selected commit
	// to its parents, because it's ambiguous otherwise. For these, we need to refresh the viewport
	// so that we show the highlighted path.
	// TODO: now that we allow scrolling, we should be smarter about what gets refreshed:
	// we should find out exactly which lines are now part of the path and refresh those.
	// We should also keep track of the previous path and refresh those lines too.
	refreshViewportOnChange bool

	// Indices of things that we display in addition to model items. These could
	// be section headers, divider lines, or other things. They are created with
	// the getNonModelItems func of ListContextTrait. Note that unlike the
	// NonModelItem.Index field, this is a view index.
	nonModelItemIndices []int
}

func (self *ListContextTrait) IsListContext() {}

func (self *ListContextTrait) GetList() types.IList {
	return self.list
}

func (self *ListContextTrait) GetNonModelItemIndices() []int {
	return self.nonModelItemIndices
}

func (self *ListContextTrait) ModelIndexToViewIndex(modelIndex int) int {
	if self.nonModelItemIndices != nil {
		for _, nonModelItemIndex := range self.nonModelItemIndices {
			if nonModelItemIndex > modelIndex {
				break
			}
			modelIndex++
		}
	}

	return modelIndex
}

func (self *ListContextTrait) ViewIndexToModelIndex(viewIndex int) int {
	if self.GetNonModelItemIndices() != nil {
		for _, nonModelItemIndex := range self.GetNonModelItemIndices() {
			if nonModelItemIndex > viewIndex {
				break
			}
			viewIndex--
		}
	}

	return viewIndex
}

func (self *ListContextTrait) FocusLine() {
	// Doing this at the end of the layout function because we need the view to be
	// resized before we focus the line, otherwise if we're in accordion mode
	// the view could be squashed and won't how to adjust the cursor/origin
	self.c.AfterLayout(func() error {
		self.GetViewTrait().FocusPoint(
			self.ModelIndexToViewIndex(self.list.GetSelectedLineIdx()))
		return nil
	})

	self.setFooter()

	if self.refreshViewportOnChange {
		self.refreshViewport()
	}
}

func (self *ListContextTrait) renderLines(startIdx int, length int) string {
	var columnAlignments []utils.Alignment
	if self.getColumnAlignments != nil {
		columnAlignments = self.getColumnAlignments()
	}
	lines, columnPositions := utils.RenderDisplayStrings(
		self.getDisplayStrings(startIdx, length),
		columnAlignments)
	if self.getNonModelItems != nil {
		nonModelItems := self.getNonModelItems(columnPositions)
		self.nonModelItemIndices = make([]int, 0, len(nonModelItems))
		offset := 0
		for _, item := range nonModelItems {
			viewIndex := item.Index + offset
			lines = slices.Insert(lines, viewIndex, item.Text)
			self.nonModelItemIndices = append(self.nonModelItemIndices, viewIndex)
			offset++
		}
	}
	return strings.Join(lines, "\n")
}

func (self *ListContextTrait) refreshViewport() {
	startIdx, length := self.GetViewTrait().ViewPortYBounds()
	content := self.renderLines(startIdx, length)
	self.GetViewTrait().SetViewPortContent(content)
}

func (self *ListContextTrait) setFooter() {
	self.GetViewTrait().SetFooter(formatListFooter(self.list.GetSelectedLineIdx(), self.list.Len()))
}

func formatListFooter(selectedLineIdx int, length int) string {
	return fmt.Sprintf("%d of %d", selectedLineIdx+1, length)
}

func (self *ListContextTrait) HandleFocus(opts types.OnFocusOpts) error {
	self.FocusLine()

	self.GetViewTrait().SetHighlight(self.list.Len() > 0)

	return self.Context.HandleFocus(opts)
}

func (self *ListContextTrait) HandleFocusLost(opts types.OnFocusLostOpts) error {
	self.GetViewTrait().SetOriginX(0)

	if self.refreshViewportOnChange {
		self.refreshViewport()
	}

	return self.Context.HandleFocusLost(opts)
}

// OnFocus assumes that the content of the context has already been rendered to the view. OnRender is the function which actually renders the content to the view
func (self *ListContextTrait) HandleRender() error {
	self.list.RefreshSelectedIdx()
	content := self.renderLines(0, self.list.Len())
	self.GetViewTrait().SetContent(content)
	self.c.Render()
	self.setFooter()

	return nil
}

func (self *ListContextTrait) OnSearchSelect(selectedLineIdx int) error {
	self.GetList().SetSelectedLineIdx(selectedLineIdx)
	return self.HandleFocus(types.OnFocusOpts{})
}
