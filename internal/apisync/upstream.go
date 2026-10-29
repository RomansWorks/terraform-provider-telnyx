// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package apisync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Lock pins the upstream spec to one commit.
type Lock struct {
	Repository string `json:"repository"`
	Ref        string `json:"ref"`
	Commit     string `json:"commit"`
	File       string `json:"file"`
	SHA256     string `json:"sha256"`
	FetchedAt  string `json:"fetched_at"`
}

// DefaultLock describes where the Telnyx spec comes from.
var DefaultLock = Lock{
	Repository: "https://github.com/team-telnyx/openapi",
	Ref:        "master",
	File:       "openapi/spec3.json",
}

// LoadLock reads a lock file.
func LoadLock(path string) (Lock, error) {
	var l Lock
	data, err := os.ReadFile(path)
	if err != nil {
		return l, err
	}
	err = json.Unmarshal(data, &l)
	return l, err
}

// Save writes the lock file.
func (l Lock) Save(path string) error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// FileSHA256 returns the hex sha256 of a file.
func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyLock checks that the pinned spec file matches its lock.
func VerifyLock(specPath, lockPath string) error {
	l, err := LoadLock(lockPath)
	if err != nil {
		return fmt.Errorf("reading lock: %w", err)
	}
	sum, err := FileSHA256(specPath)
	if err != nil {
		return err
	}
	if sum != l.SHA256 {
		return fmt.Errorf("%s does not match %s (sha256 %s, lock says %s); do not edit the pinned spec by hand, run make upstream-fetch", specPath, lockPath, sum, l.SHA256)
	}
	return nil
}

// ResolveCommit returns the commit a ref points to, using git ls-remote.
func ResolveCommit(ctx context.Context, repo, ref string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "ls-remote", repo, ref).Output()
	if err != nil {
		return "", fmt.Errorf("git ls-remote %s %s: %w", repo, ref, err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && (fields[1] == ref || fields[1] == "refs/heads/"+ref || fields[1] == "refs/tags/"+ref) {
			return fields[0], nil
		}
	}
	if len(ref) == 40 {
		return ref, nil // already a commit
	}
	return "", fmt.Errorf("ref %s not found in %s", ref, repo)
}

// RawURL returns the raw.githubusercontent.com URL of a file at a commit.
func RawURL(repo, commit, file string) (string, error) {
	const gh = "https://github.com/"
	if !strings.HasPrefix(repo, gh) {
		return "", fmt.Errorf("only github.com repositories are supported, got %s", repo)
	}
	return "https://raw.githubusercontent.com/" + strings.TrimSuffix(strings.TrimPrefix(repo, gh), ".git") + "/" + commit + "/" + file, nil
}

// Fetch downloads the spec at the commit the lock's ref points to. It
// returns the new lock and whether the commit changed.
func Fetch(ctx context.Context, current Lock, specPath string) (Lock, bool, error) {
	next := current
	if next.Repository == "" {
		next = DefaultLock
	}
	commit, err := ResolveCommit(ctx, next.Repository, next.Ref)
	if err != nil {
		return current, false, err
	}
	if commit == current.Commit {
		return current, false, nil
	}
	u, err := RawURL(next.Repository, commit, next.File)
	if err != nil {
		return current, false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return current, false, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return current, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return current, false, fmt.Errorf("GET %s: %s", u, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return current, false, err
	}
	var probe map[string]any
	if err := json.Unmarshal(data, &probe); err != nil {
		return current, false, fmt.Errorf("downloaded spec is not valid JSON: %w", err)
	}
	if err := os.WriteFile(specPath, data, 0o644); err != nil {
		return current, false, err
	}
	sum := sha256.Sum256(data)
	next.Commit = commit
	next.SHA256 = hex.EncodeToString(sum[:])
	next.FetchedAt = time.Now().UTC().Format(time.RFC3339)
	return next, true, nil
}
