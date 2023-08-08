package context

import (
	"strings"

	"github.com/jesseduffield/lazygit/pkg/gui/keybindings"
	"github.com/jesseduffield/lazygit/pkg/gui/style"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/samber/lo"
)

type MenuContext struct {
	c *ContextCommon

	*MenuViewModel
	*ListContextTrait
}

var _ types.IListContext = (*MenuContext)(nil)

func NewMenuContext(
	c *ContextCommon,
) *MenuContext {
	viewModel := NewMenuViewModel(c)

	return &MenuContext{
		c:             c,
		MenuViewModel: viewModel,
		ListContextTrait: &ListContextTrait{
			Context: NewSimpleContext(NewBaseContext(NewBaseContextOpts{
				View:                  c.Views().Menu,
				WindowName:            "menu",
				Key:                   "menu",
				Kind:                  types.TEMPORARY_POPUP,
				Focusable:             true,
				HasUncontrolledBounds: true,
			})),
			ListRenderer: ListRenderer{
				list:                viewModel,
				getDisplayStrings:   viewModel.GetDisplayStrings,
				getColumnAlignments: func() []utils.Alignment { return viewModel.columnAlignment },
				getNonModelItems:    viewModel.GetNonModelItems,
				renderNonModelItem:  viewModel.RenderNonModelItem,
			},
			c: c,
		},
	}
}

// TODO: remove this thing.
func (self *MenuContext) GetSelectedItemId() string {
	item := self.GetSelected()
	if item == nil {
		return ""
	}

	return item.Label
}

type MenuViewModel struct {
	c               *ContextCommon
	menuItems       []*types.MenuItem
	columnAlignment []utils.Alignment
	*FilteredListViewModel[*types.MenuItem]
}

func NewMenuViewModel(c *ContextCommon) *MenuViewModel {
	self := &MenuViewModel{
		menuItems: nil,
		c:         c,
	}

	self.FilteredListViewModel = NewFilteredListViewModel(
		func() []*types.MenuItem { return self.menuItems },
		func(item *types.MenuItem) []string { return item.LabelColumns },
	)

	return self
}

func (self *MenuViewModel) SetMenuItems(items []*types.MenuItem, columnAlignment []utils.Alignment) {
	self.menuItems = items
	self.columnAlignment = columnAlignment
}

// TODO: move into presentation package
func (self *MenuViewModel) GetDisplayStrings(_ int, _ int) [][]string {
	menuItems := self.FilteredListViewModel.GetItems()
	showKeys := lo.SomeBy(menuItems, func(item *types.MenuItem) bool {
		return item.Key != nil
	})

	return lo.Map(menuItems, func(item *types.MenuItem, _ int) []string {
		displayStrings := item.LabelColumns

		if !showKeys {
			return displayStrings
		}

		// These keys are used for general navigation so we'll strike them out to
		// avoid confusion
		reservedKeys := []string{
			self.c.UserConfig.Keybinding.Universal.Confirm,
			self.c.UserConfig.Keybinding.Universal.Select,
			self.c.UserConfig.Keybinding.Universal.Return,
			self.c.UserConfig.Keybinding.Universal.StartSearch,
		}
		keyLabel := keybindings.LabelFromKey(item.Key)
		keyStyle := style.FgCyan
		if lo.Contains(reservedKeys, keyLabel) {
			keyStyle = style.FgDefault.SetStrikethrough()
		}

		displayStrings = utils.Prepend(displayStrings, keyStyle.Sprint(keyLabel))
		return displayStrings
	})
}

type MenuSectionRenderInfo struct {
	text   string
	column int
}

func (self *MenuViewModel) GetNonModelItems() []*NonModelItem {
	// Don't display section headers when we are filtering. The reason is that
	// filtering changes the order of the items (they are sorted by best match),
	// so all the sections would be messed up.
	if self.FilteredListViewModel.IsFiltering() {
		return []*NonModelItem{}
	}

	result := []*NonModelItem{}
	menuItems := self.FilteredListViewModel.GetItems()
	var prevSection *types.MenuSection = nil
	for i, menuItem := range menuItems {
		if menuItem.Section != nil && menuItem.Section != prevSection {
			if prevSection != nil {
				result = append(result, &NonModelItem{
					Index:      i,
					ClientData: MenuSectionRenderInfo{},
				})
			}

			result = append(result, &NonModelItem{
				Index: i,
				ClientData: MenuSectionRenderInfo{
					text:   menuItem.Section.Title,
					column: menuItem.Section.Column,
				},
			})
			prevSection = menuItem.Section
		}
	}

	return result
}

func (self *MenuViewModel) RenderNonModelItem(item *NonModelItem, columnPositions []int) string {
	renderInfo := item.ClientData.(MenuSectionRenderInfo)

	if renderInfo.text == "" {
		return ""
	}

	padding := strings.Repeat(" ", columnPositions[renderInfo.column])
	return padding + style.FgGreen.SetBold().Sprintf("--- %s ---", renderInfo.text)
}

func (self *MenuContext) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	basicBindings := self.ListContextTrait.GetKeybindings(opts)
	menuItemsWithKeys := lo.Filter(self.menuItems, func(item *types.MenuItem, _ int) bool {
		return item.Key != nil
	})

	menuItemBindings := lo.Map(menuItemsWithKeys, func(item *types.MenuItem, _ int) *types.Binding {
		return &types.Binding{
			Key:     item.Key,
			Handler: func() error { return self.OnMenuPress(item) },
		}
	})

	// appending because that means the menu item bindings have lower precedence.
	// So if a basic binding is to escape from the menu, we want that to still be
	// what happens when you press escape. This matters when we're showing the menu
	// for all keybindings of say the files context.
	return append(basicBindings, menuItemBindings...)
}

func (self *MenuContext) OnMenuPress(selectedItem *types.MenuItem) error {
	if err := self.c.PopContext(); err != nil {
		return err
	}

	if selectedItem == nil {
		return nil
	}

	if err := selectedItem.OnPress(); err != nil {
		return err
	}

	return nil
}
