package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/diggerhq/digger/backend/middleware"
	"github.com/diggerhq/digger/backend/models"
	"github.com/diggerhq/digger/backend/services"
	"github.com/diggerhq/digger/backend/utils"
	"github.com/diggerhq/digger/libs/ci"
	"github.com/diggerhq/digger/libs/comment_utils/reporting"
	"github.com/diggerhq/digger/libs/digger_config"
	"github.com/diggerhq/digger/libs/iac_utils"
	orchestrator_scheduler "github.com/diggerhq/digger/libs/scheduler"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListProjects(c *gin.Context) {

}

func FindProjectsForRepo(c *gin.Context) {
	repo := c.Param("repo")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)
	if !exists {
		slog.Warn("Organisation ID not found in context", "repo", repo)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	var projects []models.Project

	err := models.DB.GormDB.Preload("Organisation").Preload("Repo").
		Joins("LEFT JOIN repos ON projects.repo_id = repos.id").
		Joins("LEFT JOIN organisations ON projects.organisation_id = organisations.id").
		Where("repos.name = ? AND projects.organisation_id = ?", repo, orgId).Find(&projects).Error
	if err != nil {
		slog.Error("Error fetching projects for repo",
			"repo", repo,
			"orgId", orgId,
			"error", err,
		)
		c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		return
	}

	slog.Info("Found projects for repository",
		"repo", repo,
		"orgId", orgId,
		"projectCount", len(projects),
	)

	response := make([]interface{}, 0)

	for _, p := range projects {
		jsonStruct := p.MapToJsonStruct()
		response = append(response, jsonStruct)
	}

	if err != nil {
		slog.Error("Error marshalling response",
			"repo", repo,
			"error", err,
		)
		c.String(http.StatusInternalServerError, "Unknown error occurred while marshalling response")
		return
	}

	c.JSON(http.StatusOK, response)
}

func FindProjectsForOrg(c *gin.Context) {
	requestedOrganisation := c.Param("organisationId")
	loggedInOrganisation, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if requestedOrganisation == "" {
		requestedOrganisation = fmt.Sprintf("%v", loggedInOrganisation)
		slog.Debug("Using logged in organisation as requested organisation",
			"orgId", requestedOrganisation,
		)
	}

	if !exists {
		slog.Warn("Organisation ID not found in context",
			"requestedOrganisation", requestedOrganisation,
		)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	slog.Debug("Finding projects for organisation",
		"requestedOrganisation", requestedOrganisation,
		"loggedInOrganisation", loggedInOrganisation,
	)

	var org models.Organisation
	err := models.DB.GormDB.Where("id = ?", requestedOrganisation).First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Info("Organisation not found",
				"orgId", requestedOrganisation,
			)
			c.String(http.StatusNotFound, "Could not find organisation: "+requestedOrganisation)
		} else {
			slog.Error("Error fetching organisation",
				"orgId", requestedOrganisation,
				"error", err,
			)
			c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		}
		return
	}

	if org.ID != loggedInOrganisation {
		slog.Warn("Unauthorized access attempt to organisation's projects",
			"requestedOrgId", org.ID,
			"loggedInOrgId", loggedInOrganisation,
		)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	var projects []models.Project

	err = models.DB.GormDB.Preload("Organisation").Preload("Repo").
		Joins("LEFT JOIN repos ON projects.repo_id = repos.id").
		Joins("LEFT JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", org.ID).
		Order("is_in_main_branch").
		Order("repos.repo_full_name").
		Order("name").
		Find(&projects).Error

	if err != nil {
		slog.Error("Error fetching projects for organisation",
			"orgId", org.ID,
			"error", err,
		)
		c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		return
	}

	slog.Info("Found projects for organisation",
		"orgId", org.ID,
		"projectCount", len(projects),
	)

	marshalledProjects := make([]interface{}, 0)

	for _, p := range projects {
		marshalled := p.MapToJsonStruct()
		marshalledProjects = append(marshalledProjects, marshalled)
	}

	response := make(map[string]interface{})
	response["projects"] = marshalledProjects

	if err != nil {
		slog.Error("Error marshalling response",
			"orgId", org.ID,
			"error", err,
		)
		c.String(http.StatusInternalServerError, "Unknown error occurred while marshalling response")
		return
	}

	c.JSON(http.StatusOK, response)
}

func ProjectDetails(c *gin.Context) {
	currentOrg, exists := c.Get(middleware.ORGANISATION_ID_KEY)
	projectIdStr := c.Param("project_id")

	if projectIdStr == "" {
		slog.Warn("Project ID not specified in request")
		c.String(http.StatusBadRequest, "ProjectId not specified")
		return
	}

	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		slog.Warn("Invalid Project ID format",
			"projectIdStr", projectIdStr,
			"error", err,
		)
		c.String(http.StatusBadRequest, "Invalid ProjectId")
		return
	}

	if !exists {
		slog.Warn("Organisation ID not found in context",
			"projectId", projectId,
		)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	slog.Debug("Getting project details",
		"projectId", projectId,
		"orgId", currentOrg,
	)

	var org models.Organisation
	err = models.DB.GormDB.Where("id = ?", currentOrg).First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Info("Organisation not found",
				"orgId", currentOrg,
			)
			c.String(http.StatusNotFound, fmt.Sprintf("Could not find organisation: %v", currentOrg))
		} else {
			slog.Error("Error fetching organisation",
				"orgId", currentOrg,
				"error", err,
			)
			c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		}
		return
	}

	project, err := models.DB.GetProject(uint(projectId))
	if err != nil {
		slog.Error("Could not fetch project",
			"projectId", projectId,
			"error", err,
		)
		c.String(http.StatusInternalServerError, "Could not fetch project")
		return
	}

	if project.OrganisationID != org.ID {
		slog.Warn("Unauthorized access attempt to project",
			"projectId", projectId,
			"projectOrgId", project.OrganisationID,
			"loggedInOrgId", org.ID,
		)
		c.String(http.StatusForbidden, "No access to this project")
		return
	}

	slog.Info("Successfully retrieved project details",
		"projectId", projectId,
		"projectName", project.Name,
		"repoName", project.Repo.Name,
	)

	c.JSON(http.StatusOK, project.MapToJsonStruct())
}

type CreateProjectRequest struct {
	Name              string `json:"name"`
	ConfigurationYaml string `json:"configurationYaml"`
}

func ReportProjectsForRepo(c *gin.Context) {
	var request CreateProjectRequest
	err := c.BindJSON(&request)
	if err != nil {
		slog.Error("Error binding JSON request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	repoName := c.Param("repo")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		slog.Warn("Organisation ID not found in context", "repoName", repoName)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	slog.Debug("Processing project creation request",
		"repoName", repoName,
		"orgId", orgId,
		"projectName", request.Name,
	)

	org, err := models.DB.GetOrganisationById(orgId)
	if err != nil {
		slog.Error("Error fetching organisation",
			"orgId", orgId,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching organisation"})
		return
	}

	var repo models.Repo

	err = models.DB.GormDB.Where("name = ? AND organisation_id = ?", repoName, orgId).First(&repo).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Info("Repository not found, creating new one",
				"repoName", repoName,
				"orgId", orgId,
			)

			repo := models.Repo{
				Name:           repoName,
				OrganisationID: org.ID,
				Organisation:   org,
			}

			err = models.DB.GormDB.Create(&repo).Error

			if err != nil {
				slog.Error("Error creating repository",
					"repoName", repoName,
					"orgId", orgId,
					"error", err,
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating repo"})
				return
			}

			slog.Info("Successfully created repository",
				"repoId", repo.ID,
				"repoName", repoName,
			)
		} else {
			slog.Error("Error fetching repository",
				"repoName", repoName,
				"orgId", orgId,
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching repo"})
			return
		}
	}

	var project models.Project

	err = models.DB.GormDB.Where("name = ? AND organisation_id = ? AND repo_id = ?", request.Name, org.ID, repo.ID).First(&project).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Info("Project not found, creating new one",
				"projectName", request.Name,
				"repoName", repoName,
				"orgId", orgId,
			)

			project := models.Project{
				Name:              request.Name,
				ConfigurationYaml: request.ConfigurationYaml,
				RepoID:            repo.ID,
				OrganisationID:    org.ID,
				Repo:              &repo,
				Organisation:      org,
			}

			err = models.DB.GormDB.Create(&project).Error

			if err != nil {
				slog.Error("Error creating project",
					"projectName", request.Name,
					"repoName", repoName,
					"error", err,
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating project"})
				return
			}

			slog.Info("Successfully created project",
				"projectId", project.ID,
				"projectName", project.Name,
				"repoName", repoName,
			)

			c.JSON(http.StatusOK, project.MapToJsonStruct())
		} else {
			slog.Error("Error fetching project",
				"projectName", request.Name,
				"repoName", repoName,
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching project"})
			return
		}
	} else {
		slog.Info("Project already exists",
			"projectId", project.ID,
			"projectName", project.Name,
			"repoName", repoName,
		)
		c.JSON(http.StatusOK, project.MapToJsonStruct())
	}
}

func RunHistoryForProject(c *gin.Context) {
	repoName := c.Param("repo")
	projectName := c.Param("project")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		slog.Warn("Organisation ID not found in context",
			"repoName", repoName,
			"projectName", projectName,
		)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	slog.Debug("Fetching run history for project",
		"repoName", repoName,
		"projectName", projectName,
		"orgId", orgId,
	)

	org, err := models.DB.GetOrganisationById(orgId)
	if err != nil {
		slog.Error("Error fetching organisation",
			"orgId", orgId,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching organisation"})
		return
	}

	var repo models.Repo

	err = models.DB.GormDB.Where("name = ? AND organisation_id = ?", repoName, orgId).First(&repo).Error

	if err != nil {
		slog.Error("Error fetching repository",
			"repoName", repoName,
			"orgId", orgId,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching repo"})
		return
	}

	var project models.Project

	err = models.DB.GormDB.Where("name = ? AND repo_id = ? AND organisation_id = ?", projectName, repo.ID, org.ID).First(&project).Error

	if err != nil {
		slog.Error("Error fetching project",
			"projectName", projectName,
			"repoName", repoName,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching project"})
		return
	}

	var runHistory []models.ProjectRun

	err = models.DB.GormDB.Where("project_id = ?", project.ID).Find(&runHistory).Error

	if err != nil {
		slog.Error("Error fetching run history",
			"projectId", project.ID,
			"projectName", projectName,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching run history"})
		return
	}

	slog.Info("Successfully retrieved run history",
		"projectName", projectName,
		"repoName", repoName,
		"runCount", len(runHistory),
	)

	response := make([]interface{}, 0)

	for _, r := range runHistory {
		response = append(response, r.MapToJsonStruct())
	}

	c.JSON(http.StatusOK, response)
}

type SetJobStatusRequest struct {
	Status          string                      `json:"status"`
	Timestamp       time.Time                   `json:"timestamp"`
	JobSummary      *iac_utils.IacSummary       `json:"job_summary"`
	Footprint       *iac_utils.IacPlanFootprint `json:"job_plan_footprint"`
	PrCommentUrl    string                      `json:"pr_comment_url"`
	TerraformOutput string                      `json:"terraform_output"`
}

func (d DiggerController) SetJobStatusForProject(c *gin.Context) {
	jobId := c.Param("jobId")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		slog.Warn("Organisation ID not found in context", "jobId", jobId)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	slog.Info("Setting job status", "jobId", jobId, "orgId", orgId)

	var request SetJobStatusRequest
	err := c.BindJSON(&request)

	if err != nil {
		slog.Error("Error binding JSON request", "jobId", jobId, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error binding JSON"})
		return
	}

	job, err := models.DB.GetDiggerJob(jobId)
	if err != nil {
		slog.Error("Error fetching job", "jobId", jobId, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching job"})
		return
	}

	slog.Info("Processing job status update",
		"jobId", jobId,
		"currentStatus", job.Status,
		"newStatus", request.Status,
		"batchId", job.BatchID,
	)

	switch request.Status {
	case "started":
		job.Status = orchestrator_scheduler.DiggerJobStarted
		err := models.DB.UpdateDiggerJob(job)
		if err != nil {
			slog.Error("Error updating job status",
				"jobId", jobId,
				"status", request.Status,
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job status"})
			return
		}

		slog.Info("Job status updated to started", "jobId", jobId)

	case "succeeded":
		job.Status = orchestrator_scheduler.DiggerJobSucceeded
		job.TerraformOutput = request.TerraformOutput
		if request.Footprint != nil {
			job.PlanFootprint, err = json.Marshal(request.Footprint)
			if err != nil {
				slog.Error("Error marshalling plan footprint",
					"jobId", jobId,
					"error", err,
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marshalling plan footprint"})
				return
			}

			slog.Debug("Added plan footprint to job",
				"jobId", jobId,
				"footprintSize", len(job.PlanFootprint),
			)
		}

		job.PRCommentUrl = request.PrCommentUrl
		err := models.DB.UpdateDiggerJob(job)
		if err != nil {
			slog.Error("Error updating job",
				"jobId", jobId,
				"status", request.Status,
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving job"})
			return
		}

		slog.Info("Job status updated to succeeded",
			"jobId", jobId,
			"batchId", job.BatchID,
		)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					stack := string(debug.Stack())
					slog.Error("Recovered from panic in job completion handler",
						"jobId", jobId,
						"error", r,
						slog.Group("stack",
							slog.String("trace", stack),
						),
					)
				}
			}()

			slog.Debug("Starting post-success job processing", "jobId", jobId)

			ghClientProvider := d.GithubClientProvider
			installationLink, err := models.DB.GetGithubInstallationLinkForOrg(orgId)
			if err != nil {
				slog.Error("Error fetching installation link",
					"orgId", orgId,
					"jobId", jobId,
					"error", err,
				)
				return
			}

			installations, err := models.DB.GetGithubAppInstallations(installationLink.GithubInstallationId)
			if err != nil {
				slog.Error("Error fetching installations",
					"installationId", installationLink.GithubInstallationId,
					"jobId", jobId,
					"error", err,
				)
				return
			}

			if len(installations) == 0 {
				slog.Warn("No installations found",
					"installationId", installationLink.GithubInstallationId,
					"jobId", jobId,
				)
				return
			}

			jobLink, err := models.DB.GetDiggerJobLink(jobId)
			if err != nil {
				slog.Error("Error fetching job link",
					"jobId", jobId,
					"error", err,
				)
				return
			}

			workflowFileName := "digger_workflow.yml"

			if !strings.Contains(jobLink.RepoFullName, "/") {
				slog.Error("Invalid repo full name format",
					"repoFullName", jobLink.RepoFullName,
					"jobId", jobId,
				)
				return
			}

			repoFullNameSplit := strings.Split(jobLink.RepoFullName, "/")
			client, _, err := ghClientProvider.Get(installations[0].GithubAppId, installationLink.GithubInstallationId)
			if err != nil {
				slog.Error("Error getting GitHub client",
					"appId", installations[0].GithubAppId,
					"installationId", installationLink.GithubInstallationId,
					"jobId", jobId,
					"error", err,
				)
				return
			}

			slog.Info("Handling job completion",
				"jobId", jobId,
				"repoFullName", jobLink.RepoFullName,
				"batchId", job.BatchID,
			)

			err = services.DiggerJobCompleted(
				client,
				&job.Batch.ID,
				job,
				jobLink.RepoFullName,
				repoFullNameSplit[0],
				repoFullNameSplit[1],
				workflowFileName,
				d.GithubClientProvider,
			)
			if err != nil {
				slog.Error("Error handling job completion",
					"jobId", jobId,
					"error", err,
				)
				return
			}

			slog.Debug("Successfully processed job completion", "jobId", jobId)
		}()

		// store digger job summary
		if request.JobSummary != nil {
			models.DB.UpdateDiggerJobSummary(job.DiggerJobID, request.JobSummary.ResourcesCreated, request.JobSummary.ResourcesUpdated, request.JobSummary.ResourcesDeleted)
		}

	case "failed":
		job.Status = orchestrator_scheduler.DiggerJobFailed
		job.TerraformOutput = request.TerraformOutput
		err := models.DB.UpdateDiggerJob(job)
		if err != nil {
			slog.Error("Error updating job status",
				"jobId", jobId,
				"status", request.Status,
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving job"})
			return
		}

		slog.Info("Job status updated to failed",
			"jobId", jobId,
			"batchId", job.BatchID,
		)

	default:
		slog.Warn("Unexpected job status received",
			"jobId", jobId,
			"status", request.Status,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unexpected job status"})
		return
	}

	// attempt to update workflow run url
	slog.Debug("Attempting to update workflow run URL", "jobId", jobId)

	client, _, err := utils.GetGithubClient(d.GithubClientProvider, job.Batch.GithubInstallationId, job.Batch.RepoFullName)
	if err != nil {
		slog.Warn("Error creating GitHub client for workflow URL update",
			"jobId", jobId,
			"repoFullName", job.Batch.RepoFullName,
			"error", err,
		)
	} else {
		_, workflowRunUrl, err := utils.GetWorkflowIdAndUrlFromDiggerJobId(
			client,
			job.Batch.RepoOwner,
			job.Batch.RepoName,
			job.DiggerJobID,
		)
		if err != nil {
			slog.Warn("Error getting workflow ID from job",
				"jobId", jobId,
				"error", err,
			)
		} else if workflowRunUrl != "#" && workflowRunUrl != "" {
			job.WorkflowRunUrl = &workflowRunUrl
			err = models.DB.UpdateDiggerJob(job)
			if err != nil {
				slog.Error("Error updating job with workflow URL",
					"jobId", jobId,
					"workflowUrl", workflowRunUrl,
					"error", err,
				)
			} else {
				slog.Debug("Updated job with workflow URL",
					"jobId", jobId,
					"workflowUrl", workflowRunUrl,
				)
			}
		}
	}

	job.StatusUpdatedAt = request.Timestamp
	err = models.DB.GormDB.Save(&job).Error
	if err != nil {
		slog.Error("Error saving job status timestamp",
			"jobId", jobId,
			"timestamp", request.Timestamp,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving job"})
		return
	}

	// get batch ID
	// check if all jobs have succeeded at this point
	// if so, perform merge of PR (if configured to do so)
	batch := job.Batch

	slog.Info("Updating batch status after job update",
		"batchId", batch.ID,
		"jobId", jobId,
	)

	err = models.DB.UpdateBatchStatus(batch)
	if err != nil {
		slog.Error("Error updating batch status",
			"batchId", batch.ID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating batch status"})
		return
	}

	if batch.ReportTerraformOutputs {
		slog.Info("Generating Terraform outputs summary", "batchId", batch.ID)
		err = CreateTerraformOutputsSummary(d.GithubClientProvider, batch)
		if err != nil {
			slog.Warn("Could not generate Terraform outputs summary",
				"batchId", batch.ID,
				"error", err,
			)
		}
	}

	slog.Info("Checking if PR should be auto-merged", "batchId", batch.ID)
	err = AutomergePRforBatchIfEnabled(d.GithubClientProvider, batch)
	if err != nil {
		slog.Warn("Error auto-merging PR",
			"batchId", batch.ID,
			"prNumber", batch.PrNumber,
			"error", err,
		)

		err = utils.PostCommentForBatch(job.Batch, fmt.Sprintf(":yellow_circle: Warning could not automerge PR, please remember to merge it manually. Error: (%v)", err), d.GithubClientProvider)
		if err != nil {
			slog.Error("Error posting comment about automerge failure",
				"batchId", batch.ID,
				"error", err,
			)
		}
	}

	// return batch summary to client
	res, err := batch.MapToJsonStruct()
	if err != nil {
		slog.Error("Error getting batch details",
			"batchId", batch.ID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting batch details"})
		return
	}

	slog.Debug("Updating comments for batch", "batchId", batch.ID)
	UpdateCommentsForBatchGroup(d.GithubClientProvider, batch, res.Jobs)

	slog.Info("Successfully processed job status update",
		"jobId", jobId,
		"status", request.Status,
		"batchId", batch.ID,
	)

	c.JSON(http.StatusOK, res)
}

type CreateProjectRunRequest struct {
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`
	Status    string    `json:"status"`
	Command   string    `json:"command"`
	Output    string    `json:"output"`
}

func CreateRunForProject(c *gin.Context) {
	repoName := c.Param("repo")
	projectName := c.Param("projectName")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		slog.Warn("Organisation ID not found in context",
			"repoName", repoName,
			"projectName", projectName,
		)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	slog.Info("Creating run for project",
		"repoName", repoName,
		"projectName", projectName,
		"orgId", orgId,
	)

	org, err := models.DB.GetOrganisationById(orgId)
	if err != nil {
		slog.Error("Error fetching organisation",
			"orgId", orgId,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching organisation"})
		return
	}

	var repo models.Repo
	err = models.DB.GormDB.Where("name = ? AND organisation_id = ?", repoName, orgId).First(&repo).Error
	if err != nil {
		slog.Error("Error fetching repository",
			"repoName", repoName,
			"orgId", orgId,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching repo"})
		return
	}

	var project models.Project
	err = models.DB.GormDB.Where("name = ? AND repo_id = ? AND organisation_id = ?", projectName, repo.ID, org.ID).First(&project).Error
	if err != nil {
		slog.Error("Error fetching project",
			"projectName", projectName,
			"repoName", repoName,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching project"})
		return
	}

	slog.Debug("Found project",
		"projectId", project.ID,
		"projectName", projectName,
		"repoName", repoName,
	)

	var request CreateProjectRunRequest
	err = c.BindJSON(&request)
	if err != nil {
		slog.Error("Error binding JSON request",
			"projectName", projectName,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error binding JSON"})
		return
	}

	run := models.ProjectRun{
		StartedAt: request.StartedAt.UnixMilli(),
		EndedAt:   request.EndedAt.UnixMilli(),
		Status:    request.Status,
		Command:   request.Command,
		Output:    request.Output,
		ProjectID: project.ID,
		Project:   &project,
	}

	slog.Debug("Creating project run",
		"projectId", project.ID,
		"status", request.Status,
		"command", request.Command,
		"startedAt", request.StartedAt,
		"endedAt", request.EndedAt,
	)

	err = models.DB.GormDB.Create(&run).Error
	if err != nil {
		slog.Error("Error creating run",
			"projectName", projectName,
			"projectId", project.ID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating run"})
		return
	}

	slog.Info("Successfully created run for project",
		"runId", run.ID,
		"projectName", projectName,
		"projectId", project.ID,
		"status", run.Status,
	)

	c.JSON(http.StatusOK, "")
}

func UpdateCommentsForBatchGroup(gh utils.GithubClientProvider, batch *models.DiggerBatch, serializedJobs []orchestrator_scheduler.SerializedJob) error {
	slog.Debug("Updating comments for batch group",
		"batchId", batch.ID,
		"prNumber", batch.PrNumber,
		"jobCount", len(serializedJobs),
	)

	diggerYmlString := batch.DiggerConfig
	diggerConfigYml, err := digger_config.LoadDiggerConfigYamlFromString(diggerYmlString)
	if err != nil {
		slog.Error("Error loading digger config from batch",
			"batchId", batch.ID,
			"error", err,
		)
		return fmt.Errorf("error loading digger config from batch: %v", err)
	}

	if diggerConfigYml.CommentRenderMode == nil ||
		*diggerConfigYml.CommentRenderMode != digger_config.CommentRenderModeGroupByModule {
		slog.Debug("Render mode is not group_by_module, skipping comment updates",
			"batchId", batch.ID,
			"renderMode", diggerConfigYml.CommentRenderMode,
		)
		return nil
	}

	if batch.BatchType != orchestrator_scheduler.DiggerCommandPlan && batch.BatchType != orchestrator_scheduler.DiggerCommandApply {
		slog.Debug("Command is not plan or apply, skipping comment updates",
			"batchId", batch.ID,
			"batchType", batch.BatchType,
		)
		return nil
	}

	slog.Debug("Getting GitHub service for batch",
		"batchId", batch.ID,
		"installationId", batch.GithubInstallationId,
		"repoFullName", batch.RepoFullName,
	)

	ghService, _, err := utils.GetGithubService(
		gh,
		batch.GithubInstallationId,
		batch.RepoFullName,
		batch.RepoOwner,
		batch.RepoName,
	)
	if err != nil {
		slog.Error("Error getting GitHub service",
			"batchId", batch.ID,
			"repoFullName", batch.RepoFullName,
			"error", err,
		)
		return fmt.Errorf("error getting GitHub service: %v", err)
	}

	var sourceDetails []reporting.SourceDetails
	err = json.Unmarshal(batch.SourceDetails, &sourceDetails)
	if err != nil {
		slog.Error("Failed to unmarshal source details",
			"batchId", batch.ID,
			"error", err,
		)
		return fmt.Errorf("failed to unmarshal sourceDetails: %v", err)
	}

	slog.Debug("Building project to terraform output map",
		"batchId", batch.ID,
		"jobCount", len(serializedJobs),
	)

	// project_name => terraform output
	projectToTerraformOutput := make(map[string]string)
	// TODO: add projectName as a field of Job
	for _, serialJob := range serializedJobs {
		job, err := models.DB.GetDiggerJob(serialJob.DiggerJobId)
		if err != nil {
			slog.Error("Could not get digger job",
				"jobId", serialJob.DiggerJobId,
				"error", err,
			)
			return fmt.Errorf("could not get digger job: %v", err)
		}
		projectToTerraformOutput[serialJob.ProjectName] = job.TerraformOutput

		slog.Debug("Added terraform output for project",
			"projectName", serialJob.ProjectName,
			"jobId", serialJob.DiggerJobId,
			"outputLength", len(job.TerraformOutput),
		)
	}

	slog.Info("Updating source-based comments",
		"batchId", batch.ID,
		"sourceDetailsCount", len(sourceDetails),
		"prNumber", batch.PrNumber,
	)

	for _, detail := range sourceDetails {
		slog.Debug("Updating comment for source location",
			"sourceLocation", detail.SourceLocation,
			"commentId", detail.CommentId,
		)

		reporter := reporting.SourceGroupingReporter{serializedJobs, batch.PrNumber, ghService}
		err := reporter.UpdateComment(sourceDetails, detail.SourceLocation, projectToTerraformOutput)
		if err != nil {
			slog.Warn("Error updating comment for source location",
				"sourceLocation", detail.SourceLocation,
				"commentId", detail.CommentId,
				"error", err,
			)
		}
	}

	slog.Info("Successfully updated comments for batch group",
		"batchId", batch.ID,
		"prNumber", batch.PrNumber,
	)

	return nil
}

func GetPrServiceFromBatch(batch *models.DiggerBatch, gh utils.GithubClientProvider) (ci.PullRequestService, error) {
	slog.Debug("Getting PR service for batch",
		"batchId", batch.ID,
		"vcs", batch.VCS,
		"prNumber", batch.PrNumber,
	)

	switch batch.VCS {
	case "github":
		slog.Debug("Using GitHub service for batch",
			"batchId", batch.ID,
			"installationId", batch.GithubInstallationId,
			"repoFullName", batch.RepoFullName,
		)

		service, _, err := utils.GetGithubService(
			gh,
			batch.GithubInstallationId,
			batch.RepoFullName,
			batch.RepoOwner,
			batch.RepoName,
		)

		if err != nil {
			slog.Error("Error getting GitHub service",
				"batchId", batch.ID,
				"repoFullName", batch.RepoFullName,
				"error", err,
			)
		} else {
			slog.Debug("Successfully got GitHub service",
				"batchId", batch.ID,
				"repoFullName", batch.RepoFullName,
			)
		}

		return service, err

	case "gitlab":
		slog.Debug("Using GitLab service for batch",
			"batchId", batch.ID,
			"projectId", batch.GitlabProjectId,
			"repoFullName", batch.RepoFullName,
		)

		service, err := utils.GetGitlabService(
			utils.GitlabClientProvider{},
			batch.GitlabProjectId,
			batch.RepoName,
			batch.RepoFullName,
			batch.PrNumber,
			"",
		)

		if err != nil {
			slog.Error("Error getting GitLab service",
				"batchId", batch.ID,
				"projectId", batch.GitlabProjectId,
				"error", err,
			)
		} else {
			slog.Debug("Successfully got GitLab service",
				"batchId", batch.ID,
				"projectId", batch.GitlabProjectId,
			)
		}

		return service, err
	}

	slog.Error("Unsupported VCS type", "vcs", batch.VCS, "batchId", batch.ID)
	return nil, fmt.Errorf("could not retrieve a service for %v", batch.VCS)
}

func CreateTerraformOutputsSummary(gh utils.GithubClientProvider, batch *models.DiggerBatch) error {
	slog.Info("Creating Terraform outputs summary",
		"batchId", batch.ID,
		"prNumber", batch.PrNumber,
		"batchStatus", batch.Status,
	)

	diggerYmlString := batch.DiggerConfig
	diggerConfigYml, err := digger_config.LoadDiggerConfigYamlFromString(diggerYmlString)
	if err != nil {
		slog.Error("Error loading Digger config from batch",
			"batchId", batch.ID,
			"error", err,
		)
		return fmt.Errorf("error loading digger config from batch: %v", err)
	}

	config, _, err := digger_config.ConvertDiggerYamlToConfig(diggerConfigYml)
	if err != nil {
		slog.Error("Error converting Digger YAML to config",
			"batchId", batch.ID,
			"error", err,
		)
		return fmt.Errorf("error converting Digger YAML to config: %v", err)
	}

	if batch.Status == orchestrator_scheduler.BatchJobSucceeded && config.Reporting.AiSummary == true {
		slog.Info("Batch succeeded and AI summary enabled, generating summary",
			"batchId", batch.ID,
			"prNumber", batch.PrNumber,
		)

		prService, err := GetPrServiceFromBatch(batch, gh)
		if err != nil {
			slog.Error("Error getting PR service",
				"batchId", batch.ID,
				"error", err,
			)
			updateErr := prService.EditComment(batch.PrNumber, batch.AiSummaryCommentId,
				":x: could not generate AI summary \n\n could not communicate with github")
			if updateErr != nil {
				slog.Error("Failed to update comment with error message",
					"commentId", batch.AiSummaryCommentId,
					"error", updateErr,
				)
			}
			return fmt.Errorf("error getting github service: %v", err)
		}

		if batch.AiSummaryCommentId == "" {
			slog.Error("AI summary comment ID not found", "batchId", batch.ID)
			_, err := prService.PublishComment(batch.PrNumber,
				":x: could not generate AI summary \n\n could not communicate with github")
			if err != nil {
				slog.Error("Failed to publish error comment",
					"prNumber", batch.PrNumber,
					"error", err,
				)
			}
			return fmt.Errorf("could not post summary comment, initial comment not found")
		}

		summaryEndpoint := os.Getenv("DIGGER_AI_SUMMARY_ENDPOINT")
		if summaryEndpoint == "" {
			slog.Error("AI summary endpoint not configured", "batchId", batch.ID)
			updateErr := prService.EditComment(batch.PrNumber, batch.AiSummaryCommentId,
				":x: could not generate AI summary \n\n AI summary endpoint not configured")
			if updateErr != nil {
				slog.Error("Failed to update comment with error message",
					"commentId", batch.AiSummaryCommentId,
					"error", updateErr,
				)
			}
			return fmt.Errorf("could not generate AI summary, ai summary endpoint missing")
		}
		apiToken := os.Getenv("DIGGER_AI_SUMMARY_API_TOKEN")

		slog.Debug("Fetching jobs for batch", "batchId", batch.ID)
		jobs, err := models.DB.GetDiggerJobsForBatch(batch.ID)
		if err != nil {
			slog.Error("Could not get jobs for batch",
				"batchId", batch.ID,
				"error", err,
			)
			updateErr := prService.EditComment(batch.PrNumber, batch.AiSummaryCommentId,
				":x: could not generate AI summary \n\n error fetching jobs")
			if updateErr != nil {
				slog.Error("Failed to update comment with error message",
					"commentId", batch.AiSummaryCommentId,
					"error", updateErr,
				)
			}
			return fmt.Errorf("could not get jobs for batch: %v", err)
		}

		slog.Info("Collecting Terraform outputs from jobs",
			"batchId", batch.ID,
			"jobCount", len(jobs),
		)

		var terraformOutputs = ""
		for _, job := range jobs {
			var jobSpec orchestrator_scheduler.JobJson
			err := json.Unmarshal(job.SerializedJobSpec, &jobSpec)
			if err != nil {
				slog.Error("Could not unmarshal job spec",
					"jobId", job.DiggerJobID,
					"error", err,
				)
				updateErr := prService.EditComment(batch.PrNumber, batch.AiSummaryCommentId,
					":x: could not generate AI summary \n\n error fetching job spec")
				if updateErr != nil {
					slog.Error("Failed to update comment with error message",
						"commentId", batch.AiSummaryCommentId,
						"error", updateErr,
					)
				}
				return fmt.Errorf("could not summarize plans due to unmarshalling error: %v", err)
			}

			projectName := jobSpec.ProjectName
			slog.Debug("Adding Terraform output for project",
				"projectName", projectName,
				"jobId", job.DiggerJobID,
				"outputLength", len(job.TerraformOutput),
			)

			terraformOutputs += fmt.Sprintf("terraform output for %v \n\n", projectName) + job.TerraformOutput
		}

		slog.Info("Generating AI summary from Terraform outputs",
			"batchId", batch.ID,
			"outputsLength", len(terraformOutputs),
		)

		summary, err := utils.GetAiSummaryFromTerraformPlans(terraformOutputs, summaryEndpoint, apiToken)
		if err != nil {
			slog.Error("Could not generate AI summary from Terraform outputs",
				"batchId", batch.ID,
				"error", err,
			)
			updateErr := prService.EditComment(batch.PrNumber, batch.AiSummaryCommentId,
				":x: could not generate AI summary \n\n error generating summary from plans")
			if updateErr != nil {
				slog.Error("Failed to update comment with error message",
					"commentId", batch.AiSummaryCommentId,
					"error", updateErr,
				)
			}
			return fmt.Errorf("could not summarize terraform outputs: %v", err)
		}

		summaryMessage := "## AI summary for terraform outputs \n\n" + summary

		slog.Info("Updating PR comment with AI summary",
			"batchId", batch.ID,
			"prNumber", batch.PrNumber,
			"commentId", batch.AiSummaryCommentId,
			"summaryLength", len(summary),
		)

		updateErr := prService.EditComment(batch.PrNumber, batch.AiSummaryCommentId, summaryMessage)
		if updateErr != nil {
			slog.Error("Failed to update comment with AI summary",
				"commentId", batch.AiSummaryCommentId,
				"error", updateErr,
			)
			return fmt.Errorf("failed to update comment with AI summary: %v", updateErr)
		}

		slog.Info("Successfully updated PR comment with AI summary",
			"batchId", batch.ID,
			"prNumber", batch.PrNumber,
		)
	} else {
		if batch.Status != orchestrator_scheduler.BatchJobSucceeded {
			slog.Debug("Skipping AI summary - batch not successful",
				"batchId", batch.ID,
				"status", batch.Status,
			)
		} else if !config.Reporting.AiSummary {
			slog.Debug("Skipping AI summary - not enabled in config", "batchId", batch.ID)
		}
	}

	return nil
}

func AutomergePRforBatchIfEnabled(gh utils.GithubClientProvider, batch *models.DiggerBatch) error {
	slog.Info("Checking if PR should be auto-merged",
		"batchId", batch.ID,
		"prNumber", batch.PrNumber,
		"batchStatus", batch.Status,
		"batchType", batch.BatchType,
	)

	diggerYmlString := batch.DiggerConfig
	diggerConfigYml, err := digger_config.LoadDiggerConfigYamlFromString(diggerYmlString)
	if err != nil {
		slog.Error("Error loading Digger config from batch",
			"batchId", batch.ID,
			"error", err,
		)
		return fmt.Errorf("error loading digger config from batch: %v", err)
	}

	config, _, err := digger_config.ConvertDiggerYamlToConfig(diggerConfigYml)
	if err != nil {
		slog.Error("Error converting Digger YAML to config",
			"batchId", batch.ID,
			"error", err,
		)
		return fmt.Errorf("error loading digger config from yaml: %v", err)
	}

	automerge := config.AutoMerge
	automergeStrategy := config.AutoMergeStrategy

	slog.Debug("Auto-merge settings",
		"enabled", automerge,
		"strategy", automergeStrategy,
		"batchStatus", batch.Status,
		"batchType", batch.BatchType,
	)

	if batch.Status == orchestrator_scheduler.BatchJobSucceeded &&
		batch.BatchType == orchestrator_scheduler.DiggerCommandApply &&
		automerge == true {

		slog.Info("Conditions met for auto-merge, proceeding",
			"batchId", batch.ID,
			"prNumber", batch.PrNumber,
		)

		prService, err := GetPrServiceFromBatch(batch, gh)
		if err != nil {
			slog.Error("Error getting PR service",
				"batchId", batch.ID,
				"error", err,
			)
			return fmt.Errorf("error getting github service: %v", err)
		}

		slog.Info("Merging pull request",
			"prNumber", batch.PrNumber,
			"mergeStrategy", automergeStrategy,
		)

		err = prService.MergePullRequest(batch.PrNumber, string(automergeStrategy))
		if err != nil {
			slog.Error("Error merging pull request",
				"prNumber", batch.PrNumber,
				"mergeStrategy", automergeStrategy,
				"error", err,
			)
			return fmt.Errorf("error merging pull request: %v", err)
		}

		slog.Info("Successfully merged pull request",
			"prNumber", batch.PrNumber,
			"batchId", batch.ID,
		)
	} else {
		if batch.Status != orchestrator_scheduler.BatchJobSucceeded {
			slog.Debug("Skipping auto-merge - batch not successful",
				"batchId", batch.ID,
				"status", batch.Status,
			)
		} else if batch.BatchType != orchestrator_scheduler.DiggerCommandApply {
			slog.Debug("Skipping auto-merge - not an apply command",
				"batchId", batch.ID,
				"batchType", batch.BatchType,
			)
		} else if !automerge {
			slog.Debug("Skipping auto-merge - not enabled in config", "batchId", batch.ID)
		}
	}

	return nil
}
