package execution

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteTofuPlan(t *testing.T) {
	dir := CreateTestTerraformProject()
	defer func(name string) {
		err := os.RemoveAll(name)
		if err != nil {
			panic(err)
		}
	}(dir)

	CreateValidTerraformTestFile(dir)

	tf := OpenTofu{WorkingDir: dir, Workspace: "dev"}
	tf.Init([]string{}, map[string]string{})
	_, _, _, err := tf.Plan([]string{}, map[string]string{}, "")
	assert.NoError(t, err)
}

func TestExecuteTofuApply(t *testing.T) {
	dir := CreateTestTerraformProject()
	defer func(name string) {
		err := os.RemoveAll(name)
		if err != nil {
			panic(err)
		}
	}(dir)

	CreateValidTerraformTestFile(dir)

	tf := OpenTofu{WorkingDir: dir, Workspace: "dev"}
	tf.Init([]string{}, map[string]string{})
	_, _, _, err := tf.Plan([]string{}, map[string]string{}, "")
	assert.NoError(t, err)
}

func TestExecuteTofuApplyDefaultWorkspace(t *testing.T) {
	dir := CreateTestTerraformProject()
	defer func(name string) {
		err := os.RemoveAll(name)
		if err != nil {
			panic(err)
		}
	}(dir)

	CreateValidTerraformTestFile(dir)

	tf := OpenTofu{WorkingDir: dir, Workspace: "default"}
	tf.Init([]string{}, map[string]string{})
	var planArgs []string
	planArgs = append(planArgs, "-out", "plan.tfplan")
	tf.Plan(planArgs, map[string]string{}, "")
	plan := "plan.tfplan"
	_, _, err := tf.Apply([]string{}, &plan, map[string]string{})
	assert.NoError(t, err)
}
