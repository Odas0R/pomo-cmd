package editor

import (
	"fmt"

	"github.com/odas0r/pomo-cmd/pkg/fs"
	"github.com/odas0r/pomo-cmd/pkg/shell"
)

func Edit(path string) error {
	if fs.NotExists(path) {
		return fmt.Errorf("file does not exist on given path")
	}

	return shell.Exec(fmt.Sprintf(`
    if [[ ! -S "$NVIM_SOCKET" ]]; then
      nvim "+${2}" --listen "$NVIM_SOCKET" "%s"
    else
      nvim --server "$NVIM_SOCKET" --remote-send "<C-\><C-N>:wincmd p | edit %s<CR>"
    fi
  `, path, path))
}

func Notify(text string) error {
	return shell.Exec(fmt.Sprintf(`
    if [[ -S "$NVIM_SOCKET" ]]; then
      nvim --server "$NVIM_SOCKET" --remote-send '<ESC>:lua vim.notify("%s", nil, {title="Pomodoro"})<CR>'
    fi
  `, text))
}

func NotifyByType(text string, ntype string) error {
	return shell.Exec(fmt.Sprintf(`
    if [[ -S "$NVIM_SOCKET" ]]; then
      nvim --server "$NVIM_SOCKET" --remote-send '<ESC>:lua vim.notify("%s", "%s", {title="Pomodoro"})<CR>'
    fi
  `, text, ntype))
}
