# option-pricer-tui
Allows user to enter in the five core values of the Black-Scholes model

Returns two heatmaps representing the profit/ loss of the call and put

User can edit any of the values and re-compute both heatmaps

# Instructions

1. Install Go 1.22 or later
This Program uses Go and requires a Golang installation, and only works on Linux and Mac. If you're on Windows, you'll need to use WSL. Make sure you install go in your Linux/WSL terminal, not your Windows terminal/UI. Here is one option:

The webi installer is the simplest way for most people. Just run this in your terminal:

- curl -sS https://webi.sh/golang | sh

Read the output of the command and follow any instructions.

2. Clone the Repo
- cd ~
- git clone https://github.com/JamesJLA/option-pricer-tui.git
- cd option-pricer-tui
- go build ./
- ./option-pricer-tui
