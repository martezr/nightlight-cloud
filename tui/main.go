package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/vishvananda/netlink"
)

var (
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
	focused    = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurred    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
)

type mode int

const (
	listMode mode = iota
	formMode
	confirmMode
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list   list.Model
	mode   mode
	inputs []textinput.Model
	focus  int
	err    string
	status string
}

func (m model) Init() tea.Cmd {
	return nil
}

// ================= PARSE NETWORK CONFIG =================
func loadNetworkConfig() (ip, subnet, gateway, dns1, dns2 string) {
	active, err := netlink.LinkByName("nl-external")
	if err != nil {
		log.Println("Error getting link:", err)
		return "", "", "", "", ""
	}

	// IP + Subnet
	addrs, err := netlink.AddrList(active, netlink.FAMILY_V4)
	if err == nil && len(addrs) > 0 {
		addr := addrs[0].IPNet
		ip = addr.IP.String()
		mask := addr.Mask
		subnet = fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
	}

	// Default gateway (system-wide)
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err == nil {
		for _, r := range routes {
			if r.Dst == nil && r.Gw != nil {
				gateway = r.Gw.String()
				break
			}
		}
	}

	gateway = routes[0].Gw.String()

	// DNS servers (first two from /etc/resolv.conf)
	file, err := os.Open("/etc/resolv.conf")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		count := 0
		for scanner.Scan() && count < 2 {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "nameserver") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					if count == 0 {
						dns1 = parts[1]
					} else if count == 1 {
						dns2 = parts[1]
					}
					count++
				}
			}
		}
	}

	return
}

func validateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// ================= UPDATE =================
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {

	case listMode:
		switch msg := msg.(type) {

		case tea.KeyPressMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				if i, ok := m.list.SelectedItem().(item); ok {
					switch i.title {
					case "Configure Hostname":
						m.status = "Hostname configuration coming soon..."
						err := setHostname(m.inputs[0].Value())
						if err != nil {
							log.Fatalf("Error setting hostname: %v", err)
						}
						m.status = fmt.Sprintf("Hostname set to %s", m.inputs[0].Value())
						return m, nil
					case "Reboot Host":
						m.status = "Rebooting host..."
						fmt.Println(m.status) // optional: log to stdout
						// Sync filesystems first
						syscall.Sync()
						// Trigger reboot
						err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
						if err != nil {
							m.err = fmt.Sprintf("Failed to reboot: %v", err)
							return m, nil
						}
						return m, tea.Quit
					case "Shutdown Host":
						m.status = "Shutting down host..."
						fmt.Println(m.status) // optional: log to stdout
						// Sync filesystems first
						syscall.Sync()
						// Trigger shutdown
						err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
						if err != nil {
							m.err = fmt.Sprintf("Failed to shutdown: %v", err)
							return m, nil
						}
						return m, tea.Quit
					case "Configure Networking":
						ip, subnet, gateway, dns1, dns2 := loadNetworkConfig()
						values := []string{ip, subnet, gateway, dns1, dns2}

						for i := range m.inputs {
							m.inputs[i].SetValue(values[i])
							m.inputs[i].Blur()
						}
						m.inputs[0].Focus()
						m.mode = formMode
						m.focus = 0
						m.err = ""
						return m, nil
					}
				}
			}
		case tea.WindowSizeMsg:
			h, v := docStyle.GetFrameSize()
			m.list.SetSize(msg.Width-h, msg.Height-v)
		}

		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case formMode:
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				m.mode = listMode
				return m, nil
			case "tab", "down":
				m.inputs[m.focus].Blur()
				m.focus = (m.focus + 1) % len(m.inputs)
				m.inputs[m.focus].Focus()
				return m, nil
			case "shift+tab", "up":
				m.inputs[m.focus].Blur()
				m.focus--
				if m.focus < 0 {
					m.focus = len(m.inputs) - 1
				}
				m.inputs[m.focus].Focus()
				return m, nil
			case "enter":
				ip := m.inputs[0].Value()
				subnet := m.inputs[1].Value()
				gateway := m.inputs[2].Value()
				dns1 := m.inputs[3].Value()
				dns2 := m.inputs[4].Value()

				if !validateIP(ip) || !validateIP(gateway) || !validateIP(dns1) || !validateIP(dns2) {
					m.err = "Invalid IP/Gateway/DNS"
					return m, nil
				}
				if !validateIP(subnet) {
					m.err = "Invalid Subnet"
					return m, nil
				}

				m.err = ""
				m.mode = confirmMode
				return m, nil
			}
		}

		cmds := make([]tea.Cmd, len(m.inputs))
		for i := range m.inputs {
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}
		return m, tea.Batch(cmds...)

	case confirmMode:
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "y":
				ip := m.inputs[0].Value()
				subnet := m.inputs[1].Value()
				gateway := m.inputs[2].Value()
				dns1 := m.inputs[3].Value()
				dns2 := m.inputs[4].Value()

				// Simulated apply (replace with real file write)
				m.status = fmt.Sprintf(
					"Applied: IP=%s Subnet=%s Gateway=%s DNS1=%s DNS2=%s",
					ip, subnet, gateway, dns1, dns2,
				)

				m.mode = listMode
				return m, nil
			case "n", "esc":
				m.mode = formMode
				return m, nil
			}
		}
	}

	return m, nil
}

// ================= VIEW =================
func (m model) View() tea.View {
	switch m.mode {
	case listMode:
		content := m.list.View()
		if m.status != "" {
			content += "\n\n" + okStyle.Render(m.status)
		}
		return tea.NewView(docStyle.Render(content))

	case formMode:
		s := "Configure Networking\n\n"
		labels := []string{"IP Address", "Subnet Mask", "Gateway", "DNS 1", "DNS 2"}

		for i := range m.inputs {
			style := blurred
			cursor := " "
			if i == m.focus {
				style = focused
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n%s\n\n", cursor, style.Render(labels[i]), m.inputs[i].View())
		}

		if m.err != "" {
			s += errorStyle.Render(m.err) + "\n\n"
		}

		s += "(Tab navigate • Enter continue • Esc cancel)"
		return tea.NewView(docStyle.Render(s))

	case confirmMode:
		s := "Confirm Network Configuration\n\n"
		s += fmt.Sprintf("IP: %s\n", m.inputs[0].Value())
		s += fmt.Sprintf("Subnet: %s\n", m.inputs[1].Value())
		s += fmt.Sprintf("Gateway: %s\n", m.inputs[2].Value())
		s += fmt.Sprintf("DNS 1: %s\n", m.inputs[3].Value())
		s += fmt.Sprintf("DNS 2: %s\n\n", m.inputs[4].Value())
		s += "Apply these settings? (y/n)"
		return tea.NewView(docStyle.Render(s))
	}

	return tea.NewView("")
}

// ================= MAIN =================
func main() {
	items := []list.Item{
		item{"Configure Hostname", "Set system hostname"},
		item{"Configure Networking", "Set IP, subnet, gateway, two DNS servers"},
		item{"Reboot Host", "Restart system"},
		item{"Shutdown Host", "Power off"},
	}

	inputs := make([]textinput.Model, 5)
	placeholders := []string{
		"192.168.1.10",  // IP
		"255.255.255.0", // Subnet
		"192.168.1.1",   // Gateway
		"8.8.8.8",       // DNS1
		"8.8.4.4",       // DNS2
	}

	for i := range inputs {
		ti := textinput.New()
		ti.Placeholder = placeholders[i]
		ti.SetWidth(30)
		inputs[i] = ti
	}

	inputs[0].Focus()

	m := model{
		list:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		mode:   listMode,
		inputs: inputs,
	}

	m.list.Title = "Alpine Network Config"

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// Set the system hostname
func setHostname(hostname string) error {
	err := os.WriteFile("/etc/hostname", []byte(hostname), 0644)
	if err != nil {
		return err
	}

	// Set the hostname immediately
	err = os.WriteFile("/proc/sys/kernel/hostname", []byte(hostname), 0644)
	if err != nil {
		return err
	}

	return nil
}
