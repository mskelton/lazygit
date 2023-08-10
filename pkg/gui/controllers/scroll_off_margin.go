package controllers

import (
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

// To be called after pressing up-arrow; checks whether the cursor entered the
// top scroll-off margin, and so the view needs to be scrolled up one line
func checkScrollUp(view types.IViewTrait, scrollOffMargin int, lineIdxBefore int, lineIdxAfter int) {
	startIdx, length := view.ViewPortYBounds()
	// scroll only if the view is tall enough to have room for both the top and
	// bottom scroll-off margins...
	if length > scrollOffMargin*2 &&
		// ... and the "before" position was visible (this could be false if the
		// scroll wheel was used to scroll the selected line out of view)
		lineIdxBefore >= startIdx && lineIdxBefore < startIdx+length {
		marginEnd := startIdx + scrollOffMargin
		// ... and the "after" position is within the top margin (or before it)
		if lineIdxAfter < marginEnd {
			view.ScrollUp(marginEnd - lineIdxAfter)
		}
	}
}

// To be called after pressing down-arrow; checks whether the cursor entered the
// bottom scroll-off margin, and so the view needs to be scrolled down one line
func checkScrollDown(view types.IViewTrait, scrollOffMargin int, lineIdxBefore int, lineIdxAfter int) {
	startIdx, length := view.ViewPortYBounds()
	// scroll only if the view is tall enough to have room for both the top and
	// bottom scroll-off margins...
	if length > scrollOffMargin*2 &&
		// ... and the "before" position was visible (this could be false if the
		// scroll wheel was used to scroll the selected line out of view)
		lineIdxBefore >= startIdx && lineIdxBefore < startIdx+length {
		marginStart := startIdx + length - scrollOffMargin - 1
		// ... and the "after" position is within the bottom margin (or after it)
		if lineIdxAfter > marginStart {
			view.ScrollDown(lineIdxAfter - marginStart)
		}
	}
}
