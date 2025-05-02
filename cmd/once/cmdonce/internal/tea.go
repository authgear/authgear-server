package internal

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/authgear/authgear-server/pkg/util/bubbleteautil"
)

type FatalError struct {
	Err error
}

var _ tea.Model = FatalError{}

func (m FatalError) Init() tea.Cmd {
	return nil
}

func (m FatalError) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m FatalError) View() string {
	if m.Err == nil {
		return ""
	}

	var b strings.Builder
	var errMsg string
	var actionableMsg string

	switch {
	case errors.Is(m.Err, ErrNoDocker):
		errMsg = fmt.Sprintf("%v is not installed on your machine.", BinDocker)
		actionableMsg = "Visit https://docs.docker.com/get-started/get-docker/ to install it."
	case errors.Is(m.Err, ErrDockerVolumeExists):
		errMsg = fmt.Sprintf("The docker volume %v exists already.", NameDockerVolume)
		actionableMsg = fmt.Sprintf("Either run `%v start` to start Authgear, or run `docker volume rm %v` to remove the volume (you will lose all data!).", ProgramName, NameDockerVolume)
	case errors.Is(m.Err, ErrDockerContainerNotExists):
		errMsg = fmt.Sprintf("The docker container %v does not exist. Maybe you did not run `%v setup` before?", NameDockerContainer, ProgramName)
		actionableMsg = fmt.Sprintf("Run `%v setup` to set up Authgear first.", ProgramName)
	case errors.Is(m.Err, ErrCommandUpgradeNotImplemented):
		errMsg = fmt.Sprintf("This command is not implemented yet.")
		actionableMsg = fmt.Sprintf("Re-run the oneliner in the email to upgrade this program.")
	case errors.Is(m.Err, ErrLicenseServerLicenseKeyNotFound):
		errMsg = fmt.Sprintf("The license key you entered is invalid.")
		actionableMsg = fmt.Sprintf("If you think this is an error, please contact us at once@authgear.com")
	case errors.Is(m.Err, ErrLicenseServerLicenseKeyAlreadyActivated):
		errMsg = fmt.Sprintf("The license key you entered has already been activated.")
		actionableMsg = fmt.Sprintf("If you think this is an error, please contact us at once@authgear.com")
	case errors.Is(m.Err, ErrLicenseServerLicenseKeyExpired):
		errMsg = fmt.Sprintf("The license key you entered is expired.")
		actionableMsg = fmt.Sprintf("If you think this is an error, please contact us at once@authgear.com")
	}

	if errMsg == "" || actionableMsg == "" {
		fmt.Fprintf(&b,
			"❌ Encountered this fatal error:\n\n  %v\n\n",
			bubbleteautil.StyleForegroundSemanticError.Render(m.Err.Error()),
		)
	} else {
		fmt.Fprintf(&b,
			"❌ Encountered this fatal error:\n\n  %v\n\nHere are some actions you may take:\n\n  %v\n\n",
			bubbleteautil.StyleForegroundSemanticError.Render(errMsg),
			bubbleteautil.StyleForegroundSemanticInfo.Render(actionableMsg),
		)
	}

	return b.String()
}

func (m FatalError) WithErr(err error) FatalError {
	m.Err = err
	return m
}
