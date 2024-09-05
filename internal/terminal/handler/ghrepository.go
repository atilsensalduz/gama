package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/termkit/gama/internal/github/domain"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	"github.com/termkit/gama/internal/terminal/handler/taboptions"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	"github.com/termkit/gama/pkg/browser"
	"github.com/termkit/skeleton"
	"strconv"
	"strings"
)

type ModelGithubRepository struct {
	skeleton *skeleton.Skeleton
	// current handler's properties
	syncRepositoriesContext context.Context
	cancelSyncRepositories  context.CancelFunc
	tableReady              bool

	// shared properties
	SelectedRepository *hdltypes.SelectedRepository

	// use cases
	github gu.UseCase

	// keymap
	Keys githubRepositoryKeyMap

	// models
	Help                        help.Model
	tableGithubRepository       table.Model
	searchTableGithubRepository table.Model
	modelError                  *hdlerror.ModelError

	modelTabOptions *taboptions.Options

	textInput textinput.Model

	updateChan chan updateSelf
}

func SetupModelGithubRepository(skeleton *skeleton.Skeleton, githubUseCase gu.UseCase) *ModelGithubRepository {
	var tableRowsGithubRepository []table.Row

	tableGithubRepository := table.New(
		table.WithColumns(tableColumnsGithubRepository),
		table.WithRows(tableRowsGithubRepository),
		table.WithFocused(true),
		table.WithHeight(13),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	tableGithubRepository.SetStyles(s)

	tableGithubRepository.KeyMap = table.KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", " "),
			key.WithHelp("pgdn", "page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "go to end"),
		),
	}

	ti := textinput.New()
	ti.Blur()
	ti.CharLimit = 72
	ti.Placeholder = "Type to search repository"
	ti.ShowSuggestions = false // disable suggestions, it will be enabled future.

	// setup models
	modelError := hdlerror.SetupModelError(skeleton)
	tabOptions := taboptions.NewOptions(&modelError)

	return &ModelGithubRepository{
		skeleton:                skeleton,
		Help:                    help.New(),
		Keys:                    githubRepositoryKeys,
		github:                  githubUseCase,
		tableGithubRepository:   tableGithubRepository,
		modelError:              &modelError,
		SelectedRepository:      hdltypes.NewSelectedRepository(),
		modelTabOptions:         tabOptions,
		textInput:               ti,
		syncRepositoriesContext: context.Background(),
		cancelSyncRepositories:  func() {},
		updateChan:              make(chan updateSelf),
	}
}

func (m *ModelGithubRepository) Init() tea.Cmd {
	openInBrowser := func() {
		m.modelError.SetProgressMessage("Opening in browser...")

		err := browser.OpenInBrowser(fmt.Sprintf("https://github.com/%s", m.SelectedRepository.RepositoryName))
		if err != nil {
			m.modelError.SetError(err)
			m.modelError.SetErrorMessage(fmt.Sprintf("Cannot open in browser: %v", err))
			return
		}

		m.modelError.SetSuccessMessage("Opened in browser")
	}

	m.modelTabOptions.AddOption("Open in browser", openInBrowser)
	go m.syncRepositories(m.syncRepositoriesContext)
	return tea.Batch(m.modelTabOptions.Init(), m.SelfUpdater())
}

func (m *ModelGithubRepository) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	var textInputMsg = msg
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Refresh):
			m.tableReady = false       // reset table ready status
			m.cancelSyncRepositories() // cancel previous sync
			m.syncRepositoriesContext, m.cancelSyncRepositories = context.WithCancel(context.Background())
			go m.syncRepositories(m.syncRepositoriesContext)
		case msg.String() == " " || m.isNumber(msg.String()):
			textInputMsg = tea.KeyMsg{}
		case m.isCharAndSymbol(msg.Runes):
			m.tableGithubRepository.GotoTop()
			m.tableGithubRepository.SetCursor(0)
			m.searchTableGithubRepository.GotoTop()
			m.searchTableGithubRepository.SetCursor(0)
		}
	case updateSelf:
		cmds = append(cmds, m.SelfUpdater())
	}

	m.textInput, cmd = m.textInput.Update(textInputMsg)
	cmds = append(cmds, cmd)

	m.updateTableRowsBySearchBar()

	m.tableGithubRepository, cmd = m.tableGithubRepository.Update(msg)
	cmds = append(cmds, cmd)

	m.searchTableGithubRepository, cmd = m.searchTableGithubRepository.Update(msg)
	cmds = append(cmds, cmd)

	m.modelTabOptions, cmd = m.modelTabOptions.Update(msg)
	cmds = append(cmds, cmd)

	m.handleTableInputs(m.syncRepositoriesContext)

	return m, tea.Batch(cmds...)
}

func (m *ModelGithubRepository) View() string {
	var baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).MarginLeft(1)

	helpWindowStyle := hdltypes.WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)

	var tableWidth int
	for _, t := range tableColumnsGithubRepository {
		tableWidth += t.Width
	}

	newTableColumns := tableColumnsGithubRepository
	widthDiff := m.skeleton.GetTerminalWidth() - tableWidth
	if widthDiff > 0 {
		newTableColumns[0].Width += widthDiff - 14
		m.tableGithubRepository.SetColumns(newTableColumns)
		m.tableGithubRepository.SetHeight(m.skeleton.GetTerminalHeight() - 20)
	}

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableGithubRepository.View()))

	return lipgloss.JoinVertical(lipgloss.Top, doc.String(), m.viewSearchBar(), m.modelTabOptions.View(), m.ViewStatus(), helpWindowStyle.Render(m.ViewHelp()))
}

func (m *ModelGithubRepository) SelfUpdater() tea.Cmd {
	return func() tea.Msg {
		return <-m.updateChan
	}
}

func (m *ModelGithubRepository) syncRepositories(ctx context.Context) {
	m.modelError.Reset() // reset previous errors
	m.modelTabOptions.SetStatus(taboptions.OptionWait)
	m.modelError.SetProgressMessage("Fetching repositories...")

	// delete all rows
	m.tableGithubRepository.SetRows([]table.Row{})
	m.searchTableGithubRepository.SetRows([]table.Row{})

	repositories, err := m.github.ListRepositories(ctx, gu.ListRepositoriesInput{
		Limit: 100, // limit to 100 repositories
		Page:  5,   // page 1 to page 5, at summary we fetch 500 repositories
		Sort:  domain.SortByUpdated,
	})
	if errors.Is(err, context.Canceled) {
		return
	} else if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Repositories cannot be listed")
		return
	}

	if len(repositories.Repositories) == 0 {
		m.modelTabOptions.SetStatus(taboptions.OptionNone)
		m.modelError.SetDefaultMessage("No repositories found")
		m.textInput.Blur()
		return
	}

	m.skeleton.UpdateWidgetValue("repositories", fmt.Sprintf("Repository Count: %d", len(repositories.Repositories)))

	tableRowsGithubRepository := make([]table.Row, 0, len(repositories.Repositories))
	for _, repository := range repositories.Repositories {
		tableRowsGithubRepository = append(tableRowsGithubRepository,
			table.Row{repository.Name, repository.DefaultBranch, strconv.Itoa(repository.Stars), strconv.Itoa(len(repository.Workflows))})
	}

	m.tableGithubRepository.SetRows(tableRowsGithubRepository)
	m.searchTableGithubRepository.SetRows(tableRowsGithubRepository)

	// set cursor to 0
	m.tableGithubRepository.SetCursor(0)
	m.searchTableGithubRepository.SetCursor(0)

	m.tableReady = true
	//m.updateSearchBarSuggestions()
	m.textInput.Focus()
	m.modelError.SetSuccessMessage("Repositories fetched")

	m.updateChan <- updateSelf{}
}

func (m *ModelGithubRepository) handleTableInputs(_ context.Context) {
	if !m.tableReady {
		return
	}

	// To avoid go routine leak
	selectedRow := m.tableGithubRepository.SelectedRow()

	// Synchronize selected repository name with parent model
	if len(selectedRow) > 0 && selectedRow[0] != "" {
		m.SelectedRepository.RepositoryName = selectedRow[0]
		m.SelectedRepository.BranchName = selectedRow[1]
	}

	m.modelTabOptions.SetStatus(taboptions.OptionIdle)
}

func (m *ModelGithubRepository) viewSearchBar() string {
	// Define window style
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		Padding(0, 1).
		Width(m.skeleton.GetTerminalWidth() - 6).MarginLeft(1)

	// Build the options list
	doc := strings.Builder{}

	if len(m.textInput.Value()) > 0 {
		windowStyle = windowStyle.BorderForeground(lipgloss.Color("39"))
	}

	doc.WriteString(m.textInput.View())

	return windowStyle.Render(doc.String())
}

func (m *ModelGithubRepository) updateSearchBarSuggestions() {
	m.textInput.SetSuggestions([]string{})

	var suggestions = make([]string, 0, len(m.tableGithubRepository.Rows()))
	for _, r := range m.tableGithubRepository.Rows() {
		suggestions = append(suggestions, r[0])
	}

	m.textInput.SetSuggestions(suggestions)
}

func (m *ModelGithubRepository) updateTableRowsBySearchBar() {
	var tableRowsGithubRepository = make([]table.Row, 0, len(m.tableGithubRepository.Rows()))

	for _, r := range m.searchTableGithubRepository.Rows() {
		if strings.Contains(r[0], m.textInput.Value()) {
			tableRowsGithubRepository = append(tableRowsGithubRepository, r)
		}
	}

	if len(tableRowsGithubRepository) == 0 {
		m.SelectedRepository.RepositoryName = ""
		m.SelectedRepository.BranchName = ""
		m.SelectedRepository.WorkflowName = ""
	}

	m.tableGithubRepository.SetRows(tableRowsGithubRepository)
}

func (m *ModelGithubRepository) isNumber(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}

	return false
}

func (m *ModelGithubRepository) isCharAndSymbol(r []rune) bool {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_./"
	for _, c := range r {
		if strings.ContainsRune(chars, c) {
			return true
		}
	}

	return false
}

func (m *ModelGithubRepository) ViewStatus() string {
	return m.modelError.View()
}
