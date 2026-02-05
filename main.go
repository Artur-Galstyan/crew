package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type crewConfig struct {
	ServerUrl    string `json:"server_url"`
	ServerApiKey string `json:"server_api_key"`
}
type State int

const (
	ConfiguringCrewUrl State = iota
	ConfiguringCrewApiKey
	Lobby
)

const CREW_URL_PROMPT = "Enter your Crew URL"

type model struct {
	title          string
	serverUrl      string
	serverApiKey   string
	textInput      textinput.Model
	completedLines []string
	err            error
	state          State
}

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	err := os.MkdirAll(getCrewDir(), 0755)

	if err != nil {
		slog.Error("Could not create crew directory:", "err", err)
	}

	crewDir := getCrewDir()
	crewCfgJson := filepath.Join(crewDir, "config.json")

	var cfg crewConfig

	state := ConfiguringCrewUrl

	file, err := os.Open(crewCfgJson)
	if err != nil {
		slog.Debug("No existing config found. ", "err", err)

	} else {
		defer file.Close()
		json.NewDecoder(file).Decode(&cfg)
	}

	return model{
		textInput:      ti,
		title:          "Crew",
		serverUrl:      cfg.ServerUrl,
		serverApiKey:   cfg.ServerApiKey,
		state:          state,
		completedLines: []string{},
		err:            nil,
	}
}

func getCrewDir() string {

	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Could not find home directory:", "err", err)
	}
	crewPath := filepath.Join(home, ".crew")
	return crewPath
}

func (m model) Init() tea.Cmd {

	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case ConfiguringCrewUrl:
		switch msg := msg.(type) {

		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.serverUrl = m.textInput.Value()
				m.completedLines = append(m.completedLines, CREW_URL_PROMPT, m.serverUrl)
				m.textInput.SetValue("")
				m.state = ConfiguringCrewApiKey
				return m, nil

			case "ctrl+c":
				return m, tea.Quit
			}
		}

	case ConfiguringCrewApiKey:
		switch msg := msg.(type) {

		case tea.KeyMsg:
			switch msg.String() {

			case "enter":
				// try to register
				m.serverApiKey = m.textInput.Value()
				m.completedLines = append(m.completedLines, m.serverApiKey)
				m.state = Lobby
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
		}

	default:
		switch msg := msg.(type) {

		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			}
		}

	}

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m model) View() string {

	switch m.state {
	case ConfiguringCrewUrl:
		return fmt.Sprintf("%s:\n%s", CREW_URL_PROMPT, m.textInput.View())

	case ConfiguringCrewApiKey:
		prevLines := strings.Join(m.completedLines, " ")
		return fmt.Sprintf("%s\nEnter your crew API key:\n%s", prevLines, m.textInput.View())

	case Lobby:
		return fmt.Sprintf("Welcome to your Crew @%s", m.serverUrl)

	}

	return fmt.Sprintf(
		"Enter text MAN:\n%s\n\n",
		m.textInput.View(),
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
