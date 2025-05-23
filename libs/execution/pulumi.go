package execution

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type Pulumi struct {
	WorkingDir string
	Stack      string
}

func (pl Pulumi) Init(params []string, envs map[string]string) (string, string, error) {
	// TODO: there is no equivalent of "init" in pulumi world, lets do login instead
	stdout, stderr, _, err := pl.runPululmiCommand("install", true, envs, params...)
	if err != nil {
		return stdout, stderr, err
	}
	stdout, stderr, _, err = pl.runPululmiCommand("login", true, envs, params...)
	return stdout, stderr, err
}

func (pl Pulumi) Apply(params []string, plan *string, envs map[string]string) (string, string, error) {
	pl.selectStack()
	params = append(params, "--yes")
	if plan != nil {
		params = append(params, []string{"--plan", *plan}...)
	}
	stdout, stderr, _, err := pl.runPululmiCommand("up", true, envs, params...)
	return stdout, stderr, err
}

func (pl Pulumi) Plan(params []string, envs map[string]string, planArtefactFilePath string) (bool, string, string, error) {
	pl.selectStack()
	params = append(params, []string{"--save-plan", planArtefactFilePath}...)
	stdout, stderr, statusCode, err := pl.runPululmiCommand("preview", true, envs, params...)
	if err != nil && statusCode != 2 {
		return false, "", "", err
	}
	return statusCode == 2, stdout, stderr, nil
}

func (pl Pulumi) Show(params []string, envs map[string]string, planArtefactFilePath string) (string, string, error) {
	pl.selectStack()
	// TODO figure out how to avoid running a second plan (preview) here
	params = append(params, []string{"--json"}...)
	stdout, stderr, statusCode, err := pl.runPululmiCommand("preview", false, envs, params...)
	if err != nil && statusCode != 2 {
		return "", "", err
	}
	return stdout, stderr, nil
}

func (pl Pulumi) Destroy(params []string, envs map[string]string) (string, string, error) {
	pl.selectStack()
	params = append(params, "--yes")
	stdout, stderr, _, err := pl.runPululmiCommand("destroy", true, envs, params...)
	return stdout, stderr, err
}

func (pl Pulumi) selectStack() error {
	slog.Debug("Selecting Pulumi stack", "stack", pl.Stack)
	_, _, _, err := pl.runPululmiCommand("stack", true, make(map[string]string, 0), "select", pl.Stack)
	if err != nil {
		slog.Error("Failed to select Pulumi stack", "stack", pl.Stack, "error", err)
		return err
	}
	return nil
}

func (pl Pulumi) runPululmiCommand(command string, printOutputToStdout bool, envs map[string]string, arg ...string) (string, string, int, error) {
	args := []string{command}
	args = append(args, arg...)
	envs["PULUMI_CI"] = "true"
	expandedArgs := make([]string, 0)
	for _, p := range args {
		s := os.ExpandEnv(p)
		s = strings.TrimSpace(s)
		if s != "" {
			expandedArgs = append(expandedArgs, s)
		}
	}

	var mwout, mwerr io.Writer
	var stdout, stderr bytes.Buffer
	if printOutputToStdout {
		mwout = io.MultiWriter(os.Stdout, &stdout)
		mwerr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		mwout = io.Writer(&stdout)
		mwerr = io.Writer(&stderr)
	}

	cmd := exec.Command("pulumi", expandedArgs...)
	slog.Info("Running Pulumi command",
		slog.Group("command",
			"binary", "pulumi",
			"args", expandedArgs,
			"workingDir", pl.WorkingDir,
		),
	)
	cmd.Dir = pl.WorkingDir

	env := os.Environ()
	for k, v := range envs {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env
	cmd.Stdout = mwout
	cmd.Stderr = mwerr

	err := cmd.Run()

	// terraform plan can return 2 if there are changes to be applied, so we don't want to fail in that case
	if err != nil && cmd.ProcessState.ExitCode() != 2 {
		slog.Error("Pulumi command failed",
			"command", command,
			"args", expandedArgs,
			"exitCode", cmd.ProcessState.ExitCode(),
			"error", err,
			"stdout", stdout.String(),
			"stderr", stderr.String(),
		)
	}

	return stdout.String(), stderr.String(), cmd.ProcessState.ExitCode(), err
}

func (pl Pulumi) formatPulumiWorkspaces(list string) string {
	list = strings.TrimSpace(list)
	char_replace := strings.NewReplacer("*", "", "\n", ",", " ", "")
	list = char_replace.Replace(list)
	return list
}
