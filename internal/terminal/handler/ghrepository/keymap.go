package ghrepository

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Refresh     teakey.Binding
	NextTab     teakey.Binding
	PreviousTab teakey.Binding
	Quit        teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.PreviousTab, k.NextTab, k.Refresh, k.Quit}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.PreviousTab},
		{k.NextTab},
		{k.Refresh},
		{k.Quit},
	}
}

var keys = keyMap{
	Refresh: teakey.NewBinding(
		teakey.WithKeys("r", "R"),
		teakey.WithHelp("r/R", "Refresh list"),
	),
	PreviousTab: teakey.NewBinding(
		teakey.WithKeys("left"),
		teakey.WithHelp("←", "previous tab"),
	),
	NextTab: teakey.NewBinding(
		teakey.WithKeys("right"),
		teakey.WithHelp("→", "next tab"),
	),
	Quit: teakey.NewBinding(
		teakey.WithKeys("q", "ctrl+c"),
		teakey.WithHelp("q", "quit"),
	),
}

func (m *ModelGithubRepository) ViewHelp() string {
	return m.Help.View(m.Keys)
}
