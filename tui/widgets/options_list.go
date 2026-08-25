package widgets

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	popupWidth     = 40
	maxVisibleRows = 10
)

var (
	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1D63ED")).
			Width(popupWidth)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1D63ED")).
			Bold(true)

	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#4EC7C4"))
	deselectedStyle = lipgloss.NewStyle()
	noMatchesStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// Option is a single entry in the options menu. Items with children open a
// sub-page inside the same widget; leaves emit an ItemSelectedMsg.
type Option struct {
	Title    string
	Children []Option
}

type OptionsPage struct {
	Title string
	Items []Option
}

// ItemSelectedMsg is emitted when a leaf option is selected. Path holds the
// titles from root to the chosen leaf.
type ItemSelectedMsg struct {
	Path []string
}

// OptionsCloseMsg is emitted when the user dismisses the menu at its root
// (esc/q); the host screen should close the overlay.
type OptionsCloseMsg struct{}

type OptionsModel struct {
	pages  []OptionsPage // navigation stack; last element is displayed
	query  textinput.Model
	cursor int
	width  int
	height int
}

func NewOptionsModel() OptionsModel {
	return newOptionsModel(OptionsPage{
		Title: "Options",
		Items: []Option{
			{Title: "Change Model", Children: placeholderModels()},
			{Title: "Switch Session", Children: placeholderSessions()},
		},
	})
}

func newOptionsModel(root OptionsPage) OptionsModel {
	o := OptionsModel{
		pages: []OptionsPage{root},
	}
	o.query = textinput.New()
	o.query.Placeholder = "search..."
	o.query.Prompt = "> "
	o.query.CharLimit = 64
	o.query.Focus()
	v := o.View().Content
	o.width = lipgloss.Width(v)
	o.height = lipgloss.Height(v)
	return o
}

// placeholder data until real sources are wired in.
func placeholderModels() []Option {
	return []Option{
		{Title: "gpt-4o"},
		{Title: "gpt-4o-mini"},
		{Title: "claude-sonnet"},
	}
}

func placeholderSessions() []Option {
	return []Option{
		{Title: "session-1"},
		{Title: "session-2"},
	}
}

func (o OptionsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (o *OptionsModel) currentPage() OptionsPage {
	return o.pages[len(o.pages)-1]
}

// filteredItems returns the current page's items matching the query.
func (o OptionsModel) filteredItems() []Option {
	items := o.currentPage().Items
	q := strings.ToLower(strings.TrimSpace(o.query.Value()))
	if q == "" {
		return items
	}
	var out []Option
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Title), q) {
			out = append(out, it)
		}
	}
	return out
}

func (o *OptionsModel) Update(msg tea.Msg) (OptionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if len(o.pages) > 1 {
				o.pages = o.pages[:len(o.pages)-1]
				o.resetQueryAndCursor()
				return *o, nil
			}
			return *o, func() tea.Msg { return OptionsCloseMsg{} }
		case "q":
			return *o, func() tea.Msg { return OptionsCloseMsg{} }
		case "up":
			if o.cursor > 0 {
				o.cursor--
			}
			return *o, nil
		case "down":
			if o.cursor < len(o.filteredItems())-1 {
				o.cursor++
			}
			return *o, nil
		case "enter":
			filtered := o.filteredItems()
			if len(filtered) == 0 || o.cursor >= len(filtered) {
				return *o, nil
			}
			chosen := filtered[o.cursor]
			path := append(o.pathTitles(), chosen.Title)
			if len(chosen.Children) > 0 {
				o.pages = append(o.pages, OptionsPage{Title: chosen.Title, Items: chosen.Children})
				o.resetQueryAndCursor()
				return *o, nil
			}
			return *o, tea.Batch(
				func() tea.Msg { return ItemSelectedMsg{Path: path} },
				func() tea.Msg { return OptionsCloseMsg{} }, // close the menu after a leaf selection
			)
		}
	}

	var cmd tea.Cmd
	var m OptionsModel = *o
	m.query, cmd = m.query.Update(msg)
	if m.query.Value() != o.query.Value() {
		m.clampCursor()
	}
	*o = m
	return *o, cmd
}

// pathTitles collects the page titles from root to the current page.
func (o OptionsModel) pathTitles() []string {
	path := make([]string, 0, len(o.pages))
	for _, p := range o.pages {
		path = append(path, p.Title)
	}
	if len(path) > 0 {
		path = path[1:] // drop the generic root title
	}
	return path
}

func (o *OptionsModel) resetQueryAndCursor() {
	o.query.SetValue("")
	o.cursor = 0
}

func (o *OptionsModel) clampCursor() {
	n := len(o.filteredItems())
	if n == 0 {
		o.cursor = 0
	} else if o.cursor >= n {
		o.cursor = n - 1
	}
}

func (o OptionsModel) View() tea.View {
	var b strings.Builder

	b.WriteString(titleStyle.Render(o.currentPage().Title))
	b.WriteString("\n")
	b.WriteString(o.query.View())
	b.WriteString("\n\n")

	filtered := o.filteredItems()
	if len(filtered) == 0 {
		b.WriteString(noMatchesStyle.Render("no matches"))
		return tea.NewView(popupStyle.Render(b.String()))
	}

	// Scroll window around the cursor when items exceed maxVisibleRows.
	start := 0
	if len(filtered) > maxVisibleRows && o.cursor > maxVisibleRows/2 {
		start = min(o.cursor-maxVisibleRows/2, len(filtered)-maxVisibleRows)
	}
	end := min(start+maxVisibleRows, len(filtered))

	for i := start; i < end; i++ {
		style := deselectedStyle
		prefix := "  "
		if i == o.cursor {
			style = selectedStyle
			prefix = "> "
		}
		marker := ""
		if len(filtered[i].Children) > 0 {
			marker = " >"
		}
		b.WriteString(style.Render(prefix + filtered[i].Title + marker))
		b.WriteString("\n")
	}

	if len(filtered) > maxVisibleRows {
		b.WriteString(noMatchesStyle.Render(fmt.Sprintf("%d/%d", o.cursor+1, len(filtered))))
		b.WriteString("\n")
	}

	return tea.NewView(popupStyle.Render(b.String()))
}
