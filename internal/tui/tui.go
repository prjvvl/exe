// Package tui is the interactive menu shown when `exe` is run with no
// arguments (or `exe <app>` to jump straight into one app's commands).
//
// Whichever item is picked, it's resolved the exact same way the direct CLI
// path does (internal/dispatch.Emit): the program quits, then the caller
// hands the result to dispatch.Emit. The TUI never executes anything itself.
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/prjvvl/exe/internal/builtin"
	"github.com/prjvvl/exe/internal/config"
	"github.com/prjvvl/exe/internal/dispatch"
	"github.com/prjvvl/exe/internal/registry"
)

// Same palette as docs/index.html, so the TUI and the landing page read as
// one product rather than two different tools.
var (
	colorBg      = lipgloss.Color("#0b0e14")
	colorBorder  = lipgloss.Color("#22283a")
	colorText    = lipgloss.Color("#e6e9f0")
	colorTextDim = lipgloss.Color("#8b93a7")
	colorAccent  = lipgloss.Color("#5eead4")

	appStyle = lipgloss.NewStyle().Margin(0, 1)

	titleStyle = lipgloss.NewStyle().
			Foreground(colorBg).
			Background(colorAccent).
			Bold(true).
			Padding(0, 1)

	selectKey = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select"))
	backKey   = key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc/backspace", "back"))
	quitKey   = key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit"))
)

type itemKind int

const (
	kindBuiltin itemKind = iota
	kindApp
	kindCommand
)

type menuItem struct {
	kind        itemKind
	name        string
	description string
	run         string
	builtin     *builtin.Builtin
	app         *config.App
}

func (i menuItem) Title() string { return i.name }

func (i menuItem) Description() string {
	if i.kind == kindCommand {
		return fmt.Sprintf("%s (runs: %s)", i.description, i.run)
	}
	return i.description
}

func (i menuItem) FilterValue() string { return i.name + " " + i.description }

type screen struct {
	list list.Model
}

type model struct {
	reg       registry.Registry
	screens   []screen
	result    dispatch.Resolved
	resultSet bool
	err       error
}

func newList(items []list.Item, title string) list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(colorText)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(colorTextDim)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(colorAccent).
		BorderForeground(colorAccent)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(colorTextDim).
		BorderForeground(colorAccent)

	l := list.New(items, delegate, 0, 0)
	l.Title = title
	l.Styles.Title = titleStyle
	l.Styles.TitleBar = l.Styles.TitleBar.Padding(0, 0, 0, 1)
	l.Styles.PaginationStyle = l.Styles.PaginationStyle.Foreground(colorTextDim).PaddingLeft(1)
	l.Styles.HelpStyle = l.Styles.HelpStyle.Foreground(colorTextDim).Padding(0, 0, 0, 1)
	l.Help.Styles.ShortKey = l.Help.Styles.ShortKey.Foreground(colorAccent)
	l.Help.Styles.ShortDesc = l.Help.Styles.ShortDesc.Foreground(colorTextDim)
	l.Help.Styles.ShortSeparator = l.Help.Styles.ShortSeparator.Foreground(colorBorder)
	l.Help.Styles.FullKey = l.Help.Styles.FullKey.Foreground(colorAccent)
	l.Help.Styles.FullDesc = l.Help.Styles.FullDesc.Foreground(colorTextDim)
	l.Help.Styles.FullSeparator = l.Help.Styles.FullSeparator.Foreground(colorBorder)
	l.SetShowStatusBar(false)
	l.AdditionalShortHelpKeys = func() []key.Binding { return []key.Binding{selectKey, backKey} }
	l.AdditionalFullHelpKeys = func() []key.Binding { return []key.Binding{selectKey, backKey} }
	return l
}

func buildTopItems(reg registry.Registry) []list.Item {
	items := make([]list.Item, 0, len(reg.Builtins)+len(reg.Config.Apps))
	for i := range reg.Builtins {
		b := reg.Builtins[i]
		items = append(items, menuItem{kind: kindBuiltin, name: b.Name, description: b.Description, builtin: &b})
	}
	for i := range reg.Config.Apps {
		a := reg.Config.Apps[i]
		items = append(items, menuItem{kind: kindApp, name: a.Name, description: a.Description, app: &a})
	}
	return items
}

func buildCommandItems(app config.App) []list.Item {
	items := make([]list.Item, 0, len(app.Commands))
	for _, c := range app.Commands {
		items = append(items, menuItem{kind: kindCommand, name: c.Name, description: c.Description, run: c.Run})
	}
	return items
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) currentList() list.Model {
	return m.screens[len(m.screens)-1].list
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		for i := range m.screens {
			m.screens[i].list.SetSize(msg.Width-h, msg.Height-v)
		}
		return m, nil

	case tea.KeyMsg:
		// Only claim q/esc/backspace/enter when not filtering, otherwise
		// they need to reach the list as ordinary typed/editing keys.
		if m.currentList().FilterState() == list.Unfiltered {
			switch {
			case key.Matches(msg, quitKey):
				return m, tea.Quit
			case key.Matches(msg, backKey):
				if len(m.screens) > 1 {
					m.screens = m.screens[:len(m.screens)-1]
					return m, nil
				}
				return m, tea.Quit
			case key.Matches(msg, selectKey):
				return m.handleSelect()
			}
		}
	}

	var cmd tea.Cmd
	idx := len(m.screens) - 1
	m.screens[idx].list, cmd = m.screens[idx].list.Update(msg)
	return m, cmd
}

func (m model) handleSelect() (tea.Model, tea.Cmd) {
	item, ok := m.currentList().SelectedItem().(menuItem)
	if !ok {
		return m, nil
	}

	switch item.kind {
	case kindBuiltin:
		r, err := item.builtin.Handler(m.reg.Config, nil)
		if err != nil {
			m.err = err
			return m, tea.Quit
		}
		m.result = r
		m.resultSet = true
		return m, tea.Quit

	case kindApp:
		if len(item.app.Commands) == 1 {
			m.result = dispatch.Resolved{Kind: dispatch.KindEval, Text: item.app.Commands[0].Run}
			m.resultSet = true
			return m, tea.Quit
		}
		w, h := m.currentList().Width(), m.currentList().Height()
		next := newList(buildCommandItems(*item.app), fmt.Sprintf("exe > %s", item.app.Name))
		next.SetSize(w, h)
		m.screens = append(m.screens, screen{list: next})
		return m, nil

	case kindCommand:
		m.result = dispatch.Resolved{Kind: dispatch.KindEval, Text: item.run}
		m.resultSet = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	return appStyle.Render(m.currentList().View())
}

func runProgram(initial model) error {
	p := tea.NewProgram(initial, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}

	fm := final.(model)
	if fm.err != nil {
		return fm.err
	}
	if !fm.resultSet {
		return nil
	}
	return dispatch.Emit(fm.result)
}

func Run(reg registry.Registry) error {
	l := newList(buildTopItems(reg), "exe")
	return runProgram(model{reg: reg, screens: []screen{{list: l}}})
}

func RunApp(app config.App) error {
	l := newList(buildCommandItems(app), fmt.Sprintf("exe > %s", app.Name))
	return runProgram(model{screens: []screen{{list: l}}})
}
