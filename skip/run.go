package skip

import (
	"bytes"
	"io"
)

type runCondition struct {
	cmd    Command
	logger Logger
}

func (r runCondition) Match(_ func() GitState, item map[string]any) bool {
	commandLine, ok := item["run"].(string)
	if !ok {
		return false
	}

	return r.execute(commandLine)
}

func (r runCondition) execute(commandLine string) bool {
	if commandLine == "" {
		return false
	}

	if r.cmd == nil {
		return false
	}

	sh, err := shExecutable()
	if err != nil {
		if r.logger != nil {
			r.logger.Errorf("`sh` executable not found: %s\n", err)
		}
		return false
	}

	args := []string{sh, "-c", commandLine}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	err = r.cmd.Run(args, "", nullReader{}, stdout, stderr)
	return err == nil
}

type nullReader struct{}

func (nullReader) Read([]byte) (int, error) {
	return 0, io.EOF
}
