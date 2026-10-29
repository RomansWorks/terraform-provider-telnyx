// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

// Command apisync keeps the pinned Telnyx OpenAPI spec and its derived views
// up to date. Run it through the Makefile targets rather than directly.
//
//	apisync fetch    download the latest upstream spec if the commit changed
//	apisync catalog  rebuild the per-tag split, resource catalog and coverage
//	apisync check    verify the lock and that every operation is classified
//	apisync diff     write a Markdown report comparing two spec files
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/apisync"
)

const (
	specPath     = "upstream/telnyx/spec3.json"
	lockPath     = "upstream/telnyx/lock.json"
	splitDir     = "upstream/telnyx/by-tag"
	catalogPath  = "catalog/resources.yaml"
	coveragePath = "catalog/coverage.yaml"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "fetch":
		err = fetch(os.Args[2:])
	case "catalog":
		err = catalog()
	case "check":
		err = check(os.Args[2:])
	case "diff":
		err = diff(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "apisync:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: apisync fetch [-ref REF] | catalog | check [-require-reviewed] | diff -old FILE [-old-lock FILE] [-new FILE] [-out FILE]")
	os.Exit(2)
}

func fetch(args []string) error {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	ref := fs.String("ref", "", "upstream branch, tag or commit (default: the lock's ref)")
	_ = fs.Parse(args)

	lock, err := apisync.LoadLock(lockPath)
	if os.IsNotExist(err) {
		lock = apisync.DefaultLock
	} else if err != nil {
		return err
	}
	if *ref != "" {
		lock.Ref = *ref
	}
	if err := os.MkdirAll(filepath.Dir(specPath), 0o755); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	next, changed, err := apisync.Fetch(ctx, lock, specPath)
	if err != nil {
		return err
	}
	if changed {
		if err := next.Save(lockPath); err != nil {
			return err
		}
		fmt.Printf("updated %s from %s to %s\n", specPath, short(lock.Commit), short(next.Commit))
	} else {
		fmt.Printf("%s is up to date at %s\n", specPath, short(next.Commit))
	}
	return githubOutput(map[string]string{
		"changed":    fmt.Sprint(changed),
		"old_commit": lock.Commit,
		"new_commit": next.Commit,
	})
}

func catalog() error {
	s, err := apisync.LoadSpec(specPath)
	if err != nil {
		return err
	}
	lock, err := apisync.LoadLock(lockPath)
	if err != nil {
		return err
	}
	if err := apisync.WriteSplit(s, splitDir); err != nil {
		return err
	}
	resources := apisync.DetectResources(s)
	if err := os.MkdirAll(filepath.Dir(catalogPath), 0o755); err != nil {
		return err
	}
	if err := (apisync.Catalog{UpstreamCommit: lock.Commit, Resources: resources}).Save(catalogPath); err != nil {
		return err
	}
	cov, err := apisync.LoadCoverage(coveragePath)
	if err != nil {
		return err
	}
	res := cov.Sync(s, resources)
	if err := cov.Save(coveragePath); err != nil {
		return err
	}
	fmt.Printf("catalog: %d operations, %d candidate resources; coverage +%d -%d\n",
		len(s.Operations()), len(resources), len(res.Added), len(res.Removed))
	return nil
}

func check(args []string) error {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	requireReviewed := fs.Bool("require-reviewed", false, "fail while any coverage entry is unreviewed")
	_ = fs.Parse(args)

	if err := apisync.VerifyLock(specPath, lockPath); err != nil {
		return err
	}
	s, err := apisync.LoadSpec(specPath)
	if err != nil {
		return err
	}
	cov, err := apisync.LoadCoverage(coveragePath)
	if err != nil {
		return err
	}
	rep := cov.Check(s)
	fmt.Print(rep.String())
	if !rep.OK(*requireReviewed) {
		return fmt.Errorf("coverage check failed")
	}
	return nil
}

func diff(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	oldPath := fs.String("old", "", "previous spec file")
	newPath := fs.String("new", specPath, "new spec file")
	oldCommit := fs.String("old-commit", "", "previous upstream commit")
	oldLock := fs.String("old-lock", "", "previous lock file; sets -old-commit")
	out := fs.String("out", "", "write Markdown here instead of stdout")
	_ = fs.Parse(args)
	if *oldPath == "" {
		return fmt.Errorf("-old is required")
	}
	oldSpec, err := apisync.LoadSpec(*oldPath)
	if err != nil {
		return err
	}
	newSpec, err := apisync.LoadSpec(*newPath)
	if err != nil {
		return err
	}
	d := apisync.Compare(oldSpec, newSpec)
	d.OldCommit = *oldCommit
	if *oldLock != "" {
		l, err := apisync.LoadLock(*oldLock)
		if err != nil {
			return err
		}
		d.OldCommit = l.Commit
	}
	if l, err := apisync.LoadLock(lockPath); err == nil {
		d.NewCommit = l.Commit
	}
	md := d.Markdown()
	if *out == "" {
		fmt.Print(md)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(*out, []byte(md), 0o644)
}

func githubOutput(kv map[string]string) error {
	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	for k, v := range kv {
		if _, err := fmt.Fprintf(f, "%s=%s\n", k, v); err != nil {
			return err
		}
	}
	return nil
}

func short(c string) string {
	if len(c) > 7 {
		return c[:7]
	}
	if c == "" {
		return "(none)"
	}
	return c
}
