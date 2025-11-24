package shell

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetInitScript returns the shell initialization script for the detected shell
func GetInitScript() (string, error) {
	shell := detectShell()
	gwPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	switch shell {
	case "bash", "zsh":
		return getBashZshInit(gwPath), nil
	case "fish":
		return getFishInit(gwPath), nil
	default:
		return getBashZshInit(gwPath), nil
	}
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "bash"
	}
	return filepath.Base(shell)
}

func getBashZshInit(gwPath string) string {
	return fmt.Sprintf(`gw() {
  local output
  if [ "$1" = "cd" ]; then
    output=$(%s cd "${@:2}")
    if [ $? -eq 0 ] && [ -n "$output" ]; then
      eval "$output"
    fi
  else
    %s "$@"
  fi
}`, gwPath, gwPath)
}

func getFishInit(gwPath string) string {
	return fmt.Sprintf(`function gw
  if test "$argv[1]" = "cd"
    set output (%s cd $argv[2..-1])
    if test $status -eq 0; and test -n "$output"
      eval "$output"
    end
  else
    %s $argv
  end
end`, gwPath, gwPath)
}
