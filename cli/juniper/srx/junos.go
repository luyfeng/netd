// NetD makes network device operations easy.
// Copyright (C) 2019  sky-cloud.net
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package srx

import (
	"fmt"
	"io"
	"regexp"

	"github.com/sky-cloud-tec/netd/cli"
	"golang.org/x/crypto/ssh"
)

func init() {
	// register srx 6.x
	cli.OperatorManagerInstance.Register(`(?i)juniper\.v?srx\..*`, createOpJunos())
}

type opJunos struct {
	lineBeak    string // \r\n \n
	transitions map[string][]string
	prompts     map[string][]*regexp.Regexp
	errs        []*regexp.Regexp
}

func createOpJunos() cli.Operator {
	loginPrompt := regexp.MustCompile("^[[:alnum:]_]{1,}[.]{0,1}[[:alnum:]_-]{0,}@[[:alnum:]._-]+> $")
	configPrompt := regexp.MustCompile("^[[:alnum:]_]{1,}[.]{0,1}[[:alnum:]_-]{0,}@[[:alnum:]._-]+# $")
	return &opJunos{
		// mode transition
		// login -> configure_private
		// login -> configure_exclusive
		// login -> configure
		transitions: map[string][]string{
			"login->configure_private":   {"configure private"},
			"configure_private->login":   {"exit"},
			"login->configure_exclusive": {"configure exclusive"},
			"configure_exclusive->login": {"exit"},
			"login->configure":           {"configure"},
			"configure->login":           {"exit"},
		},
		prompts: map[string][]*regexp.Regexp{
			"login":               {loginPrompt},
			"configure":           {configPrompt},
			"configure_private":   {configPrompt},
			"configure_exclusive": {configPrompt},
		},
		errs: []*regexp.Regexp{
			regexp.MustCompile("^syntax error\\.$"),
			regexp.MustCompile("^unknown command\\.$"),
			regexp.MustCompile("^missing argument\\.$"),
			regexp.MustCompile("\\^$"),
			regexp.MustCompile("^error:"),
		},
		lineBeak: "\n",
	}
}

func (s *opJunos) GetPrompts(k string) []*regexp.Regexp {
	if v, ok := s.prompts[k]; ok {
		return v
	}
	return nil
}
func (s *opJunos) GetTransitions(c, t string) []string {
	k := c + "->" + t
	if v, ok := s.transitions[k]; ok {
		return v
	}
	return nil
}

func (s *opJunos) GetErrPatterns() []*regexp.Regexp {
	return s.errs
}

func (s *opJunos) GetLinebreak() string {
	return s.lineBeak
}

func (s *opJunos) GetStartMode() string {
	return "login"
}

func (s *opJunos) GetSSHInitializer() cli.SSHInitializer {
	return func(c *ssh.Client) (io.Reader, io.WriteCloser, *ssh.Session, error) {
		var err error
		session, err := c.NewSession()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("new ssh session failed, %s", err)
		}
		// get stdout and stdin channel
		r, err := session.StdoutPipe()
		if err != nil {
			session.Close()
			return nil, nil, nil, fmt.Errorf("create stdout pipe failed, %s", err)
		}
		w, err := session.StdinPipe()
		if err != nil {
			session.Close()
			return nil, nil, nil, fmt.Errorf("create stdin pipe failed, %s", err)
		}
		if err := session.Shell(); err != nil {
			session.Close()
			return nil, nil, nil, fmt.Errorf("create shell failed, %s", err)
		}
		return r, w, session, nil
	}
}
