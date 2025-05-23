package locking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/diggerhq/digger/libs/ci"
	"github.com/diggerhq/digger/libs/comment_utils/utils"
	"github.com/diggerhq/digger/libs/locking/aws"
	"github.com/diggerhq/digger/libs/locking/azure"
	"github.com/diggerhq/digger/libs/locking/gcp"

	"cloud.google.com/go/storage"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/diggerhq/digger/libs/comment_utils/reporting"
	"github.com/diggerhq/digger/libs/locking/aws/envprovider"
)

type PullRequestLock struct {
	InternalLock     Lock
	CIService        ci.PullRequestService
	Reporter         reporting.Reporter
	ProjectName      string
	ProjectNamespace string
	PrNumber         int
}

type NoOpLock struct {
}

func (noOpLock NoOpLock) Lock(transactionId int, resource string) (bool, error) {
	return true, nil
}

func (noOpLock NoOpLock) Unlock(resource string) (bool, error) {
	return true, nil
}

func (noOpLock NoOpLock) GetLock(resource string) (*int, error) {
	return nil, nil
}

func (projectLock *PullRequestLock) Lock() (bool, error) {
	lockId := projectLock.LockId()
	slog.Info("Attempting to acquire lock", "lockId", lockId)

	noHangingLocks, err := projectLock.verifyNoHangingLocks()
	if err != nil {
		return false, err
	}

	if !noHangingLocks {
		return false, nil
	}

	existingLockTransactionId, err := projectLock.InternalLock.GetLock(lockId)
	if err != nil {
		slog.Error("Failed to get lock", "lockId", lockId, "error", err)
		return false, err
	}

	if existingLockTransactionId != nil {
		if *existingLockTransactionId == projectLock.PrNumber {
			return true, nil
		} else {
			transactionIdStr := strconv.Itoa(*existingLockTransactionId)
			comment := "Project " + projectLock.projectId() + " locked by another PR #" + transactionIdStr + " (failed to acquire lock " + projectLock.ProjectNamespace + "). The locking plan must be applied or discarded before future plans can execute"

			reportLockingFailed(projectLock.Reporter, comment)
			return false, fmt.Errorf(comment)
		}
	}

	lockAcquired, err := projectLock.InternalLock.Lock(projectLock.PrNumber, lockId)
	if err != nil {
		return false, err
	}

	_, isNoOpLock := projectLock.InternalLock.(*NoOpLock)

	if lockAcquired && !isNoOpLock {
		comment := "Project " + projectLock.projectId() + " has been locked by PR #" + strconv.Itoa(projectLock.PrNumber)
		reportingLockingSuccess(projectLock.Reporter, comment)
		slog.Info("Project locked successfully",
			"projectId", projectLock.projectId(),
			"prNumber", projectLock.PrNumber)
	}
	return lockAcquired, nil
}

func reportingLockingSuccess(r reporting.Reporter, comment string) {
	if r.SupportsMarkdown() {
		_, _, err := r.Report(comment, utils.AsCollapsibleComment("Locking successful", false))
		if err != nil {
			slog.Error("Failed to publish comment", "error", err)
		}
	} else {
		_, _, err := r.Report(comment, utils.AsComment("Locking successful"))
		if err != nil {
			slog.Error("Failed to publish comment", "error", err)
		}
	}
}

func reportLockingFailed(r reporting.Reporter, comment string) {
	if r.SupportsMarkdown() {
		_, _, err := r.Report(comment, utils.AsCollapsibleComment("Locking failed", false))
		if err != nil {
			slog.Error("Failed to publish comment", "error", err)
		}
	} else {
		_, _, err := r.Report(comment, utils.AsComment("Locking failed"))
		if err != nil {
			slog.Error("Failed to publish comment", "error", err)
		}
	}
}

func (projectLock *PullRequestLock) verifyNoHangingLocks() (bool, error) {
	// TODO: Also include CI type (github, gitlab etc. into this lockID in order to avoid collision across VCS)
	lockId := projectLock.LockId()
	transactionId, err := projectLock.InternalLock.GetLock(lockId)

	if err != nil {
		return false, err
	}

	if transactionId != nil {
		if *transactionId != projectLock.PrNumber {
			isPrClosed, err := projectLock.CIService.IsClosed(*transactionId)
			if err != nil {
				return false, fmt.Errorf("failed to check if PR holding a lock is closed: %w", err)
			}
			if isPrClosed {
				slog.Info("Releasing lock held by closed PR",
					"lockId", lockId,
					"prNumber", *transactionId)
				_, err := projectLock.InternalLock.Unlock(lockId)
				if err != nil {
					return false, fmt.Errorf("failed to unlock a lock acquired by closed PR %v: %w", transactionId, err)
				}
				return true, nil
			}
			transactionIdStr := strconv.Itoa(*transactionId)
			comment := "Project " + projectLock.projectId() + " locked by another PR #" + transactionIdStr + "(failed to acquire lock " + projectLock.ProjectName + "). The locking plan must be applied or discarded before future plans can execute"
			reportLockingFailed(projectLock.Reporter, comment)
			return false, fmt.Errorf(comment)
		}
		return true, nil
	}
	return true, nil
}

func (projectLock *PullRequestLock) Unlock() (bool, error) {
	lockId := projectLock.LockId()
	slog.Info("Attempting to release lock", "lockId", lockId)

	lock, err := projectLock.InternalLock.GetLock(lockId)
	if err != nil {
		return false, err
	}

	if lock != nil {
		transactionId := *lock
		if projectLock.PrNumber == transactionId {
			lockReleased, err := projectLock.InternalLock.Unlock(lockId)
			if err != nil {
				return false, err
			}
			if lockReleased {
				comment := "Project unlocked (" + projectLock.projectId() + ")."
				reportSuccessfulUnlocking(projectLock.Reporter, comment)

				slog.Info("Project unlocked successfully", "projectId", projectLock.projectId())
				return true, nil
			}
		}
	}
	return false, nil
}

func reportSuccessfulUnlocking(r reporting.Reporter, comment string) {
	if r.SupportsMarkdown() {
		_, _, err := r.Report(comment, utils.AsCollapsibleComment("Unlocking successful", false))
		if err != nil {
			slog.Error("Failed to publish comment", "error", err)
		}
	} else {
		_, _, err := r.Report(comment, utils.AsComment("Unlocking successful"))
		if err != nil {
			slog.Error("Failed to publish comment", "error", err)
		}
	}
}

func (projectLock *PullRequestLock) ForceUnlock() error {
	lockId := projectLock.LockId()
	slog.Info("Force unlocking", "lockId", lockId)

	lock, err := projectLock.InternalLock.GetLock(lockId)
	if err != nil {
		return err
	}
	if lock != nil {
		lockReleased, err := projectLock.InternalLock.Unlock(lockId)
		if err != nil {
			return err
		}

		if lockReleased {
			comment := "Project unlocked (" + projectLock.projectId() + ")."
			reportSuccessfulUnlocking(projectLock.Reporter, comment)
			slog.Info("Project force unlocked successfully", "projectId", projectLock.projectId())
		}
		return nil
	}
	return nil
}

func (projectLock *PullRequestLock) projectId() string {
	return projectLock.ProjectNamespace + "#" + projectLock.ProjectName
}

func (projectLock *PullRequestLock) LockId() string {
	return projectLock.ProjectNamespace + "#" + projectLock.ProjectName
}

// DoEnvVarsExist return true if any of env vars do exist, false otherwise
func DoEnvVarsExist(envVars []string) bool {
	result := false
	for _, key := range envVars {
		value := os.Getenv(key)
		if value != "" {
			result = true
		}
	}
	return result
}

func GetLock() (Lock, error) {
	awsRegion := strings.ToLower(os.Getenv("AWS_REGION"))
	awsProfile := strings.ToLower(os.Getenv("AWS_PROFILE"))
	lockProvider := strings.ToLower(os.Getenv("LOCK_PROVIDER"))
	disableLocking := strings.ToLower(os.Getenv("DISABLE_LOCKING")) == "true"

	if disableLocking {
		slog.Info("Using NoOp lock provider")
		return &NoOpLock{}, nil
	}

	if lockProvider == "" || lockProvider == "aws" {
		slog.Info("Using AWS lock provider")

		// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/
		// https://aws.github.io/aws-sdk-go-v2/docs/migrating/
		keysToCheck := []string{"DIGGER_AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY"}
		awsCredsProvided := DoEnvVarsExist(keysToCheck)

		var cfg awssdk.Config
		var err error
		if awsCredsProvided {
			cfg, err = config.LoadDefaultConfig(context.Background(),
				config.WithSharedConfigProfile(awsProfile),
				config.WithRegion(awsRegion),
				config.WithCredentialsProvider(&envprovider.EnvProvider{}))
			if err != nil {
				return nil, err
			}
		} else {
			slog.Info("Using keyless AWS configuration")
			cfg, err = config.LoadDefaultConfig(context.Background(), config.WithRegion(awsRegion))
			if err != nil {
				return nil, err
			}
		}

		stsClient := sts.NewFromConfig(cfg)
		result, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to AWS account. %v", err)
		}
		slog.Info("Successfully connected to AWS account",
			"accountId", *result.Account,
			"userId", *result.UserId)

		dynamoDb := dynamodb.NewFromConfig(cfg)
		dynamoDbLock := aws.DynamoDbLock{DynamoDb: dynamoDb}
		return &dynamoDbLock, nil
	} else if lockProvider == "gcp" {
		slog.Info("Using GCP lock provider")
		ctx, client := gcp.GetGoogleStorageClient()
		defer func(client *storage.Client) {
			err := client.Close()
			if err != nil {
				slog.Error("Failed to close Google Storage client", "error", err)
			}
		}(client)

		bucketName := strings.ToLower(os.Getenv("GOOGLE_STORAGE_LOCK_BUCKET"))
		if bucketName == "" {
			return nil, errors.New("GOOGLE_STORAGE_LOCK_BUCKET is not set")
		}
		bucket := client.Bucket(bucketName)
		lock := gcp.GoogleStorageLock{Client: client, Bucket: bucket, Context: ctx}
		return &lock, nil
	} else if lockProvider == "azure" {
		slog.Info("Using Azure lock provider")
		return azure.NewStorageAccountLock()
	}

	return nil, errors.New("failed to find lock provider")
}
