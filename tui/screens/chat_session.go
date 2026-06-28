package screens

import (
	"docker-cli/internal/core"
	"docker-cli/tui/common"
	"docker-cli/tui/widgets"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ChatSessionModel struct {
	ta             textarea.Model
	viewPort       viewport.Model
	spinner        spinner.Model
	width          int
	height         int
	messages       []string
	showMenu       bool
	agentIsRunning bool
	showWarning    bool
	optionsMenu    widgets.OptionsModel
}

func NewChatSessionModel() *ChatSessionModel {
	var cs ChatSessionModel
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	cs.spinner = sp
	cs.optionsMenu = widgets.NewOptionsModel()
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
	vp.SetContent(common.FMobyBlue.Render("Agent: ") + `Welcome To Docker AI, how can I assist you today ?`)
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)

	ta.SetStyles(s)

	ta.KeyMap.InsertNewline.SetEnabled(false)
	cs.ta = ta
	cs.viewPort = vp
	return &cs
}

func (c *ChatSessionModel) Init() tea.Cmd {
	return textarea.Blink
}

func (c *ChatSessionModel) sendUserMessage(message string) {
	c.messages = append(c.messages, common.RenderUserMessageWithBackground(message, c.viewPort.Width()))
	c.viewPort.SetContent(lipgloss.NewStyle().Width(c.viewPort.Width()).Render(strings.Join(c.messages, "\n")))
	c.ta.Reset()
	c.viewPort.GotoBottom()
}

func (c *ChatSessionModel) sendAiMessage(message string, responseType core.ResponseType) {
	switch responseType {
	case core.FinalResponse:
		message = common.FMobyBlue.Render("Agent: ") + message
		c.agentIsRunning = false
	case core.Thoughts:
		message = common.FMobyBlue.Render("Thoughts: ") + message
	case core.Warning:
		message = common.RenderWarningMessage(message, c.viewPort.Width())
		c.showWarning = true
	}
	c.messages = append(c.messages, message)
	c.viewPort.SetContent(lipgloss.NewStyle().Width(c.viewPort.Width()).Render(strings.Join(c.messages, "\n")))
	c.viewPort.GotoBottom()
}

func (c *ChatSessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ReceiveMessageResponse:
		c.agentIsRunning = true
		c.sendAiMessage(msg.Response.Message, msg.Response.Type)
		return c, c.spinner.Tick
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.ta.SetWidth(msg.Width)
		reserved := 2 // spinner
		c.viewPort.SetHeight(msg.Height - (c.ta.Height() + lipgloss.Height(common.OptionsLine) + reserved))
		c.viewPort.SetWidth(msg.Width / 2)
		if len(c.messages) > 0 {
			// Wrap content before setting it.
			c.viewPort.SetContent(lipgloss.NewStyle().Width(c.viewPort.Width()).Render(strings.Join(c.messages, "\n")))
		}
		c.viewPort.GotoBottom()
	case spinner.TickMsg:
		var cmd tea.Cmd
		c.spinner, cmd = c.spinner.Update(msg)
		return c, cmd
	case tea.KeyPressMsg:
		if c.showWarning {
			msgS := msg.String()
			confirmation := msgS == "y"
			if msgS == "y" || msgS == "n" {
				c.showWarning = false
			}
			return c, func() tea.Msg {
				return ConfirmMessage{
					ShouldContinue: confirmation,
				}
			}
		}

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
			if c.ta.Value() == "" || c.agentIsRunning || c.showWarning {
				return c, nil
			}
			sentMessage := c.ta.Value()
			c.sendUserMessage(sentMessage)
			return c, func() tea.Msg {
				return SendMessageRequest{
					UserMessage: sentMessage,
				}
			}
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

func (c *ChatSessionModel) View() tea.View {
	cur := c.ta.Cursor()
	if cur != nil {
		cur.Y += lipgloss.Height(c.viewPort.View())
	}

	// border := lipgloss.NewStyle().Border(lipgloss.BlockBorder()).Render(c.viewPort.View())
	var content string
	if c.agentIsRunning {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			c.viewPort.View(),
			fmt.Sprintf("%s Agent is Running", c.spinner.View()),
			c.ta.View(),
			common.RenderStatusLineBorder(c.width, "opus 4.7"),
		)
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			c.viewPort.View(),
			c.ta.View(),
			common.RenderStatusLineBorder(c.width, "opus 4.7"),
		)
	}

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
