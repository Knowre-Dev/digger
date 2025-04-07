package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	net "net/http"
	"os"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/diggerhq/digger/backend/models"
	"github.com/diggerhq/digger/libs/ci"
	github2 "github.com/diggerhq/digger/libs/ci/github"
	"github.com/diggerhq/digger/libs/scheduler"
	"github.com/google/go-github/v61/github"
)

func createTempDir() string {
	tempDir, err := os.MkdirTemp("", "repo")
	if err != nil {
		slog.Error("Failed to create temporary directory", "error", err)
		panic(err)
	}
	return tempDir
}

type action func(string) error

func CloneGitRepoAndDoAction(repoUrl string, branch string, commitHash string, token string, tokenUsername string, action action) error {
	dir := createTempDir()

	slog.Debug("Cloning git repository",
		"repoUrl", repoUrl,
		"branch", branch,
		"commitHash", commitHash,
		"directory", dir,
	)

	git := NewGitShellWithTokenAuth(dir, token, tokenUsername)
	err := git.Clone(repoUrl, branch)
	if err != nil {
		slog.Error("Failed to clone repository",
			"repoUrl", repoUrl,
			"branch", branch,
			"error", err,
		)
		return err
	}

	if commitHash != "" {
		err := git.Checkout(commitHash)
		if err != nil {
			slog.Error("Failed to checkout commit",
				"commitHash", commitHash,
				"error", err,
			)
			return err
		}
	}

	defer func() {
		slog.Debug("Removing cloned directory", "directory", dir)
		ferr := os.RemoveAll(dir)
		if ferr != nil {
			slog.Warn("Failed to remove directory", "directory", dir, "error", ferr)
		}
	}()

	err = action(dir)
	if err != nil {
		slog.Error("Error performing action on repository", "directory", dir, "error", err)
		return err
	}

	return nil
}

// just a wrapper around github client to be able to use mocks
type DiggerGithubRealClientProvider struct {
}

type DiggerGithubClientMockProvider struct {
	MockedHTTPClient *net.Client
}

type GithubClientProvider interface {
	NewClient(netClient *net.Client) (*github.Client, error)
	Get(githubAppId int64, installationId int64) (*github.Client, *string, error)
	FetchCredentials(githubAppId string) (string, string, string, string, error)
}

func (gh DiggerGithubRealClientProvider) NewClient(netClient *net.Client) (*github.Client, error) {
	ghClient := github.NewClient(netClient)
	return ghClient, nil
}

func (gh DiggerGithubRealClientProvider) Get(githubAppId int64, installationId int64) (*github.Client, *string, error) {
	slog.Debug("Getting GitHub client",
		"githubAppId", githubAppId,
		"installationId", installationId,
	)

	githubAppPrivateKey := ""
	githubAppPrivateKeyB64 := os.Getenv("GITHUB_APP_PRIVATE_KEY_BASE64")
	if githubAppPrivateKeyB64 != "" {
		decodedBytes, err := base64.StdEncoding.DecodeString(githubAppPrivateKeyB64)
		if err != nil {
			slog.Error("Failed to decode GITHUB_APP_PRIVATE_KEY_BASE64", "error", err)
			return nil, nil, fmt.Errorf("error initialising github app installation: please set GITHUB_APP_PRIVATE_KEY_BASE64 env variable\n")
		}
		githubAppPrivateKey = string(decodedBytes)
	} else {
		githubAppPrivateKey = os.Getenv("GITHUB_APP_PRIVATE_KEY")
		if githubAppPrivateKey != "" {
			slog.Warn("GITHUB_APP_PRIVATE_KEY will be deprecated in future releases, please use GITHUB_APP_PRIVATE_KEY_BASE64 instead")
		} else {
			slog.Error("Missing GitHub app private key", "required", "GITHUB_APP_PRIVATE_KEY_BASE64")
			return nil, nil, fmt.Errorf("error initialising github app installation: please set GITHUB_APP_PRIVATE_KEY_BASE64 env variable\n")
		}
	}

	tr := net.DefaultTransport
	itr, err := ghinstallation.New(tr, githubAppId, installationId, []byte(githubAppPrivateKey))
	if err != nil {
		slog.Error("Failed to initialize GitHub app installation",
			"githubAppId", githubAppId,
			"installationId", installationId,
			"error", err,
		)
		return nil, nil, fmt.Errorf("error initialising github app installation: %v\n", err)
	}

	token, err := itr.Token(context.Background())
	if err != nil {
		slog.Error("Failed to get GitHub app token",
			"githubAppId", githubAppId,
			"installationId", installationId,
			"error", err,
		)
		return nil, nil, fmt.Errorf("error initialising git app token: %v\n", err)
	}

	ghClient, err := gh.NewClient(&net.Client{Transport: itr})
	if err != nil {
		slog.Error("Failed to create GitHub client", "error", err)
		return nil, nil, fmt.Errorf("error creating new client: %v", err)
	}

	slog.Debug("Successfully obtained GitHub client and token",
		"githubAppId", githubAppId,
		"installationId", installationId,
	)

	return ghClient, &token, nil
}

func (gh DiggerGithubRealClientProvider) FetchCredentials(githubAppId string) (string, string, string, string, error) {
	clientId := os.Getenv("GITHUB_APP_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_APP_CLIENT_SECRET")
	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	privateKeyb64 := os.Getenv("GITHUB_APP_PRIVATE_KEY_BASE64")
	return clientId, clientSecret, webhookSecret, privateKeyb64, nil
}

func (gh DiggerGithubClientMockProvider) NewClient(netClient *net.Client) (*github.Client, error) {
	ghClient := github.NewClient(gh.MockedHTTPClient)
	return ghClient, nil
}

func (gh DiggerGithubClientMockProvider) Get(githubAppId int64, installationId int64) (*github.Client, *string, error) {
	ghClient, _ := gh.NewClient(gh.MockedHTTPClient)
	token := "token"
	return ghClient, &token, nil
}

func (gh DiggerGithubClientMockProvider) FetchCredentials(githubAppId string) (string, string, string, string, error) {
	return "clientId", "clientSecret", "", "", nil
}

func GetGithubClient(gh GithubClientProvider, installationId int64, repoFullName string) (*github.Client, *string, error) {
	installation, err := models.DB.GetGithubAppInstallationByIdAndRepo(installationId, repoFullName)
	if err != nil {
		slog.Error("Failed to get GitHub installation",
			"installationId", installationId,
			"repoFullName", repoFullName,
			"error", err,
		)
		return nil, nil, fmt.Errorf("Error getting installation: %v", err)
	}

	ghClient, token, err := gh.Get(installation.GithubAppId, installation.GithubInstallationId)
	return ghClient, token, err
}

func GetGithubClientFromAppId(gh GithubClientProvider, installationId int64, githubAppId int64, repoFullName string) (*github.Client, *string, error) {
	ghClient, token, err := gh.Get(githubAppId, installationId)
	return ghClient, token, err
}

func GetGithubService(gh GithubClientProvider, installationId int64, repoFullName string, repoOwner string, repoName string) (*github2.GithubService, *string, error) {
	ghClient, token, err := GetGithubClient(gh, installationId, repoFullName)
	if err != nil {
		slog.Error("Failed to create GitHub client",
			"installationId", installationId,
			"repoFullName", repoFullName,
			"error", err,
		)
		return nil, nil, fmt.Errorf("Error creating github app client: %v", err)
	}

	ghService := github2.GithubService{
		Client:   ghClient,
		RepoName: repoName,
		Owner:    repoOwner,
	}

	slog.Debug("Created GitHub service",
		"owner", repoOwner,
		"repoName", repoName,
	)

	return &ghService, token, nil
}

func SetPRStatusForJobs(prService ci.PullRequestService, prNumber int, jobs []scheduler.Job) error {
	slog.Info("Setting PR status for jobs",
		"prNumber", prNumber,
		"jobCount", len(jobs),
	)

	for _, job := range jobs {
		for _, command := range job.Commands {
			var err error
			switch command {
			case "digger plan":
				slog.Debug("Setting PR status for plan",
					"prNumber", prNumber,
					"project", job.ProjectName,
				)
				err = prService.SetStatus(prNumber, "pending", job.ProjectName+"/plan")
			case "digger apply":
				slog.Debug("Setting PR status for apply",
					"prNumber", prNumber,
					"project", job.ProjectName,
				)
				err = prService.SetStatus(prNumber, "pending", job.ProjectName+"/apply")
			}
			if err != nil {
				slog.Error("Failed to set PR status",
					"prNumber", prNumber,
					"project", job.ProjectName,
					"command", command,
					"error", err,
				)
				return fmt.Errorf("Error setting pr status: %v", err)
			}
		}
	}

	// Report aggregate status for digger/plan or digger/apply
	if len(jobs) > 0 {
		var err error
		if scheduler.IsPlanJobs(jobs) {
			slog.Debug("Setting aggregate plan status", "prNumber", prNumber)
			err = prService.SetStatus(prNumber, "pending", "digger/plan")
		} else {
			slog.Debug("Setting aggregate apply status", "prNumber", prNumber)
			err = prService.SetStatus(prNumber, "pending", "digger/apply")
		}
		if err != nil {
			slog.Error("Failed to set aggregate PR status",
				"prNumber", prNumber,
				"error", err,
			)
			return fmt.Errorf("error setting pr status: %v", err)
		}
	} else {
		slog.Debug("Setting success status for empty job list", "prNumber", prNumber)

		err := prService.SetStatus(prNumber, "success", "digger/plan")
		if err != nil {
			slog.Error("Failed to set success plan status", "prNumber", prNumber, "error", err)
			return fmt.Errorf("error setting pr status: %v", err)
		}

		err = prService.SetStatus(prNumber, "success", "digger/apply")
		if err != nil {
			slog.Error("Failed to set success apply status", "prNumber", prNumber, "error", err)
			return fmt.Errorf("error setting pr status: %v", err)
		}
	}

	slog.Info("Successfully set PR status", "prNumber", prNumber)
	return nil
}

func GetGithubHostname() string {
	githubHostname := os.Getenv("DIGGER_GITHUB_HOSTNAME")
	if githubHostname == "" {
		githubHostname = "github.com"
	}
	return githubHostname
}

func GetWorkflowIdAndUrlFromDiggerJobId(client *github.Client, repoOwner string, repoName string, diggerJobID string) (int64, string, error) {
	slog.Debug("Looking for workflow for job",
		"diggerJobId", diggerJobID,
		slog.Group("repository",
			slog.String("owner", repoOwner),
			slog.String("name", repoName),
		),
	)

	timeFilter := time.Now().Add(-5 * time.Minute)
	runs, _, err := client.Actions.ListRepositoryWorkflowRuns(context.Background(), repoOwner, repoName, &github.ListWorkflowRunsOptions{
		Created: ">=" + timeFilter.Format(time.RFC3339),
	})
	if err != nil {
		slog.Error("Failed to list workflow runs",
			"repoOwner", repoOwner,
			"repoName", repoName,
			"error", err,
		)
		return 0, "#", fmt.Errorf("error listing workflow runs %v", err)
	}

	slog.Debug("Searching through workflow runs",
		"count", len(runs.WorkflowRuns),
		"timeFilter", timeFilter.Format(time.RFC3339),
	)

	for _, workflowRun := range runs.WorkflowRuns {
		workflowjobs, _, err := client.Actions.ListWorkflowJobs(context.Background(), repoOwner, repoName, *workflowRun.ID, nil)
		if err != nil {
			slog.Error("Failed to list workflow jobs",
				"workflowRunId", *workflowRun.ID,
				"error", err,
			)
			return 0, "#", fmt.Errorf("error listing workflow jobs for run %v %v", workflowRun.ID, err)
		}

		for _, workflowjob := range workflowjobs.Jobs {
			for _, step := range workflowjob.Steps {
				if strings.Contains(*step.Name, diggerJobID) {
					workflowUrl := fmt.Sprintf("https://%v/%v/%v/actions/runs/%v", GetGithubHostname(), repoOwner, repoName, *workflowRun.ID)

					slog.Info("Found workflow for job",
						"diggerJobId", diggerJobID,
						"workflowRunId", *workflowRun.ID,
						"workflowUrl", workflowUrl,
					)

					return *workflowRun.ID, workflowUrl, nil
				}
			}
		}
	}

	slog.Warn("No workflow found for job",
		"diggerJobId", diggerJobID,
		"repoOwner", repoOwner,
		"repoName", repoName,
	)

	return 0, "#", fmt.Errorf("workflow not found")
}
