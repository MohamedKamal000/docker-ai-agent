package widgets

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1D63ED")).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(1, 2).
			Width(30)

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
		item{title: "Change Model", desc: "In other words, towel fabric"},
		item{title: "Switch Session", desc: "In other words, towel fabric"},
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
	return opt
}

func (o OptionsModel) Init() tea.Cmd {
	return nil
}

func (o OptionsModel) Update(msg tea.Msg) (OptionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			return o, nil
		}
	}
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
