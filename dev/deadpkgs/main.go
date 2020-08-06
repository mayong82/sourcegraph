package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

var concurrency = runtime.GOMAXPROCS(0)

const root = "github.com/sourcegraph/sourcegraph"

func main() {
	if err := mainErr(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	cmd := exec.Command("go", "list", fmt.Sprintf("%s/...", root))
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	var pkgs []string
	for _, pkg := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		pkgs = append(pkgs, strings.TrimPrefix(strings.TrimPrefix(pkg, root), "/"))
	}

	imports, err := getAllImports(pkgs)
	if err != nil {
		return err
	}

	imported := map[string]struct{}{}
	for _, vs := range imports {
		for _, v := range vs {
			imported[v] = struct{}{}
		}
	}

	for k := range imports {
		if _, ok := imported[k]; !ok {
			fmt.Printf("Dead: %v\n", k)
		}
	}

	return nil
}

func getAllImports(pkgs []string) (map[string][]string, error) {
	ch := make(chan string, len(pkgs))
	for _, pkg := range pkgs {
		ch <- pkg
	}
	close(ch)

	type pair struct {
		pkg     string
		imports []string
		err     error
	}

	var wg sync.WaitGroup
	pairs := make(chan pair, len(pkgs))

	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for pkg := range ch {
				imports, err := getImports(pkg)
				pairs <- pair{pkg, imports, err}
			}
		}()
	}
	wg.Wait()
	close(pairs)

	allImports := map[string][]string{}
	for pair := range pairs {
		if err := pair.err; err != nil {
			return nil, err
		}

		allImports[pair.pkg] = pair.imports
	}

	return allImports, nil
}

func getImports(pkg string) ([]string, error) {
	cmd := exec.Command("go", "list", "-f", `{{ join .Imports "\n" }}`, fmt.Sprintf("%s/%s", root, pkg))
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var importPackages []string
	for _, importPkg := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.HasPrefix(importPkg, root) {
			importPackages = append(importPackages, strings.TrimPrefix(importPkg, root+"/"))
		}
	}

	return importPackages, nil
}
