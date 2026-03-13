package internal

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/grewwc/go_tools/src/cw"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupSQLiteTest(t *testing.T) func() {
	t.Helper()
	tempDir := t.TempDir()
	oldHomeDir := homeDir
	oldSQLitePath := localSQLitePath
	oldSQLiteReady := localSQLiteReady
	oldBackendMode := localBackendMode.Get().(string)
	oldClient := Client
	oldLocalMongoChecked := localMongoChecked
	oldLocalMongoErr := localMongoErr
	oldRemote := Remote.Get().(bool)
	oldListSpecial := ListSpecial
	oldPatterns := SpecialTagPatterns

	homeDir = tempDir
	localSQLitePath = filepath.Join(tempDir, "memo.sqlite3")
	localSQLiteReady = false
	Client = nil
	localMongoChecked = false
	localMongoErr = nil
	SetLocalBackendMode(LocalBackendSQLite)
	Remote.Set(false)
	ListSpecial = false
	SpecialTagPatterns = cw.NewSet("learn")
	initLocalSQLite()

	return func() {
		homeDir = oldHomeDir
		localSQLitePath = oldSQLitePath
		localSQLiteReady = oldSQLiteReady
		Client = oldClient
		localMongoChecked = oldLocalMongoChecked
		localMongoErr = oldLocalMongoErr
		SetLocalBackendMode(oldBackendMode)
		Remote.Set(oldRemote)
		ListSpecial = oldListSpecial
		SpecialTagPatterns = oldPatterns
	}

}

func TestSQLiteQueriesAndTagCounts(t *testing.T) {
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_000_000, 0)
	r1 := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"todo", "urgent"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "Alpha task",
	}
	r2 := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"todo", "foo.bar"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Finished:     true,
		Title:        "Beta note",
	}
	r3 := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"learn.go", "misc"},
		AddDate:      base.Add(2 * time.Second),
		ModifiedDate: base.Add(2 * time.Second),
		MyProblem:    true,
		Title:        "Gamma learn",
	}

	r1.Save(true)
	r2.Save(true)
	r3.Save(true)

	records, _ := ListRecords(-1, false, false, []string{"todo"}, false, "", false)
	if len(records) != 1 || records[0].ID != r1.ID {
		t.Fatalf("exact tag query mismatch: %+v", records)
	}

	records, _ = ListRecords(-1, false, true, []string{"todo", "urgent"}, true, "", false)
	if len(records) != 1 || records[0].ID != r1.ID {
		t.Fatalf("and tag query mismatch: %+v", records)
	}

	records, _ = ListRecords(-1, false, true, []string{"foo"}, false, "", true)
	if len(records) != 1 || records[0].ID != r2.ID {
		t.Fatalf("substring tag query mismatch: %+v", records)
	}

	records, _ = ListRecords(-1, false, true, nil, false, "beta", false)
	if len(records) != 1 || records[0].ID != r2.ID {
		t.Fatalf("title query mismatch: %+v", records)
	}

	tags, err := ListTags(-1, false)
	if err != nil {
		t.Fatal(err)
	}
	tagCounts := make(map[string]int64, len(tags))
	for _, tag := range tags {
		tagCounts[tag.Name] = tag.Count
	}
	if _, ok := tagCounts["learn.go"]; ok {
		t.Fatalf("special tag should be filtered out: %+v", tagCounts)
	}
	if tagCounts["todo"] != 2 || tagCounts["urgent"] != 1 || tagCounts["foo.bar"] != 1 {
		t.Fatalf("unexpected tag counts: %+v", tagCounts)
	}

	ListSpecial = true
	tags, err = ListTags(-1, false)
	if err != nil {
		t.Fatal(err)
	}
	foundSpecial := false
	for _, tag := range tags {
		if tag.Name == "learn.go" {
			foundSpecial = true
			break
		}
	}
	if !foundSpecial {
		t.Fatal("expected special tag to appear when ListSpecial is enabled")
	}

	ListSpecial = false
	r1.Delete()
	tags, err = ListTags(-1, false)
	if err != nil {
		t.Fatal(err)
	}
	tagCounts = make(map[string]int64, len(tags))
	for _, tag := range tags {
		tagCounts[tag.Name] = tag.Count
	}
	if _, ok := tagCounts["urgent"]; ok {
		t.Fatalf("deleted tag count should be removed: %+v", tagCounts)
	}
	if tagCounts["todo"] != 1 {
		t.Fatalf("todo count should decrease after delete: %+v", tagCounts)
	}
}

func TestSQLiteEmptyResults(t *testing.T) {
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	records, written := ListRecords(-1, false, false, nil, false, "", false)
	if written {
		t.Fatal("expected no url info to be written for empty results")
	}
	if len(records) != 0 {
		t.Fatalf("expected empty record list, got %+v", records)
	}

	tags, err := ListTags(-1, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 0 {
		t.Fatalf("expected empty tag list, got %+v", tags)
	}

	r := &Record{ID: primitive.NewObjectID()}
	r.LoadByID()
	if !r.Invalid {
		t.Fatal("expected nonexistent record to be marked invalid")
	}
}
