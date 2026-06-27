package screens

import (
	"docker-cli/tui/common"
	"docker-cli/tui/widgets"
	"strings"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ChatSessionModel struct {
	ta          textarea.Model
	viewPort    viewport.Model
	width       int
	height      int
	messages    []string
	senderStyle lipgloss.Style
	agentStyle  lipgloss.Style
	showMenu    bool
	optionsMenu widgets.OptionsModel
}

func NewChatSessionModel() *ChatSessionModel {
	var cs ChatSessionModel
	cs.optionsMenu = widgets.NewOptionsModel()
	cs.senderStyle = common.FGreenish
	cs.agentStyle = common.FMobyBlue
	ta := textarea.New()
	ta.Placeholder = "send a message...."
	ta.SetVirtualCursor(false)
	ta.Focus()

	ta.SetWidth(30)
	ta.SetHeight(5)
	ta.Prompt = common.FWhiteBlue.Render("┃ ")
	ta.CharLimit = 500 // need to be adjusted based on the maximum number of tokens the user set no ?
	s := ta.Styles()
	s.Focused.CursorLine = lipgloss.NewStyle()

	vp := viewport.New(viewport.WithWidth(50), viewport.WithHeight(5))
	aiMessage := common.BGray.Render(common.AdjustTextForBackground(
		`Welcome To Docker AI, how can I assist you today ?`,
		50,
		5))
	vp.SetContent(cs.agentStyle.Render("Agent: ") + aiMessage)
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)

	ta.SetStyles(s)

	ta.KeyMap.InsertNewline.SetEnabled(false)
	cs.ta = ta
	cs.viewPort = vp
	return &cs
}

func (c ChatSessionModel) Init() tea.Cmd {
	return textarea.Blink
}

func (c ChatSessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.ta.SetWidth(msg.Width)
		c.viewPort.SetHeight(msg.Height - c.ta.Height() - lipgloss.Height(common.OptionsLine))
		c.viewPort.SetWidth(msg.Width / 2)
		if len(c.messages) > 0 {
			// Wrap content before setting it.
			c.viewPort.SetContent(lipgloss.NewStyle().Width(c.viewPort.Width()).Render(strings.Join(c.messages, "\n")))
		}
		c.viewPort.GotoBottom()
	case tea.KeyPressMsg:
		if c.showMenu {
			switch msg.String() {
			case "esc":
				c.showMenu = false
				return c, nil
			default:
				var cmd tea.Cmd
				c.optionsMenu, cmd = c.optionsMenu.Update(msg)
				return c, cmd
			}
		}

		switch msg.String() {
		case "enter":
			if c.ta.Value() == "" {
				return c, nil
			}
			c.messages = append(c.messages, c.senderStyle.Render("You: ")+c.ta.Value())
			c.viewPort.SetContent(lipgloss.NewStyle().Width(c.viewPort.Width()).Render(strings.Join(c.messages, "\n")))
			c.ta.Reset()
			c.viewPort.GotoBottom()
			return c, nil
		case "ctrl+o":
			c.showMenu = true
			return c, nil
		default:
			var cmd tea.Cmd
			c.ta, cmd = c.ta.Update(msg)
			return c, cmd
		}
	case cursor.BlinkMsg:
		// Textarea should also process cursor blinks.
		var cmd tea.Cmd
		c.ta, cmd = c.ta.Update(msg)
		return c, cmd
	}

	return c, nil
}

func (c ChatSessionModel) View() tea.View {
	cur := c.ta.Cursor()
	if cur != nil {
		cur.Y += lipgloss.Height(c.viewPort.View())
	}

	// border := lipgloss.NewStyle().Border(lipgloss.BlockBorder()).Render(c.viewPort.View())

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		c.viewPort.View(),
		c.ta.View(),
		common.RenderStatusLineBorder(c.width, "opus 4.7"),
	)

	if c.showMenu {
		listView := c.optionsMenu.View()
		overlayContent := listView.Content
		overlayLayer := lipgloss.NewLayer(overlayContent).
			X((c.width - lipgloss.Width(overlayContent)) / 2).
			Y((c.height - lipgloss.Height(overlayContent)) / 2).
			Z(1)
		baseLayer := lipgloss.NewLayer(content)
		content = lipgloss.NewCompositor(baseLayer, overlayLayer).Render()
	}

	v := tea.NewView(lipgloss.Place(
		c.width,
		c.height,
		lipgloss.Left,
		lipgloss.Top,
		content,
	))

	v.AltScreen = true
	v.Cursor = cur
	return v
}
