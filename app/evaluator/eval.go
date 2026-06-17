package evaluator

import (
	"bytes"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func EvalProgram(n parser.Node, s *storage.Storage, c *command.RedisClient) (string, error) {
	var buf bytes.Buffer
	switch v := n.(type) {
	case (parser.Array):
		cmdList := []*command.CmdToExecute{}
		for i := 0; i < len(v.Elements); i++ {
			currentTok := v.Elements[i]
			switch t := currentTok.(type) {
			case (parser.BulkString):
				cmd, isCmd := command.CommandMenu[strings.ToLower(t.Literal)]
				if isCmd {
					cmdToAdd := &command.CmdToExecute{
						Cmd:  cmd,
						Args: []string{},
					}
					cmdList = append(cmdList, cmdToAdd)
				} else if len(cmdList) > 0 {
					latestCmd := cmdList[len(cmdList)-1]
					latestCmd.Args = append(latestCmd.Args, t.Literal)
				}
			}
		}

		if len(cmdList) == 1 {
			currentCmd := cmdList[0]
			if c.InTransaction {
				c.TxQueue = append(c.TxQueue, currentCmd)
				return command.FormatSimpleString("QUEUED"), nil
			}
			err := currentCmd.Cmd(nil, &buf, currentCmd.Args, s, c)
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
