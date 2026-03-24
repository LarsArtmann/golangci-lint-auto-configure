package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SpinnerModel is a simple tea model that displays a spinner.
type SpinnerModel struct {
	spinner  spinner.Model
	message  string
	status   string
	done     bool
	err      error
	quitting bool
}

// NewSpinnerModel creates a new spinner model.
func NewSpinnerModel(message string) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	return SpinnerModel{
		spinner: s,
		message: message,
		status:  "Working...",
	}
}

// Init initializes the spinner.
func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles update messages.
func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		newSpinner, cmd := m.spinner.Update(msg)
		m.spinner = newSpinner
		return m, cmd

	case SetStatusMsg:
		m.status = msg.Status
		return m, nil

	case SetErrorMsg:
		m.err = msg.Err
		m.done = true
		return m, tea.Quit

	case SetDoneMsg:
		m.done = true
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the spinner.
func (m SpinnerModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  %s %s\n\n  %s\n\n",
			ErrorStyle.Render("✗"),
			ErrorStyle.Render(m.message),
			ErrorStyle.Render(m.err.Error()),
		)
	}

	if m.quitting {
		return fmt.Sprintf("\n  %s %s\n\n  %s\n\n",
			WarningStyle.Render("⚠"),
			WarningStyle.Render(m.message),
			WarningStyle.Render("Cancelled"),
		)
	}

	if m.done {
		return fmt.Sprintf("\n  %s %s\n\n  %s\n\n",
			SuccessStyle.Render("✓"),
			SuccessStyle.Render(m.message),
			InfoStyle.Render(m.status),
		)
	}

	return fmt.Sprintf("\n  %s %s %s\n\n  Press q to quit\n\n",
		m.spinner.View(),
		TitleStyle.Render(m.message),
		BodyStyle.Render(m.status),
	)
}

// Messages for the spinner model.
type (
	SetStatusMsg struct{ Status string }
	SetErrorMsg struct{ Err error }
	SetDoneMsg  struct{}
)

// SetStatus sets the status message.
func SetStatus(s string) tea.Msg {
	return SetStatusMsg{Status: s}
}

// SetError sets an error and ends the spinner.
func SetError(err error) tea.Msg {
	return SetErrorMsg{Err: err}
}

// SetDone ends the spinner successfully.
func SetDone() tea.Msg {
	return SetDoneMsg{}
}

// RunSpinner runs a spinner with the given message and work function.
func RunSpinner(message string, work func() error) error {
	m := NewSpinnerModel(message)

	p := tea.NewProgram(m)

	// Run work in a goroutine
	errChan := make(chan error, 1)
	go func() {
		err := work()
		if err != nil {
			p.Send(SetError(err))
		} else {
			p.Send(SetDone())
		}
	}()

	// Run the program
	_, err := p.Run()
	if err != nil {
		return err
	}

	// Check if we received an error from the work function
	select {
	case e := <-errChan:
		return e
	default:
		return nil
	}
}

// RunSpinnerWithStatus runs a spinner with status updates.
func RunSpinnerWithStatus(message string, work func(func(string)) error) error {
	m := NewSpinnerModel(message)

	p := tea.NewProgram(m)

	// Run work in a goroutine
	errChan := make(chan error, 1)
	go func() {
		statusFunc := func(s string) {
			p.Send(SetStatus(s))
		}
		err := work(statusFunc)
		if err != nil {
			errChan <- err
			p.Send(SetError(err))
		} else {
			p.Send(SetDone())
		}
	}()

	// Run the program
	_, err := p.Run()
	if err != nil {
		return err
	}

	// Check if we received an error from the work function
	select {
	case e := <-errChan:
		return e
	default:
		return nil
	}
}

// ProgressDisplay displays a simple progress indicator.
func ProgressDisplay(current, total int, label string) string {
	width := 40
	filled := (current * width) / total
	empty := width - filled

	progressBar := SuccessStyle.Render(repeatString("█", filled)) +
		MutedColor.Render(repeatString("░", empty))

	return fmt.Sprintf("%s [%s] %d%% %s",
		BodyStyle.Render(label),
		progressBar,
		(current*100)/total,
		BodyStyle.Render(fmt.Sprintf("(%d/%d)", current, total)),
	)
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// MultiSpinner allows running multiple spinners with output capture.
type MultiSpinner struct {
	output  io.Writer
	message string
}

// NewMultiSpinner creates a new multi-spinner.
func NewMultiSpinner(message string, output io.Writer) *MultiSpinner {
	return &MultiSpinner{
		output:  output,
		message: message,
	}
}

// Start starts the spinner.
func (m *MultiSpinner) Start() {
	fmt.Fprintf(m.output, "\n  %s %s\n", spinner.New().View(), TitleStyle.Render(m.message))
}

// Stop stops the spinner with a success message.
func (m *MultiSpinner) Stop() {
	fmt.Fprintf(m.output, "\n  %s %s\n", SuccessStyle.Render("✓"), SuccessStyle.Render(m.message))
}

// StopWithError stops the spinner with an error message.
func (m *MultiSpinner) StopWithError(err error) {
	fmt.Fprintf(m.output, "\n  %s %s\n  %s\n", ErrorStyle.Render("✗"), ErrorStyle.Render(m.message), ErrorStyle.Render(err.Error()))
}
