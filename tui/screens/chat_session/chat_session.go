package chat_session

import (
	"strings"

	"docker-cli/tui/common"
	"docker-cli/tui/widgets"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type state uint

func (s state) Value() uint {
	return uint(s)
}

const (
	NormalState state = iota
	ShowWarningState
	AgentRunningState
	optionsMenuState
)

var (
	executeStateFunctions = []common.ExecuteStateFunc[*ChatSessionModel]{
		NormalStateExecute,
		ShowWarningStateExecute,
		AgentRunningStateExecute,
		OptionsMenuStateExecute,
	}
	renderStatesFunctions = []common.RenderStateFunc[*ChatSessionModel]{
		NormalStateRender,
		ShowWarningStateRender,
		AgentRunningStateRender,
		OptionsMenuStateRender,
	}
)

type ChatSessionModel struct {
	ta                    textarea.Model
	viewPort              viewport.Model
	spinner               spinner.Model
	width                 int
	height                int
	messages              []string
	stateManager          *common.StateManager[*ChatSessionModel]
	OptionsMenu           widgets.OptionsModel
	pendingWarningMessage string
}

func NewChatSessionModel() *ChatSessionModel {
	stateManager := common.NewStateManager(executeStateFunctions, renderStatesFunctions, NormalState.Value())
	var cs ChatSessionModel
	cs.stateManager = stateManager
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	cs.spinner = sp
	cs.OptionsMenu = widgets.NewOptionsModel()
	ta := textarea.New()
	ta.Placeholder = "send a message...."
	ta.SetVirtualCursor(false)
	ta.Focus()

	ta.SetWidth(30)
	ta.SetHeight(4)
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

func (c *ChatSessionModel) confirmMessage(confirmed bool) {
	if len(c.messages) > 0 {
		c.messages = c.messages[:len(c.messages)-1]
	}
	if confirmed {
		c.messages = append(c.messages, common.RenderConfirmedMessage(c.pendingWarningMessage, c.viewPort.Width()))
	} else {
		c.messages = append(c.messages, common.RenderWarningBody(c.pendingWarningMessage, c.viewPort.Width()))
	}
	c.viewPort.SetContent(lipgloss.NewStyle().Width(c.viewPort.Width()).Render(strings.Join(c.messages, "\n")))
	c.viewPort.GotoBottom()
}

func (c *ChatSessionModel) sendUserMessage(message string) {
	c.messages = append(c.messages, common.RenderUserMessageWithBackground(message, c.viewPort.Width()))
	c.viewPort.SetContent(lipgloss.NewStyle().Width(c.viewPort.Width()).Render(strings.Join(c.messages, "\n")))
	c.ta.Reset()
	c.viewPort.GotoBottom()
}

func (c *ChatSessionModel) appendNewMessage(styledMessage string) {
	c.messages = append(c.messages, styledMessage)
	c.viewPort.SetContent(lipgloss.NewStyle().Width(c.viewPort.Width()).Render(strings.Join(c.messages, "\n")))
	c.viewPort.GotoBottom()
}

func (c *ChatSessionModel) updateChildren(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	var cmd tea.Cmd

	c.ta, cmd = c.ta.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "down", "left", "right", "pgup", "pgdown":
			c.viewPort, cmd = c.viewPort.Update(msg)
			cmds = append(cmds, cmd)
		}
	default:
		c.viewPort, cmd = c.viewPort.Update(msg)
		cmds = append(cmds, cmd)
	}

	c.spinner, cmd = c.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (c *ChatSessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.ta.SetWidth(msg.Width)
		reserved := 2 // spinner
		c.viewPort.SetHeight(msg.Height - (c.ta.Height() + lipgloss.Height(common.RenderStatusLineBorder(msg.Width, "")) + reserved))
		c.viewPort.SetWidth(msg.Width)
		if len(c.messages) > 0 {
			// Wrap content before setting it.
			c.viewPort.SetContent(lipgloss.NewStyle().Width(c.viewPort.Width()).Render(strings.Join(c.messages, "\n")))
		}
		c.viewPort.GotoBottom()
	}

	return c.stateManager.ExecuteCurrent(c, msg)
}

func (c *ChatSessionModel) View() tea.View {
	line := common.RenderStatusLineBorder(c.width, "model name")
	content := c.stateManager.RenderCurrent(c).Content
	v := tea.NewView(content)

	cur := c.ta.Cursor()
	if cur != nil {
		cur.Y += c.height - c.ta.Height() - lipgloss.Height(line)
	}

	v.AltScreen = true
	v.Cursor = cur
	v.MouseMode = tea.MouseModeNone
	return v
}
