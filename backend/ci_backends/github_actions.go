package ci_backends

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/diggerhq/digger/backend/utils"
	orchestrator_scheduler "github.com/diggerhq/digger/libs/scheduler"
	"github.com/diggerhq/digger/libs/spec"
	"github.com/google/go-github/v61/github"
)

type GithubActionCi struct {
	Client *github.Client
}

func (g GithubActionCi) TriggerWorkflow(spec spec.Spec, runName string, vcsToken string) error {
	slog.Info("TriggerGithubWorkflow", "repoOwner", spec.VCS.RepoOwner, "repoName", spec.VCS.RepoName, "commentId", spec.CommentId)
	client := g.Client
	specBytes, err := json.Marshal(spec)

	inputs := orchestrator_scheduler.WorkflowInput{
		Spec:    string(specBytes),
		RunName: runName,
	}

	_, err = client.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), spec.VCS.RepoOwner, spec.VCS.RepoName, spec.VCS.WorkflowFile, github.CreateWorkflowDispatchEventRequest{
		Ref:    spec.Job.Branch,
		Inputs: inputs.ToMap(),
	})

	return err
}

func (g GithubActionCi) GetWorkflowUrl(spec spec.Spec) (string, error) {
	if spec.JobId == "" {
		slog.Error("Cannot get workflow URL: JobId is empty")
		return "", fmt.Errorf("job ID is required to fetch workflow URL")
	}

	_, workflowRunUrl, err := utils.GetWorkflowIdAndUrlFromDiggerJobId(g.Client, spec.VCS.RepoOwner, spec.VCS.RepoName, spec.JobId)
	if err != nil {
		slog.Error("Error getting workflow ID from job", "error", err)
		return "", err
	} else {
		return workflowRunUrl, nil
	}
}
