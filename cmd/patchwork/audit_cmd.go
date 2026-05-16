package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/patchwork/internal/audit"
)

func runAudit(args []string) {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	auditFile := fs.String("audit", "patchwork-audit.json", "path to audit log file")
	env := fs.String("env", "", "filter by environment")
	kind := fs.String("kind", "", "filter by event kind (capture, diff, baseline, rollback, prune, export)")
	_ = fs.Parse(args)

	log, err := audit.Load(*auditFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading audit log: %v\n", err)
		os.Exit(1)
	}

	events := audit.Filter(log, audit.EventKind(*kind), *env)
	fmt.Print(audit.Render(events))
}

func recordAudit(auditFile string, kind audit.EventKind, env, message string) {
	log, err := audit.Load(auditFile)
	if err != nil {
		// Non-fatal: audit logging should not block primary operations.
		fmt.Fprintf(os.Stderr, "audit: load warning: %v\n", err)
		log = &audit.Log{}
	}
	audit.Append(log, kind, env, message)
	if err := audit.Save(auditFile, log); err != nil {
		fmt.Fprintf(os.Stderr, "audit: save warning: %v\n", err)
	}
}
