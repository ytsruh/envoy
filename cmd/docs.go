/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var DocsDir string

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "View your project documentation",
	Long:  `Navigate & view your project documentation in your terminal. Markdown files in the .docs directory are automatically parsed and displayed. Alternatively set your own documentation directory using the --dir flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		dirFlag, err := cmd.Flags().GetString("dir")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		showDocs(dirFlag)
	},
}

func init() {
	// Hide the auto-generated help command & register with the root command
	docsCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.AddCommand(docsCmd)

	// Add the --dir flag to the docs command
	docsCmd.Flags().StringVarP(&DocsDir, "dir", "d", ".docs", "Set the documentation directory")
}

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type item struct {
	name, path, fileName string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return i.path }
func (i item) FilterValue() string { return i.name }

type model struct {
	list     list.Model
	selected struct {
		title   string
		content string
	}
	viewport viewport.Model
	ready    bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			// If content is not empty, clear it and update viewport & go back to list view
			if m.selected.content != "" {
				m.selected.content = ""
				m.selected.title = ""
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)

				return m, tea.Batch(cmds...)
			}
			return m, tea.Quit
		}

		if msg.String() == "enter" {
			if item, ok := m.list.SelectedItem().(item); ok {
				content, err := os.ReadFile(item.path)
				if err != nil {
					fmt.Println("could not load file:", err)
					os.Exit(1)
				}
				m.selected.content = string(content)
				m.selected.title = item.fileName
				if m.ready {
					m.viewport.SetContent(m.selected.content)
				}
			}
		}

		// When viewing content, pass keys to viewport, otherwise to list
		if m.selected.content != "" {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.MouseMsg:
		if m.selected.content != "" {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

		headerHeight := lipgloss.Height(m.headerView(m.selected.title))
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			if m.selected.content != "" {
				m.viewport.SetContent(m.selected.content)
			}
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		if m.selected.content != "" {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	default:
		// Pass all other messages to the appropriate component
		if m.selected.content != "" {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.selected.content != "" {
		return fmt.Sprintf("%s\n%s\n%s", m.headerView(m.selected.title), m.viewport.View(), m.footerView())
	}
	return docStyle.Render(m.list.View())
}

func (m model) headerView(t string) string {
	if t == "" {
		t = "File"
	}
	title := titleStyle.Render(t)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// isMarkdownFile checks if a filename has a markdown extension
func isMarkdownFile(filename string) bool {
	fn := strings.ToLower(filename)
	return strings.HasSuffix(fn, ".md") ||
		strings.HasSuffix(fn, ".markdown") ||
		strings.HasSuffix(fn, ".mdown") ||
		strings.HasSuffix(fn, ".mkd")
}

func showDocs(dirFlag string) {
	files, err := os.ReadDir(dirFlag)
	if err != nil {
		fmt.Printf("Error: Unable to read directory '%s'.\nPlease check if the directory exists and you have proper permissions.\n", dirFlag)
		return
	}

	items := []list.Item{}
	for _, file := range files {
		if file.IsDir() {
			// Skip directories for now
			continue
		}
		// Skip if DS_Store or not markdown file
		if file.Name() == ".DS_Store" || !isMarkdownFile(file.Name()) {
			continue
		}
		title := strings.Split(file.Name(), ".")[0]
		items = append(items, item{
			name:     title,
			path:     dirFlag + "/" + file.Name(),
			fileName: file.Name(),
		})
	}

	// Check if there are no markdown files
	if len(items) == 0 {
		fmt.Println("No markdown files found in the directory.")
		return
	}

	// Show the docs list
	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0), viewport: viewport.New(0, 0)}
	m.list.Title = "Project Documentation"

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithMouseAllMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
