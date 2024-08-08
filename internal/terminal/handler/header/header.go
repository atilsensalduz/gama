package header

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/termkit/gama/internal/terminal/handler/spirit"
	"github.com/termkit/gama/internal/terminal/handler/types"
	ts "github.com/termkit/gama/internal/terminal/style"
	"strings"
	"sync"
	"time"
)

// Header is a helper for rendering the Header of the terminal.
type Header struct {
	Viewport *viewport.Model

	keys keyMap

	currentTab int

	modelSpirit *spirit.ModelSpirit

	commonHeaders         []commonHeader
	specialHeader         specialHeader
	specialHeaderInterval time.Duration
	currentSpecialStyle   int
}

type commonHeader struct {
	header    string
	rawHeader string

	inactiveStyle lipgloss.Style
	activeStyle   lipgloss.Style
}

type specialHeader struct {
	header    string
	rawHeader string

	styles []lipgloss.Style
}

// Define sync.Once and NewHeader should return same instance
var (
	once sync.Once
	h    *Header
)

// NewHeader returns a new Header.
func NewHeader() *Header {
	once.Do(func() {
		s := spirit.NewSpirit()
		s.SetLockTabs(false)
		h = &Header{
			modelSpirit:           s,
			specialHeaderInterval: time.Millisecond * 100,
			Viewport:              types.NewTerminalViewport(),
			currentTab:            0,
			keys:                  keys,
		}
	})
	return h
}

func (h *Header) SetCurrentTab(tab int) {
	h.currentTab = tab
}

func (h *Header) GetCurrentTab() int {
	return h.currentTab
}

func (h *Header) AddCommonHeader(header string, activeStyle, inactiveStyle lipgloss.Style) {
	h.commonHeaders = append(h.commonHeaders, commonHeader{
		header:        header,
		rawHeader:     header,
		inactiveStyle: inactiveStyle,
		activeStyle:   activeStyle,
	})
}

func (h *Header) SetSpecialHeader(header string, interval time.Duration, styles ...lipgloss.Style) {
	h.specialHeaderInterval = interval
	if len(styles) == 0 {
		styles = append(styles, ts.TitleStyleDisabled)
	}
	h.specialHeader = specialHeader{
		header:    header,
		rawHeader: header,
		styles:    styles,
	}
}

type UpdateMsg struct {
	Msg               string
	UpdatingComponent string
}

func (h *Header) Init() tea.Cmd {
	return h.tick()
}

func (h *Header) tick() tea.Cmd {
	t := time.NewTimer(h.specialHeaderInterval)
	return func() tea.Msg {
		select {
		case <-t.C:
			return UpdateMsg{
				Msg:               "tick",
				UpdatingComponent: "header",
			}
		}
	}
}

func (h *Header) Update(msg tea.Msg) (*Header, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, h.keys.SwitchTabLeft):
			if !h.modelSpirit.GetLockTabs() {
				h.currentTab = max(h.currentTab-1, 0)
			}
		case key.Matches(msg, h.keys.SwitchTabRight):
			if !h.modelSpirit.GetLockTabs() {
				h.currentTab = min(h.currentTab+1, len(h.commonHeaders)-1)
			}
		}
	case UpdateMsg:
		if msg.UpdatingComponent == "header" {
			if h.currentSpecialStyle >= len(h.specialHeader.styles)-1 {
				h.currentSpecialStyle = 0
			} else {
				h.currentSpecialStyle++
			}
		}

		return h, h.Init()
	}

	return h, nil
}

// View renders the Header.
func (h *Header) View() string {
	var titleLen int
	for _, title := range h.commonHeaders {
		titleLen += len(title.rawHeader)
		titleLen += title.activeStyle.GetPaddingLeft() + title.activeStyle.GetPaddingRight()
		titleLen += 2 // for the border between titles
	}

	var renderedTitles []string
	renderedTitles = append(renderedTitles, "")
	for i, title := range h.commonHeaders {
		if h.modelSpirit.GetLockTabs() {
			if i == 0 {
				renderedTitles = append(renderedTitles, title.activeStyle.Render(title.header))
			} else {
				renderedTitles = append(renderedTitles, ts.TitleStyleDisabled.Render(title.header))
			}
		} else {
			if i == h.currentTab {
				renderedTitles = append(renderedTitles, title.activeStyle.Render(title.header))
			} else {
				renderedTitles = append(renderedTitles, title.inactiveStyle.Render(title.header))
			}
		}
	}

	leftCorner := lipgloss.JoinVertical(lipgloss.Top, "╭", "│")
	rightCorner := lipgloss.JoinVertical(lipgloss.Top, "╮", "│")
	leftCorner = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(leftCorner)
	rightCorner = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(rightCorner)

	line := strings.Repeat("─", h.Viewport.Width-(titleLen+2))
	line = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(line)

	return lipgloss.JoinHorizontal(lipgloss.Bottom, leftCorner, lipgloss.JoinHorizontal(lipgloss.Center, append(renderedTitles, line)...), rightCorner)
}
