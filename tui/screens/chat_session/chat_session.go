package chat_session

import (
	"strings"

	"docker-cli/internal/core"
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
	sidebarState
)

var chatSessionStates = map[uint]common.StateDefinition[*ChatSessionModel]{
	NormalState.Value():       {Execute: NormalStateExecute, Render: NormalStateRender},
	ShowWarningState.Value():  {Execute: ShowWarningStateExecute, Render: ShowWarningStateRender},
	AgentRunningState.Value(): {Execute: AgentRunningStateExecute, Render: AgentRunningStateRender},
	optionsMenuState.Value():  {Execute: OptionsMenuStateExecute, Render: OptionsMenuStateRender},
	sidebarState.Value():      {Execute: SidebarStateExecute, Render: SidebarStateRender},
}

type toolMessage struct {
	toolName string
	command  string
	output   string
}

type ChatSessionModel struct {
	ta                    textarea.Model
	viewPort              viewport.Model
	spinner               spinner.Model
	width                 int
	height                int
	messages              []string
	toolMessages          map[int]toolMessage
	expandedToolMessages  map[int]bool
	stateManager          *common.StateManager[*ChatSessionModel]
	OptionsMenu           widgets.OptionsModel
	pendingWarningMessage string
	modelName             string
	sidebar               *SidebarModel
}

func NewChatSessionModel(modelName string) *ChatSessionModel {
	stateManager := common.NewStateManager(chatSessionStates, NormalState.Value())
	var cs ChatSessionModel
	cs.stateManager = stateManager
	cs.modelName = modelName
	cs.toolMessages = make(map[int]toolMessage)
	cs.expandedToolMessages = make(map[int]bool)
	cs.sidebar = NewSidebarModel()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(common.ColorDockerBlue))
	cs.spinner = sp

	cs.OptionsMenu = widgets.NewOptionsModel()

	ta := textarea.New()
	ta.Placeholder = "Ask about your Docker environment..."
	ta.SetVirtualCursor(false)
	ta.Focus()
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.Prompt = common.StyleInputPrompt.Render("> ")
	ta.CharLimit = 500

	s := ta.Styles()
	s.Focused.CursorLine = lipgloss.NewStyle()
	ta.SetStyles(s)
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(viewport.WithWidth(50), viewport.WithHeight(5))
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)

	cs.ta = ta
	cs.viewPort = vp
	cs.refreshContent()
	return &cs
}

func (c *ChatSessionModel) renderLogoHeader() string {
	logo := common.StylePrimary.Render(common.Logo)
	if lipgloss.Width(logo) < c.viewPort.Width() {
		logo = lipgloss.PlaceHorizontal(c.viewPort.Width(), lipgloss.Center, lipgloss.PlaceVertical(c.viewPort.Height(), lipgloss.Center, logo))
	}
	subtitle := common.StyleTextSecondary.Render("Docker Environment AI Assistant")
	return logo + "\n" + lipgloss.PlaceHorizontal(c.viewPort.Width(), lipgloss.Center, subtitle) + "\n"
}

func (c *ChatSessionModel) refreshContent() {
	if len(c.messages) == 0 {
		c.viewPort.SetContent(c.renderLogoHeader())
		return
	}
	rendered := make([]string, len(c.messages))
	for i, msg := range c.messages {
		if tm, ok := c.toolMessages[i]; ok {
			expanded := c.expandedToolMessages[i]
			rendered[i] = common.RenderToolExecution(tm.toolName, tm.command, tm.output, expanded, c.viewPort.Width())
		} else {
			rendered[i] = msg
		}
	}
	wrapped := lipgloss.NewStyle().Width(c.viewPort.Width()).Render(strings.Join(rendered, "\n\n"))
	c.viewPort.SetContent(wrapped)
}

func (c *ChatSessionModel) Init() tea.Cmd {
	return textarea.Blink
}

func (c *ChatSessionModel) confirmMessage(confirmed bool) {
	if len(c.messages) > 0 {
		c.messages = c.messages[:len(c.messages)-1]
	}
	if confirmed {
		c.messages = append(c.messages, common.RenderConfirmed(c.pendingWarningMessage, c.viewPort.Width()))
	} else {
		c.messages = append(c.messages, common.RenderRejected(c.pendingWarningMessage, c.viewPort.Width()))
	}
	c.refreshContent()
	c.viewPort.GotoBottom()
}

func (c *ChatSessionModel) sendUserMessage(message string) {
	c.messages = append(c.messages, common.RenderUserMessage(message, c.viewPort.Width()))
	c.refreshContent()
	c.ta.Reset()
	c.viewPort.GotoBottom()
}

func (c *ChatSessionModel) appendNewMessage(styledMessage string) {
	c.messages = append(c.messages, styledMessage)
	c.refreshContent()
	c.viewPort.GotoBottom()
}

func (c *ChatSessionModel) trackToolMessage(data *core.ToolExecutionData) {
	idx := len(c.messages)
	c.toolMessages[idx] = toolMessage{
		toolName: data.ToolName,
		command:  data.Command,
		output:   data.Output,
	}
	c.messages = append(c.messages, "")
	c.refreshContent()
	c.viewPort.GotoBottom()
}

func (c *ChatSessionModel) toggleToolMessage(idx int) {
	if _, ok := c.toolMessages[idx]; !ok {
		return
	}
	c.expandedToolMessages[idx] = !c.expandedToolMessages[idx]
	c.refreshContent()
	c.viewPort.GotoBottom()
}

func (c *ChatSessionModel) toggleLastToolMessage() {
	if len(c.toolMessages) == 0 {
		return
	}
	lastIdx := -1
	for idx := range c.toolMessages {
		if idx > lastIdx {
			lastIdx = idx
		}
	}
	if lastIdx == -1 {
		return
	}
	c.toggleToolMessage(lastIdx)
}

func (c *ChatSessionModel) messageIndexAtLine(clickY int) int {
	if len(c.messages) == 0 {
		return -1
	}
	line := 0
	for i, msg := range c.messages {
		var rendered string
		if tm, ok := c.toolMessages[i]; ok {
			expanded := c.expandedToolMessages[i]
			rendered = common.RenderToolExecution(tm.toolName, tm.command, tm.output, expanded, c.viewPort.Width())
		} else {
			rendered = msg
		}
		rendered = lipgloss.NewStyle().Width(c.viewPort.Width()).Render(rendered)
		height := lipgloss.Height(rendered)
		if clickY >= line && clickY < line+height {
			return i
		}
		line += height
	}
	return -1
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
	if cmd := c.sidebar.HandleMessage(msg); cmd != nil {
		c.refreshContent()
		return c, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.ta.SetWidth(msg.Width)
		reserved := 2
		c.viewPort.SetHeight(msg.Height - (c.ta.Height() + lipgloss.Height(common.RenderStatusLine(msg.Width, c.modelName)) + reserved))
		c.viewPort.SetWidth(msg.Width)
		c.refreshContent()
		c.viewPort.GotoBottom()
	}

	return c.stateManager.ExecuteCurrent(c, msg)
}

func (c *ChatSessionModel) View() tea.View {
	content := c.stateManager.RenderCurrent(c).Content
	v := tea.NewView(content)

	cur := c.ta.Cursor()
	if cur != nil {
		cur.Y += c.height - c.ta.Height() - 1
	}

	v.AltScreen = true
	v.Cursor = cur
	v.MouseMode = tea.MouseModeAllMotion
	return v
}
