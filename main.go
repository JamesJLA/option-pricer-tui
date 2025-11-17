package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Black-Scholes helper
func NormCDF(x float64) float64 {
	return 0.5 * (1.0 + math.Erf(x/math.Sqrt2))
}

func Calc(S, K, T, r, q, v float64, option_type string) float64 {
	d1 := (math.Log(S/K) + (r-q+0.5*v*v)*T) / (v * math.Sqrt(T))
	d2 := d1 - v*math.Sqrt(T)
	if option_type == "call" {
		return S*math.Exp(-q*T)*NormCDF(d1) - K*math.Exp(-r*T)*NormCDF(d2)
	}
	return K*math.Exp(-r*T)*NormCDF(-d2) - S*math.Exp(-q*T)*NormCDF(-d1)
}

// ANSI color helper for heatmap cells
func colorizeBlock(val, minV, maxV float64) string {
	var t float64
	if maxV > minV {
		t = (val - minV) / (maxV - minV)
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
	} else {
		t = 0
	}
	// gradient from green (low) to red (high)
	r := int(255 * t)
	g := int(255 * (1 - t))
	b := 0
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm  \x1b[0m", r, g, b)
}

func renderColoredHeatmap(data [][]float64) []string {
	if len(data) == 0 || len(data[0]) == 0 {
		return []string{}
	}
	minV, maxV := data[0][0], data[0][0]
	for i := range data {
		for j := range data[i] {
			if data[i][j] < minV {
				minV = data[i][j]
			}
			if data[i][j] > maxV {
				maxV = data[i][j]
			}
		}
	}
	lines := make([]string, len(data))
	for i := range data {
		cells := make([]string, len(data[i]))
		for j := range data[i] {
			cells[j] = colorizeBlock(data[i][j], minV, maxV)
		}
		lines[i] = strings.Join(cells, "")
	}
	return lines
}

type model struct {
	S, K, T, r, q, v float64
	editing          bool
	input            string
	focus            int // 0..6; 6 is Compute heatmaps, 7 is Quit

	callData [][]float64
	putData  [][]float64
}

func initialModel() model {
	return model{S: 100.0, K: 100.0, T: 1.0, r: 0.05, q: 0.02, v: 0.2, editing: false, input: "", focus: 0}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			s := msg.String()
			if s == "enter" {
				if val, err := strconv.ParseFloat(m.input, 64); err == nil {
					switch m.focus {
					case 0:
						m.S = val
					case 1:
						m.K = val
					case 2:
						m.T = val
					case 3:
						m.r = val
					case 4:
						m.q = val
					case 5:
						m.v = val
					}
				}
				m.editing = false
				m.input = ""
				return m, nil
			}
			if s == "backspace" {
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
				return m, nil
			}
			if len(s) == 1 && strings.ContainsAny(s, "0123456789.-") {
				m.input += s
				return m, nil
			}
			return m, nil
		}
		switch msg.String() {
		case "up":
			if m.focus > 0 {
				m.focus--
			}
		case "down":
			if m.focus < 7 {
				m.focus++
			}
		case "enter":
			if m.focus < 6 {
				m.editing = true
				m.input = m.fieldValueAsString(m.focus)
			} else if m.focus == 6 {
				m.computeHeatmaps()
			} else {
				return m, tea.Quit
			}
		case "left":
		case "a":
			if m.v > 0.01 {
				m.v -= 0.01
				m.computeHeatmaps()
			}
		case "right":
		case "d":
			m.v += 0.01
			m.computeHeatmaps()
		case "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		// ignore
	}
	return m, nil
}

func (m model) fieldValueAsString(idx int) string {
	switch idx {
	case 0:
		return fmt.Sprintf("%g", m.S)
	case 1:
		return fmt.Sprintf("%g", m.K)
	case 2:
		return fmt.Sprintf("%g", m.T)
	case 3:
		return fmt.Sprintf("%g", m.r)
	case 4:
		return fmt.Sprintf("%g", m.q)
	case 5:
		return fmt.Sprintf("%g", m.v)
	default:
		return ""
	}
}

func (m *model) computeHeatmaps() {
	n := 20
	mgrid := 20
	Smin := m.S * 0.8
	Smax := m.S * 1.2
	Vmin := m.v * 0.8
	if Vmin < 0.01 {
		Vmin = 0.01
	}
	Vmax := m.v * 1.4

	call := make([][]float64, n)
	put := make([][]float64, n)
	for i := 0; i < n; i++ {
		call[i] = make([]float64, mgrid)
		put[i] = make([]float64, mgrid)
	}
	for i := 0; i < n; i++ {
		Si := Smin + float64(i)*(Smax-Smin)/float64(n-1)
		for j := 0; j < mgrid; j++ {
			vj := Vmin + float64(j)*(Vmax-Vmin)/float64(mgrid-1)
			call[i][j] = Calc(Si, m.K, m.T, m.r, m.q, vj, "call")
			put[i][j] = Calc(Si, m.K, m.T, m.r, m.q, vj, "put")
		}
	}
	m.callData = call
	m.putData = put
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString("Option Pricer TUI (Black-Scholes)\n")
	b.WriteString("Use Up/Down to select field, Enter to edit, Left/Right to adjust volatility.\n\n")
	fields := []string{
		fmt.Sprintf("S (spot): %g", m.S),
		fmt.Sprintf("K (strike): %g", m.K),
		fmt.Sprintf("T (maturity): %g", m.T),
		fmt.Sprintf("r (risk-free rate): %g", m.r),
		fmt.Sprintf("q (dividend yield): %g", m.q),
		fmt.Sprintf("v (volatility): %g", m.v),
	}
	for i, f := range fields {
		prefix := "  "
		if m.focus == i && !m.editing {
			prefix = "->"
		}
		if m.editing && m.focus == i {
			f = fmt.Sprintf("%s (editing: %s)", fields[i], m.input)
		}
		b.WriteString(fmt.Sprintf("%s %s\n", prefix, f))
	}
	prefix := "  "
	if m.focus == 6 {
		prefix = "->"
	}
	b.WriteString(fmt.Sprintf("%s Compute heatmaps (Enter)\n", prefix))
	// Axis descriptions for the heatmaps
	b.WriteString("X axis: v (volatility)   Y axis: S (spot price)\n\n")
	// Current prices above heatmaps
	currentCall := Calc(m.S, m.K, m.T, m.r, m.q, m.v, "call")
	currentPut := Calc(m.S, m.K, m.T, m.r, m.q, m.v, "put")
	b.WriteString(fmt.Sprintf("Current prices: Call = $%.2f, Put = $%.2f\n\n", currentCall, currentPut))
	// Heatmaps rendering
	if len(m.callData) > 0 && len(m.putData) > 0 {
		callLines := renderColoredHeatmap(m.callData)
		putLines := renderColoredHeatmap(m.putData)
		maxRows := len(callLines)
		if len(putLines) > maxRows {
			maxRows = len(putLines)
		}
		b.WriteString("Call Heatmap                              Put Heatmap\n")
		for i := 0; i < maxRows; i++ {
			cl := ""
			pp := ""
			if i < len(callLines) {
				cl = callLines[i]
			}
			if i < len(putLines) {
				pp = putLines[i]
			}
			b.WriteString(fmt.Sprintf("%-40s  %-40s\n", cl, pp))
		}
	}
	// Quit button
	prefixQuit := "  "
	if m.focus == 7 {
		prefixQuit = "->"
	}
	b.WriteString(fmt.Sprintf("%s Quit\n", prefixQuit))
	return b.String()
}

func (m *model) fieldShortName(idx int) string {
	switch idx {
	case 0:
		return "S"
	case 1:
		return "K"
	case 2:
		return "T"
	case 3:
		return "r"
	case 4:
		return "q"
	case 5:
		return "v"
	default:
		return ""
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
