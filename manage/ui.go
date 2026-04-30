package manage

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Styles ────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ffff")).
			Bold(true).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffff00"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000"))

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ffff")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	menuStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00ffff")).
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Padding(0, 1)
)

// ─── Model Types ──────────────────────────────────────────

type screen int

const (
	screenMain screen = iota
	screenStatus
	screenPrefix
	screenTheme
	screenUpdate
	screenLogs
	screenShell
	screenMessage
)

// MainModel is the top-level Bubble Tea model.
type MainModel struct {
	screen      screen
	composeFile string
	quit        bool
	message     string
	msgIsError  bool

	// Sub-models
	statusModel *statusModel
	prefixModel *prefixModel
	themeModel  *themeModel
	updateModel *updateModel
	logsModel   *logsModel
	shellModel  *shellModel
}

// ─── Status Model (read-only info) ────────────────────────

type statusModel struct {
	loading bool
	ready   bool
}

// ─── Prefix Model ─────────────────────────────────────────

type prefixModel struct {
	input      textinput.Model
	done       bool
	restart    bool
	message    string
	msgIsError bool
}

func newPrefixModel(current string) *prefixModel {
	ti := textinput.New()
	ti.Placeholder = current
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40
	ti.SetValue(current)
	return &prefixModel{input: ti}
}

// ─── Theme Model ──────────────────────────────────────────

type themeModel struct {
	list       list.Model
	done       bool
	restart    bool
	message    string
	msgIsError bool
}

func newThemeModel(current string) *themeModel {
	items := make([]list.Item, len(DaisyUIThemes))
	for i, t := range DaisyUIThemes {
		items[i] = themeItem{name: t}
	}
	l := list.New(items, themeDelegate{}, 60, 18)
	l.Title = T("select_theme")
	l.SetShowHelp(true)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.HelpStyle = helpStyle
	// Select current theme
	for i, t := range DaisyUIThemes {
		if t == current {
			l.Select(i)
			break
		}
	}
	return &themeModel{list: l}
}

type themeItem struct{ name string }

func (t themeItem) FilterValue() string { return t.name }
func (t themeItem) Title() string       { return t.name }
func (t themeItem) Description() string { return "" }

type themeDelegate struct{}

func (d themeDelegate) Height() int                               { return 1 }
func (d themeDelegate) Spacing() int                              { return 0 }
func (d themeDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d themeDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(themeItem)
	if !ok {
		return
	}
	cursor := "  "
	if index == m.Index() {
		cursor = "▶ "
		fmt.Fprintf(w, "%s%s", highlightStyle.Render(cursor), highlightStyle.Render(i.name))
	} else {
		fmt.Fprintf(w, "%s%s", dimStyle.Render("  "), i.name)
	}
}

// ─── Update Model (version selector) ──────────────────────

type updateModel struct {
	list            list.Model
	versions        []string
	loading         bool
	loadErr         string
	done            bool
	message         string
	msgIsError      bool
	manual          bool
	input           textinput.Model
	selectedVersion string
	spinner         int
}

func newUpdateModel() *updateModel {
	l := list.New([]list.Item{}, updateDelegate{}, 60, 14)
	l.Title = T("select_version")
	l.SetShowHelp(true)
	l.SetShowStatusBar(true)
	l.Styles.Title = titleStyle
	l.Styles.HelpStyle = helpStyle
	ti := textinput.New()
	ti.Placeholder = "latest"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40
	return &updateModel{list: l, loading: true, input: ti}
}

type updateItem struct{ name string }

func (u updateItem) FilterValue() string { return u.name }
func (u updateItem) Title() string       { return u.name }
func (u updateItem) Description() string { return "" }

type updateDelegate struct{}

func (d updateDelegate) Height() int                               { return 1 }
func (d updateDelegate) Spacing() int                              { return 0 }
func (d updateDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d updateDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(updateItem)
	if !ok {
		return
	}
	cursor := "  "
	name := i.name
	if index == m.Index() {
		cursor = "▶ "
		if name == "latest" {
			fmt.Fprintf(w, "%s%s %s", highlightStyle.Render(cursor), highlightStyle.Render(name), dimStyle.Render("(most recent)"))
		} else if name == T("manual_input_option") || name == "[Manual input]" || name == "[手动输入]" {
			fmt.Fprintf(w, "%s%s", warnStyle.Render(cursor), warnStyle.Render(name))
		} else {
			fmt.Fprintf(w, "%s%s", highlightStyle.Render(cursor), highlightStyle.Render(name))
		}
	} else {
		if name == "latest" {
			fmt.Fprintf(w, "%s%s %s", dimStyle.Render("  "), infoStyle.Render(name), dimStyle.Render("(most recent)"))
		} else if name == T("manual_input_option") || name == "[Manual input]" || name == "[手动输入]" {
			fmt.Fprintf(w, "%s%s", dimStyle.Render("  "), warnStyle.Render(name))
		} else {
			fmt.Fprintf(w, "%s%s", dimStyle.Render("  "), name)
		}
	}
}

// ─── Logs Model ───────────────────────────────────────────

type logsModel struct {
	done       bool
	choice     int // 0=all, 1=backend, 2=frontend, 3=db
	follow     bool
	message    string
	msgIsError bool
}

// ─── Shell Model ───────────────────────────────────────────

type shellModel struct {
	done       bool
	choice     int // 1=backend, 2=frontend, 3=db
	message    string
	msgIsError bool
}

// ─── Main Model Constructor ───────────────────────────────

func NewMainModel() *MainModel {
	return &MainModel{}
}

// ─── Init ─────────────────────────────────────────────────

func (m *MainModel) Init() tea.Cmd {
	if m.statusModel == nil {
		m.statusModel = &statusModel{loading: true}
		return tea.Batch(tea.EnterAltScreen, loadVersionsCmd)
	}
	return nil
}

// ─── Messages ─────────────────────────────────────────────

type versionsLoadedMsg []string
type versionsLoadErrMsg struct{ err error }
type tickMsg struct{}

type updateDoneMsg struct {
	err     error
	version string
}
type restartDoneMsg struct{ err error }

// ─── Commands ─────────────────────────────────────────────

func loadVersionsCmd() tea.Msg {
	versions, err := FetchVersions()
	if err != nil {
		return versionsLoadErrMsg{err}
	}
	return versionsLoadedMsg(versions)
}

func tickCmd() tea.Cmd {
	return tea.Every(100, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func doUpdateCmd(composeFile, version string) tea.Cmd {
	return func() tea.Msg {
		err := PullAndRestart(composeFile, version)
		return updateDoneMsg{err: err, version: version}
	}
}

func doRestartCmd(composeFile string) tea.Cmd {
	return func() tea.Msg {
		err := RestartServices(composeFile)
		return restartDoneMsg{err: err}
	}
}

// ─── Update ───────────────────────────────────────────────

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c", "q":
			if m.screen != screenMain {
				m.screen = screenMain
				return m, nil
			}
			m.quit = true
			return m, tea.Quit
		}
	}

	switch m.screen {
	case screenMain:
		return m.updateMain(msg)
	case screenStatus:
		return m.updateStatus(msg)
	case screenPrefix:
		return m.updatePrefix(msg)
	case screenTheme:
		return m.updateTheme(msg)
	case screenUpdate:
		return m.updateVersionSelector(msg)
	case screenLogs:
		return m.updateLogs(msg)
	case screenShell:
		return m.updateShell(msg)
	case screenMessage:
		return m.updateMessage(msg)
	}
	return m, nil
}

// ─── Main Menu Update ─────────────────────────────────────

func (m *MainModel) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.prefixModel = newPrefixModel(GetEnv("ADMIN_ROUTE_PREFIX", "control-panel"))
			m.screen = screenPrefix
			return m, textinput.Blink
		case "2":
			m.themeModel = newThemeModel(GetEnv("THEME", "sunset"))
			m.screen = screenTheme
			return m, nil
		case "3":
			m.screen = screenUpdate
			if m.updateModel == nil {
				m.updateModel = newUpdateModel()
				return m, tea.Batch(loadVersionsCmd, textinput.Blink)
			}
			return m, nil
		case "4":
			m.screen = screenStatus
			return m, nil
		case "5":
			return m, doRestartCmd(m.composeFile)
		case "6":
			m.logsModel = &logsModel{}
			m.screen = screenLogs
			return m, nil
		case "7":
			m.shellModel = &shellModel{}
			m.screen = screenShell
			return m, nil
		case "8":
			ToggleLang()
			return m, nil
		case "0":
			m.quit = true
			return m, tea.Quit
		}
	// Handle async update result from main menu restart
	case restartDoneMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("%s: %v", T("restart_fail"), msg.err)
			m.msgIsError = true
		} else {
			m.message = T("restart_ok")
			m.msgIsError = false
		}
		m.screen = screenMessage
		return m, nil
	}
	return m, nil
}

// ─── Status Update ────────────────────────────────────────

func (m *MainModel) updateStatus(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		m.screen = screenMain
		return m, nil
	}
	return m, nil
}

// ─── Prefix Update ────────────────────────────────────────

func (m *MainModel) updatePrefix(msg tea.Msg) (tea.Model, tea.Cmd) {
	pm := m.prefixModel
	if pm.done {
		if _, ok := msg.(tea.KeyMsg); ok {
			m.screen = screenMain
			return m, nil
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.screen = screenMain
			return m, nil
		case "enter":
			newPrefix := pm.input.Value()
			if newPrefix == "" {
				newPrefix = GetEnv("ADMIN_ROUTE_PREFIX", "control-panel")
			}
			if err := UpdatePrefix(newPrefix); err != nil {
				pm.done = true
				pm.message = fmt.Sprintf("Error: %v", err)
				pm.msgIsError = true
				return m, nil
			}
			pm.done = true
			pm.message = fmt.Sprintf(T("prefix_updated"), newPrefix) + "\n\n" + T("restart_question")
			return m, nil
		case "y", "Y":
			if pm.done {
				pm.restart = true
				m.screen = screenMain
				return m, doRestartCmd(m.composeFile)
			}
		case "n", "N":
			if pm.done {
				pm.restart = false
				m.screen = screenMain
				return m, nil
			}
		}
	case restartDoneMsg:
		if !pm.restart {
			return m, nil
		}
		if msg.err != nil {
			m.message = fmt.Sprintf("%s: %v", T("restart_fail"), msg.err)
			m.msgIsError = true
		} else {
			m.message = T("restart_ok")
			m.msgIsError = false
		}
		m.screen = screenMessage
		return m, nil
	}

	var cmd tea.Cmd
	pm.input, cmd = pm.input.Update(msg)
	return m, cmd
}

// ─── Theme Update ─────────────────────────────────────────

func (m *MainModel) updateTheme(msg tea.Msg) (tea.Model, tea.Cmd) {
	tm := m.themeModel
	if tm.done {
		if _, ok := msg.(tea.KeyMsg); ok {
			m.screen = screenMain
			return m, nil
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.screen = screenMain
			return m, nil
		case "enter":
			if item, ok := tm.list.SelectedItem().(themeItem); ok {
				newTheme := item.name
				if err := UpdateTheme(newTheme); err != nil {
					tm.done = true
					tm.message = fmt.Sprintf("Error: %v", err)
					tm.msgIsError = true
					return m, nil
				}
				tm.done = true
				tm.message = fmt.Sprintf(T("theme_updated"), newTheme) + "\n\n" + T("restart_question")
				return m, nil
			}
		case "y", "Y":
			if tm.done {
				tm.restart = true
				m.screen = screenMain
				return m, doRestartCmd(m.composeFile)
			}
		case "n", "N":
			if tm.done {
				tm.restart = false
				m.screen = screenMain
				return m, nil
			}
		}
	case restartDoneMsg:
		if !tm.restart {
			return m, nil
		}
		if msg.err != nil {
			m.message = fmt.Sprintf("%s: %v", T("restart_fail"), msg.err)
			m.msgIsError = true
		} else {
			m.message = T("restart_ok")
			m.msgIsError = false
		}
		m.screen = screenMessage
		return m, nil
	}

	var cmd tea.Cmd
	tm.list, cmd = tm.list.Update(msg)
	return m, cmd
}

// ─── Version Selector Update ──────────────────────────────

func (m *MainModel) updateVersionSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	um := m.updateModel
	if um.done {
		switch msg := msg.(type) {
		case tickMsg:
			um.spinner = (um.spinner + 1) % 10
			return m, tickCmd()
		case updateDoneMsg:
			if msg.err != nil {
				m.message = fmt.Sprintf("%s: %v", T("update_failed"), msg.err)
				m.msgIsError = true
			} else {
				m.message = fmt.Sprintf(T("update_done"), msg.version)
				m.msgIsError = false
			}
			m.screen = screenMessage
			return m, nil
		case tea.KeyMsg:
			m.screen = screenMain
			return m, nil
		}
		return m, tickCmd()
	}

	if um.manual {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				um.manual = false
				return m, nil
			case "enter":
				ver := um.input.Value()
				if ver == "" {
					ver = "latest"
				}
				um.selectedVersion = ver
				um.done = true
				return m, tea.Batch(doUpdateCmd(m.composeFile, ver), tickCmd())
			}
		}
		var cmd tea.Cmd
		um.input, cmd = um.input.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case versionsLoadedMsg:
		um.loading = false
		items := make([]list.Item, len(msg)+1)
		for i, v := range msg {
			items[i] = updateItem{name: v}
		}
		items[len(msg)] = updateItem{name: T("manual_input_option")}
		um.list.SetItems(items)
		return m, nil
	case versionsLoadErrMsg:
		um.loading = false
		um.loadErr = msg.err.Error()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.screen = screenMain
			return m, nil
		case "enter":
			if um.loading || um.loadErr != "" {
				return m, nil
			}
			if item, ok := um.list.SelectedItem().(updateItem); ok {
				if item.name == T("manual_input_option") || item.name == "[Manual input]" || item.name == "[手动输入]" {
					um.manual = true
					um.input.Focus()
					return m, textinput.Blink
				}
				ver := item.name
				um.selectedVersion = ver
				um.done = true
				return m, tea.Batch(doUpdateCmd(m.composeFile, ver), tickCmd())
			}
		}
	}

	var cmd tea.Cmd
	um.list, cmd = um.list.Update(msg)
	return m, cmd
}

// ─── Logs Update ──────────────────────────────────────────

func (m *MainModel) updateLogs(msg tea.Msg) (tea.Model, tea.Cmd) {
	lm := m.logsModel
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if lm.choice != 0 {
				m.screen = screenMain
				return m, nil
			}
		case "1":
			lm.choice = 1
			return m, tea.Quit // Quit TUI to show logs
		case "2":
			lm.choice = 2
			return m, tea.Quit
		case "3":
			lm.choice = 3
			return m, tea.Quit
		case "4":
			lm.choice = 0
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

// ─── Shell Update ─────────────────────────────────────────

func (m *MainModel) updateShell(msg tea.Msg) (tea.Model, tea.Cmd) {
	sm := m.shellModel
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.screen = screenMain
			return m, nil
		case "1":
			sm.choice = 1
			return m, tea.Quit
		case "2":
			sm.choice = 2
			return m, tea.Quit
		case "3":
			sm.choice = 3
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

// ─── Message Update ───────────────────────────────────────

func (m *MainModel) updateMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		m.screen = screenMain
		return m, nil
	}
	return m, nil
}

// ─── View ─────────────────────────────────────────────────

func (m *MainModel) View() string {
	switch m.screen {
	case screenMain:
		return m.viewMain()
	case screenStatus:
		return m.viewStatus()
	case screenPrefix:
		return m.viewPrefix()
	case screenTheme:
		return m.viewTheme()
	case screenUpdate:
		return m.viewVersionSelector()
	case screenLogs:
		return m.viewLogs()
	case screenShell:
		return m.viewShell()
	case screenMessage:
		return m.viewMessage()
	}
	return ""
}

// ─── Header ───────────────────────────────────────────────

func (m *MainModel) header() string {
	logo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ffff")).
		Bold(true).
		Render(`
  ██████╗ ██╗   ██╗██████╗ ███████╗ ██████╗ ██████╗ ██████╗ ███████╗
  ██╔══██╗██║   ██║██╔══██╗██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝
  ██████╔╝██║   ██║██████╔╝█████╗  ██║     ██║   ██║██████╔╝█████╗
  ██╔═══╝ ██║   ██║██╔══██╗██╔══╝  ██║     ██║   ██║██╔══██╗██╔══╝
  ██║     ╚██████╔╝██║  ██║███████╗╚██████╗╚██████╔╝██║  ██║███████╗
  ╚═╝      ╚═════╝ ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝`)
	return lipgloss.JoinVertical(lipgloss.Center,
		logo,
		titleStyle.Render(T("title"))+"  "+subtitleStyle.Render(T("title_sub")),
		dimStyle.Render("────────────────────────────────────────────────"),
	)
}

func (m *MainModel) footer() string {
	return helpStyle.Render("1-8: select  •  q/ctrl+c: quit  •  esc: back")
}

// ─── Main Menu View ───────────────────────────────────────

func (m *MainModel) viewMain() string {
	s := m.header() + "\n\n"
	s += menuStyle.Render("┌─ "+T("menu_title")+" ──────────────────────────┐") + "\n"
	s += menuStyle.Render("│") + "\n"
	s += fmt.Sprintf("│  %s  %s\n", highlightStyle.Render("1)"), T("menu_switch_prefix"))
	s += fmt.Sprintf("│  %s  %s\n", highlightStyle.Render("2)"), T("menu_switch_theme"))
	s += fmt.Sprintf("│  %s  %s\n", highlightStyle.Render("3)"), T("menu_update"))
	s += menuStyle.Render("│") + "\n"
	s += fmt.Sprintf("│  %s  %s\n", highlightStyle.Render("4)"), T("menu_status"))
	s += fmt.Sprintf("│  %s  %s\n", highlightStyle.Render("5)"), T("menu_restart"))
	s += fmt.Sprintf("│  %s  %s\n", highlightStyle.Render("6)"), T("menu_logs"))
	s += fmt.Sprintf("│  %s  %s\n", highlightStyle.Render("7)"), T("menu_shell"))
	s += menuStyle.Render("│") + "\n"
	s += fmt.Sprintf("│  %s  %s\n", highlightStyle.Render("8)"), T("menu_lang"))
	s += fmt.Sprintf("│  %s  %s\n", highlightStyle.Render("0)"), T("menu_exit"))
	s += menuStyle.Render("│") + "\n"
	s += menuStyle.Render("└────────────────────────────────────────────┘") + "\n"
	s += "\n" + m.footer()
	return s
}

// ─── Status View ──────────────────────────────────────────

func (m *MainModel) viewStatus() string {
	s := m.header() + "\n"
	s += titleStyle.Render("  "+T("status_title")) + "\n\n"

	ver := GetEnv("PURECORE_VERSION", "latest")
	prefix := GetEnv("ADMIN_ROUTE_PREFIX", "control-panel")
	theme := GetEnv("THEME", "sunset")
	fePort := GetEnv("FRONTEND_PORT", "9001")

	s += fmt.Sprintf("  %s       %s\n", dimStyle.Render(T("status_version")+":"), highlightStyle.Render(ver))
	s += fmt.Sprintf("  %s  %s\n", dimStyle.Render(T("status_prefix")+":"), highlightStyle.Render(prefix))
	s += fmt.Sprintf("  %s         %s\n", dimStyle.Render(T("status_theme")+":"), highlightStyle.Render(theme))
	s += "\n"
	s += fmt.Sprintf("  %s  %s\n", dimStyle.Render(T("status_frontend")+":"), infoStyle.Render("http://localhost:"+fePort))
	s += fmt.Sprintf("  %s     %s\n", dimStyle.Render(T("status_admin")+":"), infoStyle.Render("http://localhost:"+fePort+"/"+prefix))

	running := GetRunningContainers(m.composeFile)
	s += "\n"
	if running > 0 {
		s += fmt.Sprintf("  "+T("containers_found")+"\n", running)
	} else {
		s += "  " + warnStyle.Render(T("no_containers")) + "\n"
	}

	s += "\n" + helpStyle.Render("Press any key to go back")
	return s
}

// ─── Prefix View ──────────────────────────────────────────

func (m *MainModel) viewPrefix() string {
	pm := m.prefixModel
	if pm.done {
		if pm.msgIsError {
			return m.header() + "\n\n" + errorStyle.Render(pm.message) + "\n\n" + helpStyle.Render("Press any key to go back")
		}
		return m.header() + "\n\n" + infoStyle.Render(pm.message) + "\n\n" + helpStyle.Render("y/N to confirm or any other key to skip")
	}

	s := m.header() + "\n"
	s += titleStyle.Render("  "+T("menu_switch_prefix")) + "\n\n"
	s += dimStyle.Render(fmt.Sprintf("  "+T("current_prefix"), GetEnv("ADMIN_ROUTE_PREFIX", "control-panel"))) + "\n\n"
	s += dimStyle.Render("  "+T("enter_new_prefix")) + "\n"
	s += "  " + pm.input.View() + "\n\n"
	s += helpStyle.Render("Enter: confirm  •  Esc: cancel")
	return s
}

// ─── Theme View ───────────────────────────────────────────

func (m *MainModel) viewTheme() string {
	tm := m.themeModel
	if tm.done {
		if tm.msgIsError {
			return m.header() + "\n\n" + errorStyle.Render(tm.message) + "\n\n" + helpStyle.Render("Press any key to go back")
		}
		return m.header() + "\n\n" + infoStyle.Render(tm.message) + "\n\n" + helpStyle.Render("y/N to confirm or any other key to skip")
	}

	s := m.header() + "\n"
	s += titleStyle.Render("  "+T("select_theme")) + "\n"
	s += dimStyle.Render(fmt.Sprintf("  "+T("current_theme")+"\n", GetEnv("THEME", "sunset")))
	s += "\n"
	s += tm.list.View()
	return s
}

// ─── Version Selector View ────────────────────────────────

func (m *MainModel) viewVersionSelector() string {
	um := m.updateModel
	if um.done {
		spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		msg := fmt.Sprintf(T("updating"), um.selectedVersion)
		return m.header() + "\n\n" + infoStyle.Render(spinnerChars[um.spinner]+" "+msg) + "\n\n" + helpStyle.Render("Please wait...")
	}

	if um.manual {
		s := m.header() + "\n"
		s += titleStyle.Render("  "+T("enter_version_manual")) + "\n\n"
		s += "  " + um.input.View() + "\n\n"
		s += helpStyle.Render("Enter: confirm  •  Esc: back")
		return s
	}

	if um.loading {
		return m.header() + "\n\n" + infoStyle.Render(T("fetching_versions")) + "\n\n" + helpStyle.Render("Please wait...")
	}

	if um.loadErr != "" {
		s := m.header() + "\n"
		s += errorStyle.Render(T("loading_failed")+": "+um.loadErr) + "\n\n"
		// Show manual input fallback
		s += dimStyle.Render("  "+T("enter_version_manual")) + "\n"
		s += "  " + um.input.View() + "\n\n"
		s += helpStyle.Render("Enter: confirm  •  Esc: cancel")
		return s
	}

	s := m.header() + "\n"
	s += titleStyle.Render("  "+T("update_title")) + "\n"
	s += dimStyle.Render(fmt.Sprintf("  "+T("current_version")+"\n", GetEnv("PURECORE_VERSION", "latest")))
	s += "\n"
	s += um.list.View()
	return s
}

// ─── Logs View ────────────────────────────────────────────

func (m *MainModel) viewLogs() string {
	s := m.header() + "\n"
	s += titleStyle.Render("  "+T("logs_title")) + "\n\n"
	s += dimStyle.Render("  "+T("select_service")) + "\n\n"
	s += fmt.Sprintf("     %s  %s\n", highlightStyle.Render("1)"), T("logs_all"))
	s += fmt.Sprintf("     %s  %s\n", highlightStyle.Render("2)"), T("logs_backend"))
	s += fmt.Sprintf("     %s  %s\n", highlightStyle.Render("3)"), T("logs_frontend"))
	s += fmt.Sprintf("     %s  %s\n", highlightStyle.Render("4)"), T("logs_database"))
	return s
}

// ─── Shell View ───────────────────────────────────────────

func (m *MainModel) viewShell() string {
	s := m.header() + "\n"
	s += titleStyle.Render("  "+T("shell_title")) + "\n\n"
	s += dimStyle.Render("  "+T("select_container")) + "\n\n"
	s += fmt.Sprintf("     %s  %s\n", highlightStyle.Render("1)"), T("shell_backend"))
	s += fmt.Sprintf("     %s  %s\n", highlightStyle.Render("2)"), T("shell_frontend"))
	s += fmt.Sprintf("     %s  %s\n", highlightStyle.Render("3)"), T("shell_database"))
	return s
}

// ─── Message View ─────────────────────────────────────────

func (m *MainModel) viewMessage() string {
	if m.msgIsError {
		return m.header() + "\n\n" + errorStyle.Render(m.message) + "\n\n" + helpStyle.Render("Press any key to go back")
	}
	return m.header() + "\n\n" + infoStyle.Render(m.message) + "\n\n" + helpStyle.Render("Press any key to go back")
}

// ─── Public API ───────────────────────────────────────────

// Run starts the interactive TUI.
func Run() error {
	// Check installation
	valid, composeFile, _ := CheckInstallation()
	if !valid {
		fmt.Println(errorStyle.Render(T("not_installed_exit")))
		return fmt.Errorf("not a PureCore project directory")
	}

	_, err := tea.NewProgram(&MainModel{composeFile: composeFile}, tea.WithAltScreen()).Run()
	return err
}

// RunLogs displays logs for the chosen service and exits.
func RunLogs(composeFile, service string) error {
	// Validate docker compose
	cmd := FindComposeCmd()
	if cmd == "" {
		fmt.Println(errorStyle.Render(T("compose_not_found")))
		return fmt.Errorf(T("compose_not_found"))
	}
	parts := strings.Fields(cmd)
	args := append(parts, "-f", composeFile, "logs", "--tail=100")
	if service != "" && service != "all" {
		// Map friendly names to compose service names
		switch service {
		case "backend":
			service = "backend"
		case "frontend":
			service = "frontend"
		case "database", "db":
			service = "postgres"
		}
		args = append(args, service)
	}
	c := exec.Command(args[0], args[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// RunShell opens a shell in the specified container.
func RunShell(container string) error {
	containers := map[string]string{
		"backend":  "purecore-backend",
		"frontend": "purecore-frontend",
		"db":       "purecore-db",
	}
	cn, ok := containers[container]
	if !ok {
		cn = container
	}

	// Try /bin/sh first, then /bin/bash
	c := exec.Command("docker", "exec", "-it", cn, "/bin/sh")
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		c = exec.Command("docker", "exec", "-it", cn, "/bin/bash")
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}
	return nil
}
