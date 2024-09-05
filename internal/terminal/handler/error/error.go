package error

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	ts "github.com/termkit/gama/internal/terminal/handler/types"
	"github.com/termkit/skeleton"
	"strings"
)

type ModelError struct {
	skeleton *skeleton.Skeleton
	// err is hold the error
	err error

	// errorMessage is hold the error message
	errorMessage string

	// message is hold the message, if there is no error
	message string

	// messageType is hold the message type
	messageType MessageType
}

type UpdateSelf struct {
	Message    string
	InProgress bool
}

type MessageType string

const (
	// MessageTypeDefault is the message type for default
	MessageTypeDefault MessageType = "default"

	// MessageTypeProgress is the message type for progress
	MessageTypeProgress MessageType = "progress"

	// MessageTypeSuccess is the message type for success
	MessageTypeSuccess MessageType = "success"
)

func SetupModelError(skeleton *skeleton.Skeleton) ModelError {
	return ModelError{
		skeleton:     skeleton,
		err:          nil,
		errorMessage: "",
	}
}

func (m *ModelError) View() string {
	var windowStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())
	width := m.skeleton.GetTerminalWidth() - 4
	doc := strings.Builder{}

	if m.HaveError() {
		windowStyle = ts.WindowStyleError.Width(width)
		doc.WriteString(windowStyle.Render(m.viewError()))
		return lipgloss.JoinHorizontal(lipgloss.Top, doc.String())
	}

	switch m.messageType {
	case MessageTypeDefault:
		windowStyle = ts.WindowStyleDefault.Width(width)
	case MessageTypeProgress:
		windowStyle = ts.WindowStyleProgress.Width(width)
	case MessageTypeSuccess:
		windowStyle = ts.WindowStyleSuccess.Width(width)
	default:
		windowStyle = ts.WindowStyleDefault.Width(width)
	}

	doc.WriteString(windowStyle.Render(m.viewMessage()))
	return doc.String()
}

func (m *ModelError) SetError(err error) {
	m.err = err
}

func (m *ModelError) SetErrorMessage(message string) {
	m.errorMessage = message
}

func (m *ModelError) SetProgressMessage(message string) {
	m.messageType = MessageTypeProgress
	m.message = message
}

func (m *ModelError) SetSuccessMessage(message string) {
	m.messageType = MessageTypeSuccess
	m.message = message
}

func (m *ModelError) SetDefaultMessage(message string) {
	m.messageType = MessageTypeDefault
	m.message = message
}

func (m *ModelError) GetError() error {
	return m.err
}

func (m *ModelError) GetErrorMessage() string {
	return m.errorMessage
}

func (m *ModelError) GetMessage() string {
	return m.message
}

func (m *ModelError) ResetError() {
	m.err = nil
	m.errorMessage = ""
}

func (m *ModelError) ResetMessage() {
	m.message = ""
}

func (m *ModelError) Reset() {
	m.ResetError()
	m.ResetMessage()
}

func (m *ModelError) HaveError() bool {
	return m.err != nil
}

func (m *ModelError) viewError() string {
	doc := strings.Builder{}
	doc.WriteString(fmt.Sprintf("Error [%v]: %s", m.err, m.errorMessage))
	return doc.String()
}

func (m *ModelError) viewMessage() string {
	doc := strings.Builder{}
	doc.WriteString(m.message)
	return doc.String()
}
