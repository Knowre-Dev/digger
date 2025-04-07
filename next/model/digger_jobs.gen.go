// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameDiggerJob = "digger_jobs"

// DiggerJob mapped from table <digger_jobs>
type DiggerJob struct {
	ID                 string         `gorm:"column:id;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt          time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
	DiggerJobID        string         `gorm:"column:digger_job_id;not null" json:"digger_job_id"`
	Status             int16          `gorm:"column:status;not null" json:"status"`
	BatchID            string         `gorm:"column:batch_id;not null" json:"batch_id"`
	StatusUpdatedAt    time.Time      `gorm:"column:status_updated_at" json:"status_updated_at"`
	DiggerJobSummaryID string         `gorm:"column:digger_job_summary_id" json:"digger_job_summary_id"`
	WorkflowFile       string         `gorm:"column:workflow_file" json:"workflow_file"`
	WorkflowRunURL     string         `gorm:"column:workflow_run_url" json:"workflow_run_url"`
	PlanFootprint      []uint8        `gorm:"column:plan_footprint" json:"plan_footprint"`
	PrCommentURL       string         `gorm:"column:pr_comment_url" json:"pr_comment_url"`
	TerraformOutput    string         `gorm:"column:terraform_output" json:"terraform_output"`
	JobSpec            []uint8        `gorm:"column:job_spec" json:"job_spec"`
	VariablesSpec      []uint8        `gorm:"column:variables_spec" json:"variables_spec"`
}

// TableName DiggerJob's table name
func (*DiggerJob) TableName() string {
	return TableNameDiggerJob
}
