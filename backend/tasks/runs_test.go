package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/diggerhq/digger/backend/models"
	"github.com/diggerhq/digger/backend/utils"
	github2 "github.com/diggerhq/digger/libs/ci/github"
	orchestrator_scheduler "github.com/diggerhq/digger/libs/scheduler"
	"github.com/diggerhq/digger/libs/spec"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockCiBackend struct {
}

func (m MockCiBackend) TriggerWorkflow(spec spec.Spec, runName string, vcsToken string) error {
	return nil
}

func (m MockCiBackend) GetWorkflowUrl(spec spec.Spec) (string, error) {
	return "", nil
}

func setupSuite(tb testing.TB) (func(tb testing.TB), *models.Database) {
	// database file name
	dbName := "database_tasks_test.db"

	// remove old database
	e := os.Remove(dbName)
	if e != nil {
		if !strings.Contains(e.Error(), "no such file or directory") {
			panic(e)
		}
	}

	// open and create a new database
	gdb, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// migrate tables
	err = gdb.AutoMigrate(&models.Policy{}, &models.Organisation{}, &models.Repo{}, &models.Project{}, &models.Token{},
		&models.User{}, &models.ProjectRun{}, &models.GithubAppInstallation{}, &models.VCSConnection{}, &models.GithubAppInstallationLink{},
		&models.GithubDiggerJobLink{}, &models.DiggerJob{}, &models.DiggerJobParentLink{}, &models.DiggerRun{}, &models.DiggerRunQueueItem{})
	if err != nil {
		panic(err)
	}

	database := &models.Database{GormDB: gdb}
	models.DB = database

	// create an org
	orgTenantId := "11111111-1111-1111-1111-111111111111"
	externalSource := "test"
	orgName := "testOrg"
	org, err := database.CreateOrganisation(orgName, externalSource, orgTenantId)
	if err != nil {
		panic(err)
	}

	// create digger repo
	repoName := "test repo"
	repo, err := database.CreateRepo(repoName, "", "", "", "", org, "")
	if err != nil {
		panic(err)
	}

	// create test project
	projectName := "test project"
	_, err = database.CreateProject(projectName, org, repo, false, false)
	if err != nil {
		panic(err)
	}

	models.DB = database
	// Return a function to teardown the test
	return func(tb testing.TB) {
		err = os.Remove(dbName)
		if err != nil {
			panic(err)
		}
	}, database
}

func TestThatRunQueueItemMovesFromQueuedToPlanningAfterPickup(t *testing.T) {
	// TODO: Fix the failing tests and unskip
	t.Skip()

	teardownSuite, _ := setupSuite(t)
	defer teardownSuite(t)

	type params struct {
		BatchStatus        orchestrator_scheduler.DiggerBatchStatus
		InitialStatus      models.DiggerRunStatus
		NextExpectedStatus models.DiggerRunStatus
	}

	testParameters := []params{
		{
			BatchStatus:        orchestrator_scheduler.BatchJobSucceeded,
			InitialStatus:      models.RunQueued,
			NextExpectedStatus: models.RunPlanning,
		},
		// TODO: Uncomment when functionality working
		//{
		//	BatchStatus:        orchestrator_scheduler.BatchJobFailed,
		//	InitialStatus:      models.RunPlanning,
		//	NextExpectedStatus: models.RunFailed,
		//},
		{
			BatchStatus:        orchestrator_scheduler.BatchJobSucceeded,
			InitialStatus:      models.RunPlanning,
			NextExpectedStatus: models.RunPendingApproval,
		},
		{
			BatchStatus:        orchestrator_scheduler.BatchJobSucceeded,
			InitialStatus:      models.RunApproved,
			NextExpectedStatus: models.RunApplying,
		},
		// TODO: uncomment when functional
		//{
		//	BatchStatus:        orchestrator_scheduler.BatchJobFailed,
		//	InitialStatus:      models.RunApplying,
		//	NextExpectedStatus: models.RunFailed,
		//},
		{
			BatchStatus:        orchestrator_scheduler.BatchJobSucceeded,
			InitialStatus:      models.RunApplying,
			NextExpectedStatus: models.RunSucceeded,
		},
	}

	for i, testParam := range testParameters {
		ciService := github2.MockCiService{}
		batch, _ := models.DB.CreateDiggerBatch(models.DiggerVCSGithub, 123, "", "", "", 22, "", "", "", nil, 0, "", false, nil)
		project, _ := models.DB.CreateProject(fmt.Sprintf("test%v", i), nil, nil, false, false)
		planStage, _ := models.DB.CreateDiggerRunStage(batch.ID.String())
		applyStage, _ := models.DB.CreateDiggerRunStage(batch.ID.String())
		diggerRun, _ := models.DB.CreateDiggerRun("", 1, testParam.InitialStatus, "sha", "", 123, 1, project.Name, "apply", &planStage.ID, &applyStage.ID)
		queueItem, _ := models.DB.CreateDiggerRunQueueItem(project.ID, diggerRun.ID)
		batch.Status = testParam.BatchStatus
		models.DB.UpdateDiggerBatch(batch)
		queueItem, _ = models.DB.GetDiggerRunQueueItem(queueItem.ID)

		RunQueuesStateMachine(queueItem, ciService, utils.DiggerGithubClientMockProvider{})
		diggerRunRefreshed, _ := models.DB.GetDiggerRun(diggerRun.ID)
		assert.Equal(t, testParam.NextExpectedStatus, diggerRunRefreshed.Status)
	}

}
