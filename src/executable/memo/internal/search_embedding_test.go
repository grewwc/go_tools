package internal

import (
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fakeSearchEmbeddingProvider struct {
	enabled bool
	err     error
	vectors map[string][]float64
}

func (f *fakeSearchEmbeddingProvider) Enabled() bool {
	return f != nil && f.enabled
}

func (f *fakeSearchEmbeddingProvider) EmbedTexts(texts []string) ([][]float64, error) {
	if f.err != nil {
		return nil, f.err
	}
	result := make([][]float64, len(texts))
	for i, text := range texts {
		if vector, ok := f.vectors[text]; ok {
			result[i] = append([]float64(nil), vector...)
		}
	}
	return result, nil
}

func withFakeSearchEmbeddingProvider(t *testing.T, provider searchEmbeddingProvider) {
	t.Helper()
	old := currentSearchEmbeddingProvider
	currentSearchEmbeddingProvider = provider
	t.Cleanup(func() {
		currentSearchEmbeddingProvider = old
	})
}

func TestSearchRecordsUsesSemanticEmbeddingsForCrossLingualQuery(t *testing.T) {
	disableSearchLLMProvider(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	relevant := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"links"},
		AddDate:      time.Unix(1_700_900_000, 0),
		ModifiedDate: time.Unix(1_700_900_000, 0),
		MyProblem:    true,
		Title:        "风神交接文档：https://handoff.example.com/doc",
	}
	noise := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"dash"},
		AddDate:      time.Unix(1_700_900_100, 0),
		ModifiedDate: time.Unix(1_700_900_100, 0),
		MyProblem:    true,
		Title:        "性能指标仪表盘：https://dash.example.com",
	}
	relevant.Save(true)
	noise.Save(true)

	withFakeSearchEmbeddingProvider(t, &fakeSearchEmbeddingProvider{
		enabled: true,
		vectors: map[string][]float64{
			buildSearchEmbeddingTextForQuery("handoff"):                                  {1, 0},
			buildSearchEmbeddingTextForRecord(relevant):                                  {0.98, 0.02},
			buildSearchEmbeddingTextForRecord(noise):                                     {0.05, 0.95},
			buildSearchEmbeddingTextForPreview("风神交接文档：https://handoff.example.com/doc"): {0.97, 0.03},
			buildSearchEmbeddingTextForPreview("性能指标仪表盘：https://dash.example.com"):       {0.04, 0.96},
		},
	})

	results := SearchRecords("handoff", 10, true, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected semantic search results")
	}
	if results[0].Record.ID != relevant.ID {
		t.Fatalf("expected semantic hit to rank first, got %+v", results[0].Record)
	}
}

func TestSearchPreviewUsesSemanticScoringWhenLexicalSignalIsMissing(t *testing.T) {
	disableSearchLLMProvider(t)
	record := &Record{
		ID:   primitive.NewObjectID(),
		Tags: []string{"links"},
		Title: "风神交接文档：https://handoff.example.com/doc\n" +
			"性能指标仪表盘：https://dash.example.com",
	}

	withFakeSearchEmbeddingProvider(t, &fakeSearchEmbeddingProvider{
		enabled: true,
		vectors: map[string][]float64{
			buildSearchEmbeddingTextForQuery("handoff"):                                  {1, 0},
			buildSearchEmbeddingTextForPreview("风神交接文档：https://handoff.example.com/doc"): {0.99, 0.01},
			buildSearchEmbeddingTextForPreview("性能指标仪表盘：https://dash.example.com"):       {0.02, 0.98},
		},
	})

	preview := SearchPreview(record, "handoff")
	if len(preview) == 0 {
		t.Fatal("expected semantic preview lines")
	}
	joined := strings.Join(preview, "\n")
	if !strings.Contains(joined, "交接文档") {
		t.Fatalf("expected semantic preview to include handoff line, got %q", joined)
	}
	if strings.Contains(joined, "性能指标仪表盘") {
		t.Fatalf("expected semantic preview to omit unrelated dash line, got %q", joined)
	}
}
