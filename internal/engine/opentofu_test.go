// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// openTofu reports whether the tests run with the OpenTofu CLI (the
// OpenTofu CI job sets TF_ACC_TERRAFORM_PATH to tofu).
func openTofu() bool {
	return strings.HasPrefix(filepath.Base(os.Getenv("TF_ACC_TERRAFORM_PATH")), "tofu")
}

// ephemeralStep adapts a step that uses ephemeral resources to OpenTofu.
// OpenTofu lists every ephemeral resource in the JSON plan with a non-no-op
// action, even when it prints "No changes", so the test framework's
// empty-plan check after apply always fails there. Under OpenTofu the step
// expects a non-empty plan instead, and onlyEphemeralChanges keeps the real
// check: nothing but ephemeral resources may change.
func ephemeralStep(s rtest.TestStep) rtest.TestStep {
	if !openTofu() {
		return s
	}
	s.ExpectNonEmptyPlan = true
	s.ConfigPlanChecks.PostApplyPostRefresh = append(s.ConfigPlanChecks.PostApplyPostRefresh, onlyEphemeralChanges{})
	return s
}

type onlyEphemeralChanges struct{}

func (onlyEphemeralChanges) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	for _, rc := range req.Plan.ResourceChanges {
		if string(rc.Mode) == "ephemeral" || rc.Change == nil {
			continue
		}
		if !rc.Change.Actions.NoOp() {
			resp.Error = fmt.Errorf("after apply, %s plans %v", rc.Address, rc.Change.Actions)
			return
		}
	}
}
