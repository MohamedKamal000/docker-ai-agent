package common

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

type (
	ExecuteStateFunc[T tea.Model] func(s *StateManager[T], m T, msg tea.Msg) (T, tea.Cmd)
	RenderStateFunc[T tea.Model]  func(s *StateManager[T], m T) tea.View
)

// StateDefinition groups the execute and render functions of a single state.
// Registering both together guarantees they can never drift out of sync.
type StateDefinition[T tea.Model] struct {
	Execute ExecuteStateFunc[T]
	Render  RenderStateFunc[T]
}

type StateManager[T tea.Model] struct {
	states        map[uint]StateDefinition[T]
	currentState  uint
	previousState uint // not sure if we would need a queue for state history later or not
}

func NewStateManager[T tea.Model](states map[uint]StateDefinition[T], currentState uint) *StateManager[T] {
	return &StateManager[T]{
		states:       states,
		currentState: currentState,
	}
}

func (s *StateManager[T]) SwitchTo(nextState uint) {
	s.previousState = s.currentState
	s.currentState = nextState
}

func (s *StateManager[T]) definition(state uint) StateDefinition[T] {
	def, ok := s.states[state]
	if !ok {
		panic(fmt.Sprintf("state manager: no definition registered for state %d", state))
	}
	return def
}

func (s *StateManager[T]) ExecuteCurrent(m T, msg tea.Msg) (tea.Model, tea.Cmd) {
	return s.definition(s.currentState).Execute(s, m, msg)
}

func (s *StateManager[T]) RenderCurrent(m T) tea.View {
	return s.definition(s.currentState).Render(s, m)
}

func (s *StateManager[T]) RenderPrevious(m T) tea.View {
	return s.definition(s.previousState).Render(s, m)
}

func (s *StateManager[T]) SwitchToPreviousState() {
	s.SwitchTo(s.previousState)
}
