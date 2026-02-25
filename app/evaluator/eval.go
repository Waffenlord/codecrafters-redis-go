package evaluator

import (
	"bytes"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type cmdToExecute struct {
	cmd command.Builtin
	args []string
}


func EvalProgram(n parser.Node, s *storage.Storage) (string, error) {
	var buf bytes.Buffer
	switch v := n.(type) {
	case (parser.Array):
		cmdList := []*cmdToExecute{}
		for i := 0; i < len(v.Elements); i++ {
			currentTok := v.Elements[i]
			switch t := currentTok.(type) {
			case (parser.BulkString):
				cmd, isCmd := command.CommandMenu[strings.ToLower(t.Literal)]
				if isCmd {
					cmdToAdd := cmdToExecute{
						cmd: cmd,
						args: []string{},
					}
					cmdList = append(cmdList, &cmdToAdd)
				} else if len(cmdList) > 0 {
					latestCmd := cmdList[len(cmdList) - 1]
					latestCmd.args = append(latestCmd.args, t.Literal)
				}
			}
		}

		if len(cmdList) == 1 {
			currentCmd := cmdList[0]
			err := currentCmd.cmd(nil, &buf, currentCmd.args, s)
			if err != nil {
				return "", err 
			}
			if buf.Len() > 0 {
				return buf.String(), nil 
			}
			return command.FormatSimpleString("OK"), nil
		}
	}
	return command.FormatSimpleString("OK"), nil
}


