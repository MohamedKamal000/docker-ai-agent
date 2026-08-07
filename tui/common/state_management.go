package common

import tea "charm.land/bubbletea/v2"

type (
	ExecuteStateFunc[T tea.Model] func(s *StateManager[T], m T, msg tea.Msg) (T, tea.Cmd)
	RenderStateFunc[T tea.Model]  func(s *StateManager[T], m T) tea.View
)

type StateManager[T tea.Model] struct {
	executeStates []ExecuteStateFunc[T]
	renderStates  []RenderStateFunc[T]
	currentState  uint
	previousState uint // not sure if we would need a queue for state history later or not
}

func NewStateManager[T tea.Model](states []ExecuteStateFunc[T], renderStates []RenderStateFunc[T], currentState uint) *StateManager[T] {
	return &StateManager[T]{
		executeStates: states,
		renderStates:  renderStates,
		currentState:  currentState,
	}
}

func (s *StateManager[T]) SwitchTo(nextState uint) {
	s.previousState = s.currentState
	s.currentState = nextState
}

func (s *StateManager[T]) ExecuteCurrent(m T, msg tea.Msg) (tea.Model, tea.Cmd) {
	return s.executeStates[s.currentState](s, m, msg)
}

func (s *StateManager[T]) RenderCurrent(m T) tea.View {
	return s.renderStates[s.currentState](s, m)
}

func (s *StateManager[T]) RenderPrevious(m T) tea.View {
	return s.renderStates[s.previousState](s, m)
}

func (s *StateManager[T]) SwitchToPreviousState() {
	s.SwitchTo(s.previousState)
}
