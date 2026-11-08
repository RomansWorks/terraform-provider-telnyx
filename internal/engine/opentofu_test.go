// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var (
	tofu111Once sync.Once
	tofu111     bool
)

// openTofu111 reports whether the tests run with OpenTofu 1.11 (the OpenTofu
// CI job sets TF_ACC_TERRAFORM_PATH to tofu).
func openTofu111() bool {
	tofu111Once.Do(func() {
		path := os.Getenv("TF_ACC_TERRAFORM_PATH")
		if !strings.HasPrefix(filepath.Base(path), "tofu") {
			return
		}
		out, err := exec.Command(path, "version", "-json").Output()
		if err != nil {
			return
		}
		var v struct {
			Version string `json:"terraform_version"`
		}
		if json.Unmarshal(out, &v) == nil {
			tofu111 = strings.HasPrefix(v.Version, "1.11.")
		}
	})
	return tofu111
}

// ephemeralStep adapts a step that uses ephemeral resources to OpenTofu 1.11.
// It stores ephemeral resources in the plan (1.12 stopped doing so) and lists
// them in the JSON plan with a non-no-op action, even when it prints "No
// changes", so the test framework's empty-plan check after apply fails there.
// Under OpenTofu 1.11 the step expects a non-empty plan instead, and
// onlyEphemeralChanges keeps the real check: nothing but ephemeral resources
// may change.
func ephemeralStep(s rtest.TestStep) rtest.TestStep {
	if !openTofu111() {
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
