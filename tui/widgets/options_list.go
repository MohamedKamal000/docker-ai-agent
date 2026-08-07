package widgets

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1D63ED")).
			Width(40)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1D63ED")).
			Bold(true).
			MarginBottom(1)
)

type OptionsModel struct {
	List     list.Model
	delegate list.DefaultDelegate
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func NewOptionsModel() OptionsModel {
	items := []list.Item{
		item{title: "Change Model", desc: ""},
		item{title: "Switch Session", desc: ""},
	}

	d := list.NewDefaultDelegate()
	d.SetSpacing(1)
	d.ShowDescription = false
	opt := OptionsModel{
		List:     list.New(items, d, 0, 0),
		delegate: d,
	}
	v := opt.View().Content
	opt.List.SetSize(lipgloss.Width(v), lipgloss.Height(v))
	opt.List.SetFilteringEnabled(true)
	opt.List.SetSize(50, 15)
	opt.List.KeyMap = CustomKeyMap()
	return opt
}

func (o OptionsModel) Init() tea.Cmd {
	return nil
}

func (o OptionsModel) Update(msg tea.Msg) (OptionsModel, tea.Cmd) {
	var cmd tea.Cmd
	o.List, cmd = o.List.Update(msg)
	return o, cmd
}

func (o OptionsModel) View() tea.View {
	header := titleStyle.Render("Options")
	body := o.List.View()
	rendered := popupStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, body))
	return tea.NewView(rendered)
}

func CustomKeyMap() list.KeyMap {
	return list.KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left", "h", "pgup", "b", "u"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "l", "pgdown", "f", "d"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys("home", "g"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end", "G"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear filter"),
		),

		// Filtering.
		CancelWhileFiltering: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "cancel filter"),
		),

		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys("enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"),
			key.WithHelp("enter", "apply filter"),
		),

		// Toggle help.
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),

		// Quitting.
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
}
