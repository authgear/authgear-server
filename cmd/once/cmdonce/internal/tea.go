package internal

import (
	"bufio"
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
	var errTCPPortAlreadyListening *ErrTCPPortAlreadyListening
	var errCertbotFailedToGetCertificates *ErrCertbotFailedToGetCertificates

	switch {
	case errors.Is(m.Err, ErrNoDocker):
		errMsg = fmt.Sprintf("%v is not installed on your machine.", BinDocker)
		actionableMsg = "Visit https://docs.docker.com/get-started/get-docker/ to install it."
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
	case errors.As(m.Err, &errTCPPortAlreadyListening):
		errMsg = fmt.Sprintf("The port %v is already bound on your machine.", errTCPPortAlreadyListening.Port)
		actionableMsg = fmt.Sprintf("Maybe another service on your machine is listening on %v. You may need to stop that first.", errTCPPortAlreadyListening.Port)
	case errors.As(m.Err, &errCertbotFailedToGetCertificates):
		var errMsgBuf strings.Builder
		errMsgBuf.WriteString("Failed to request TLS certificates from Let's Encrypt for these domains:\n")
		for _, domain := range errCertbotFailedToGetCertificates.Domains {
			errMsgBuf.WriteString(fmt.Sprintf("- %v\n", domain))
		}
		errMsg = errMsgBuf.String()
		actionableMsg = fmt.Sprintf(`- Integration with Let's Encrypt is enabled by default. If you do not want this, you can run this command to turn it off:

  TODO

- Or, you may have made a typo in the domains. You can re-run this command to correct it.
- Or, you may not have set up the required DNS records. Please check that.
`)
	}

	if errMsg == "" || actionableMsg == "" {
		fmt.Fprintf(&b,
			"❌ Encountered this fatal error:\n\n  %v\n\n",
			bubbleteautil.StyleForegroundSemanticError.Render(m.Err.Error()),
		)
	} else {
		fmt.Fprintf(&b,
			"❌ Encountered this fatal error:\n\n%v\n\nHere are some actions you may take:\n\n%v\n\n",
			bubbleteautil.StyleForegroundSemanticError.Render(indentLines(errMsg, "  ")),
			bubbleteautil.StyleForegroundSemanticInfo.Render(indentLines(actionableMsg, "  ")),
		)
	}

	return b.String()
}

func (m FatalError) WithErr(err error) FatalError {
	m.Err = err
	return m
}

func indentLines(lines string, indent string) string {
	var out strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(lines))
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(&out, "%v%v\n", indent, line)
	}
	return out.String()
}
