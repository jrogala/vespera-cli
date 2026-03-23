package tests

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/jrogala/vespera-cli/client"
	"github.com/jrogala/vespera-cli/pkg/ops"
)

func initializeScenario(ctx *godog.ScenarioContext) {
	sc := newScenarioCtx()

	ctx.Before(func(ctx context.Context, sc2 *godog.Scenario) (context.Context, error) {
		sc.lastErr = nil
		sc.statusResult = nil
		sc.observationsList = nil
		sc.filesList = nil
		sc.treeResult = nil
		sc.mockConn = newMockFTPConn()
		sc.client = client.NewFTPClientWithConn("10.0.0.1", 21, sc.mockConn)
		return ctx, nil
	})

	// --- Background ---
	ctx.Step(`^a connected telescope$`, sc.aConnectedTelescope)

	// --- Status ---
	ctx.Step(`^the telescope has (\d+) observations$`, sc.theTelescopeHasNObservations)
	ctx.Step(`^I request the telescope status$`, sc.iRequestTheTelescopeStatus)
	ctx.Step(`^I should see the host address$`, sc.iShouldSeeTheHostAddress)
	ctx.Step(`^I should see (\d+) observations$`, sc.iShouldSeeNObservations)

	// --- List observations ---
	ctx.Step(`^the telescope has observations:$`, sc.theTelescopeHasObservations)
	ctx.Step(`^I list observations$`, sc.iListObservations)
	ctx.Step(`^I should get (\d+) observations$`, sc.iShouldGetNObservations)
	ctx.Step(`^observation "([^"]*)" should be in the list$`, sc.observationShouldBeInTheList)

	// --- Files ---
	ctx.Step(`^observation "([^"]*)" has files:$`, sc.observationHasFiles)
	ctx.Step(`^observation "([^"]*)" has no files$`, sc.observationHasNoFiles)
	ctx.Step(`^I list files for "([^"]*)"$`, sc.iListFilesFor)
	ctx.Step(`^I should get (\d+) files$`, sc.iShouldGetNFiles)
	ctx.Step(`^file "([^"]*)" should have type "([^"]*)"$`, sc.fileShouldHaveType)

	// --- Tree ---
	ctx.Step(`^the telescope has a directory tree:$`, sc.theTelescopeHasDirectoryTree)
	ctx.Step(`^I request the tree for "([^"]*)"$`, sc.iRequestTheTreeFor)
	ctx.Step(`^the tree should contain directory "([^"]*)"$`, sc.theTreeShouldContainDirectory)
	ctx.Step(`^the tree should contain file "([^"]*)"$`, sc.theTreeShouldContainFile)
}

// --- Background ---

func (sc *scenarioCtx) aConnectedTelescope() error {
	if sc.client == nil {
		return fmt.Errorf("no client available")
	}
	return nil
}

// --- Status steps ---

func (sc *scenarioCtx) theTelescopeHasNObservations(n int) error {
	sc.mockConn.ensureDir("/user")
	sc.mockConn.ensureDir("/system/captures")
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("obs-%d", i+1)
		sc.mockConn.addDir("/user", name, time.Date(2024, 1, 15+i, 21, 30, 0, 0, time.UTC))
	}
	return nil
}

func (sc *scenarioCtx) iRequestTheTelescopeStatus() error {
	result, err := ops.GetStatus(sc.client)
	sc.statusResult = result
	sc.lastErr = err
	return nil
}

func (sc *scenarioCtx) iShouldSeeTheHostAddress() error {
	if sc.lastErr != nil {
		return sc.lastErr
	}
	result := sc.statusResult.(*ops.StatusResult)
	if result.Host == "" {
		return fmt.Errorf("expected non-empty host")
	}
	return nil
}

func (sc *scenarioCtx) iShouldSeeNObservations(n int) error {
	if sc.lastErr != nil {
		return sc.lastErr
	}
	result := sc.statusResult.(*ops.StatusResult)
	if result.ObservationCount != n {
		return fmt.Errorf("expected %d observations, got %d", n, result.ObservationCount)
	}
	return nil
}

// --- List observations steps ---

func (sc *scenarioCtx) theTelescopeHasObservations(table *godog.Table) error {
	sc.mockConn.ensureDir("/user")
	sc.mockConn.ensureDir("/system/captures")
	for i, row := range table.Rows {
		if i == 0 {
			continue // skip header
		}
		name := row.Cells[0].Value
		dateStr := row.Cells[1].Value
		t, err := time.Parse("2006-01-02 15:04", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date %q: %w", dateStr, err)
		}
		sc.mockConn.addDir("/user", name, t)
	}
	return nil
}

func (sc *scenarioCtx) iListObservations() error {
	result, err := ops.ListObservations(sc.client)
	sc.observationsList = result
	sc.lastErr = err
	return nil
}

func (sc *scenarioCtx) iShouldGetNObservations(n int) error {
	if sc.lastErr != nil {
		return sc.lastErr
	}
	list := sc.observationsList.([]client.ObservationEntry)
	if len(list) != n {
		return fmt.Errorf("expected %d observations, got %d", n, len(list))
	}
	return nil
}

func (sc *scenarioCtx) observationShouldBeInTheList(name string) error {
	if sc.lastErr != nil {
		return sc.lastErr
	}
	list := sc.observationsList.([]client.ObservationEntry)
	for _, o := range list {
		if o.Name == name {
			return nil
		}
	}
	return fmt.Errorf("observation %q not found in list", name)
}

// --- Files steps ---

func (sc *scenarioCtx) observationHasFiles(observation string, table *godog.Table) error {
	sc.mockConn.ensureDir("/user")
	sc.mockConn.ensureDir("/system/captures")
	// Add the observation directory to /user
	sc.mockConn.addDir("/user", observation, time.Now())
	obsPath := "/user/" + observation
	sc.mockConn.ensureDir(obsPath)

	for i, row := range table.Rows {
		if i == 0 {
			continue // skip header
		}
		fullPath := row.Cells[0].Value
		sizeStr := row.Cells[1].Value
		size, _ := strconv.ParseUint(sizeStr, 10, 64)

		// Determine directory and filename from full path
		dir := fullPath[:strings.LastIndex(fullPath, "/")]
		base := fullPath[strings.LastIndex(fullPath, "/")+1:]

		sc.mockConn.ensureDir(dir)
		sc.mockConn.addFile(dir, base, size, "mock-content")
	}
	return nil
}

func (sc *scenarioCtx) observationHasNoFiles(observation string) error {
	sc.mockConn.ensureDir("/user")
	sc.mockConn.ensureDir("/system/captures")
	sc.mockConn.addDir("/user", observation, time.Now())
	sc.mockConn.ensureDir("/user/" + observation)
	// Also register in system/captures so fallback doesn't error
	sc.mockConn.addDir("/system/captures", observation, time.Now())
	sc.mockConn.ensureDir("/system/captures/" + observation)
	return nil
}

func (sc *scenarioCtx) iListFilesFor(observation string) error {
	result, err := ops.ListFiles(sc.client, observation)
	sc.filesList = result
	sc.lastErr = err
	return nil
}

func (sc *scenarioCtx) iShouldGetNFiles(n int) error {
	if sc.lastErr != nil {
		return sc.lastErr
	}
	list := sc.filesList.([]client.FileEntry)
	if len(list) != n {
		return fmt.Errorf("expected %d files, got %d", n, len(list))
	}
	return nil
}

func (sc *scenarioCtx) fileShouldHaveType(name, expectedType string) error {
	if sc.lastErr != nil {
		return sc.lastErr
	}
	list := sc.filesList.([]client.FileEntry)
	for _, f := range list {
		base := filepath.Base(f.Name)
		if base == name {
			if f.Type != expectedType {
				return fmt.Errorf("file %q has type %q, expected %q", name, f.Type, expectedType)
			}
			return nil
		}
	}
	return fmt.Errorf("file %q not found", name)
}

// --- Tree steps ---

func (sc *scenarioCtx) theTelescopeHasDirectoryTree(table *godog.Table) error {
	for i, row := range table.Rows {
		if i == 0 {
			continue // skip header
		}
		path := row.Cells[0].Value
		entryType := row.Cells[1].Value
		sizeStr := row.Cells[2].Value
		size, _ := strconv.ParseUint(sizeStr, 10, 64)

		dir := path[:strings.LastIndex(path, "/")]
		if dir == "" {
			dir = "/"
		}
		base := path[strings.LastIndex(path, "/")+1:]

		sc.mockConn.ensureDir(dir)

		if entryType == "dir" {
			sc.mockConn.addDir(dir, base, time.Now())
			sc.mockConn.ensureDir(path)
		} else {
			sc.mockConn.addFile(dir, base, size, "mock-content")
		}
	}
	return nil
}

func (sc *scenarioCtx) iRequestTheTreeFor(path string) error {
	result, err := ops.GetTree(sc.client, path)
	sc.treeResult = result
	sc.lastErr = err
	return nil
}

func (sc *scenarioCtx) theTreeShouldContainDirectory(name string) error {
	if sc.lastErr != nil {
		return sc.lastErr
	}
	tree := sc.treeResult.([]ops.TreeEntry)
	if findInTree(tree, name, true) {
		return nil
	}
	return fmt.Errorf("directory %q not found in tree", name)
}

func (sc *scenarioCtx) theTreeShouldContainFile(name string) error {
	if sc.lastErr != nil {
		return sc.lastErr
	}
	tree := sc.treeResult.([]ops.TreeEntry)
	if findInTree(tree, name, false) {
		return nil
	}
	return fmt.Errorf("file %q not found in tree", name)
}

func findInTree(entries []ops.TreeEntry, name string, isDir bool) bool {
	for _, e := range entries {
		if e.Name == name && e.IsDir == isDir {
			return true
		}
		if e.IsDir && len(e.Children) > 0 {
			if findInTree(e.Children, name, isDir) {
				return true
			}
		}
	}
	return false
}
