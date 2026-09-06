package service

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"ai-chat-backend/internal/config"
	"ai-chat-backend/internal/store"
)

func promptTestService(t *testing.T) *Service {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "t.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.DB.Close() })
	return New(&config.Config{SystemPrompt: "出厂默认", SystemPromptDefault: "出厂默认"}, st, nil)
}

// T4: restore copies old content to a NEW head version; history is
// append-only and the effective prompt follows the highest version.
func TestPromptRevisionLifecycle(t *testing.T) {
	svc := promptTestService(t)

	// no revisions yet → env fallback
	if got := svc.EffectiveSystemPrompt(); got != "出厂默认" {
		t.Fatalf("env fallback broken: %q", got)
	}

	if _, serr := svc.AdminUpdatePrompt("第一版", "tester"); serr != nil {
		t.Fatalf("save v1: %+v", serr)
	}
	info, serr := svc.AdminUpdatePrompt("第二版", "tester")
	if serr != nil {
		t.Fatalf("save v2: %+v", serr)
	}
	if info["version"] != int64(2) {
		t.Fatalf("version must be 2, got %v", info["version"])
	}
	if got := svc.EffectiveSystemPrompt(); got != "第二版" {
		t.Fatalf("effective prompt must be latest, got %q", got)
	}

	// restore v1 → new version 3 with v1 content; rows only grow
	restored, serr := svc.AdminRestorePrompt(1, "tester")
	if serr != nil {
		t.Fatalf("restore: %+v", serr)
	}
	if restored["version"] != int64(3) || restored["content"] != "第一版" {
		t.Fatalf("restore result: %+v", restored)
	}
	if got := svc.EffectiveSystemPrompt(); got != "第一版" {
		t.Fatalf("restored prompt must be effective, got %q", got)
	}
	revs, serr := svc.AdminListPromptRevisions()
	if serr != nil || len(revs) != 3 {
		t.Fatalf("history must keep all 3 versions: %+v %d", serr, len(revs))
	}
	if revs[0]["version"] != int64(3) || revs[2]["version"] != int64(1) {
		t.Fatalf("revisions must be newest-first: %+v", revs)
	}

	// unknown version → 404
	if _, serr := svc.AdminRestorePrompt(99, "tester"); serr == nil || serr.Status != 404 {
		t.Fatalf("unknown version must 404, got %+v", serr)
	}
}

// T6: 10 concurrent saves must allocate 10 distinct consecutive versions.
func TestPromptConcurrentSavesUniqueVersions(t *testing.T) {
	svc := promptTestService(t)

	const n = 10
	versions := make([]int64, n)
	errs := make(chan error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			info, serr := svc.AdminUpdatePrompt(fmt.Sprintf("并发内容%d", i), "tester")
			if serr != nil {
				errs <- serr
				return
			}
			versions[i] = info["version"].(int64)
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent save failed: %+v", err)
	}

	seen := map[int64]bool{}
	for _, v := range versions {
		if v < 1 || v > n || seen[v] {
			t.Fatalf("versions must be exactly 1..%d, saw %v", n, versions)
		}
		seen[v] = true
	}
}
