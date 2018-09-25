//docker:user sourcegraph

// Package frontend contains the enterprise frontend implementation.
//
// It wraps the open source frontend command and merely injects a few
// proprietary things into it via e.g. blank/underscore imports in this file
// which register side effects with the frontend package.
package main

import (
	"log"
	"os"
	"strconv"

	_ "github.com/sourcegraph/enterprise/cmd/frontend/auth"
	_ "github.com/sourcegraph/enterprise/cmd/frontend/internal/httpapi"
	_ "github.com/sourcegraph/enterprise/cmd/frontend/internal/licensing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
)

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}
	shared.Main()
}
