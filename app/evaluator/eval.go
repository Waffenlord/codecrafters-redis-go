package evaluator

import (
	"bytes"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func EvalProgram(n parser.Node, s *storage.Storage, c *command.RedisClient, config config.Config) (string, error) {
	var buf bytes.Buffer
	switch v := n.(type) {
	case (parser.Array):
		cmdList := []*command.CmdToExecute{}
		for i := 0; i < len(v.Elements); i++ {
			currentTok := v.Elements[i]
			switch t := currentTok.(type) {
			case (parser.BulkString):
				key := strings.ToLower(t.Literal)
				cmd, isCmd := command.CommandMenu[key]
				if isCmd {
					cmdToAdd := &command.CmdToExecute{
						Name: key,
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
			if c.InTransaction && currentCmd.Name != "exec" && currentCmd.Name != "discard" {
				c.TxQueue = append(c.TxQueue, currentCmd)
				return command.FormatSimpleString("QUEUED"), nil
			}
			err := currentCmd.Cmd(&command.CommandContext{Out: &buf, Args: currentCmd.Args, C: c, ServerConfig: config}, s)
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
