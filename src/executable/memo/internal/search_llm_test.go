package internal

import (
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fakeSearchLLMProvider struct {
	enabled   bool
	rewriteFn func(string) (*searchLLMRewriteResult, error)
	rerankFn  func(searchQueryPlan, []searchLLMRerankCandidate) (*searchLLMRerankResponse, error)
}

func (f *fakeSearchLLMProvider) Enabled() bool {
	return f != nil && f.enabled
}

func (f *fakeSearchLLMProvider) RewriteQuery(query string) (*searchLLMRewriteResult, error) {
	if f.rewriteFn == nil {
		return nil, nil
	}
	return f.rewriteFn(query)
}

func (f *fakeSearchLLMProvider) RerankResults(plan searchQueryPlan, candidates []searchLLMRerankCandidate) (*searchLLMRerankResponse, error) {
	if f.rerankFn == nil {
		return nil, nil
	}
	return f.rerankFn(plan, candidates)
}

func withFakeSearchLLMProvider(t *testing.T, provider searchLLMProvider) {
	t.Helper()
	old := currentSearchLLMProvider
	currentSearchLLMProvider = provider
	t.Cleanup(func() {
		currentSearchLLMProvider = old
	})
}

func withSearchLLMHardFilter(t *testing.T, enabled bool) {
	t.Helper()
	old := currentSearchLLMHardFilterEnabled
	currentSearchLLMHardFilterEnabled = enabled
	t.Cleanup(func() {
		currentSearchLLMHardFilterEnabled = old
	})
}

func TestSearchRecordsUsesLLMRewriteToExpandCrossLingualIntent(t *testing.T) {
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	withFakeSearchEmbeddingProvider(t, &fakeSearchEmbeddingProvider{enabled: false})
	withFakeSearchLLMProvider(t, &fakeSearchLLMProvider{
		enabled: true,
		rewriteFn: func(query string) (*searchLLMRewriteResult, error) {
			if query != "handoff doc" {
				t.Fatalf("unexpected query rewrite input: %q", query)
			}
			return &searchLLMRewriteResult{Rewrites: []string{"交接文档 链接"}, PreferredTags: []string{"links"}, NeedURL: true}, nil
		},
	})

	best := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"links", "交接文档"},
		AddDate:      time.Unix(1_701_000_000, 0),
		ModifiedDate: time.Unix(1_701_000_000, 0),
		MyProblem:    true,
		Title:        "风神交接文档：https://handoff.example.com/doc",
	}
	noise := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"doc"},
		AddDate:      time.Unix(1_701_000_100, 0),
		ModifiedDate: time.Unix(1_701_000_100, 0),
		MyProblem:    true,
		Title:        "团队值班文档：https://duty.example.com/doc",
	}
	best.Save(true)
	noise.Save(true)

	results := SearchRecords("handoff doc", 10, true, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected llm-expanded search results")
	}
	if results[0].Record.ID != best.ID {
		t.Fatalf("expected llm rewrite to rank handoff document first, got %+v", results[0].Record)
	}
}

func TestSearchRecordsUsesLLMRerankPreviewLines(t *testing.T) {
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	withFakeSearchEmbeddingProvider(t, &fakeSearchEmbeddingProvider{enabled: false})
	withFakeSearchLLMProvider(t, &fakeSearchLLMProvider{
		enabled: true,
		rewriteFn: func(query string) (*searchLLMRewriteResult, error) {
			return &searchLLMRewriteResult{Rewrites: []string{"交接文档"}, PreferredTags: []string{"links"}}, nil
		},
		rerankFn: func(plan searchQueryPlan, candidates []searchLLMRerankCandidate) (*searchLLMRerankResponse, error) {
			if len(candidates) < 2 {
				t.Fatalf("expected at least two rerank candidates, got %d", len(candidates))
			}
			return &searchLLMRerankResponse{Results: []searchLLMRerankItem{
				{Index: 1, Score: 0.99, LineNumbers: []int{2, 3}},
				{Index: 0, Score: 0.41, LineNumbers: []int{1}},
			}}, nil
		},
	})

	base := time.Unix(1_701_100_000, 0)
	first := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"links"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "handoff auto mv guide",
	}
	best := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"links"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Title:        "24. auto-mv guide: https://auto.example.com\n3. 风神交接文档:\nhttps://handoff.example.com/doc",
	}
	first.Save(true)
	best.Save(true)

	results := SearchRecords("handoff", 10, true, nil, false, false)
	if len(results) < 2 {
		t.Fatalf("expected rerank search results, got %d", len(results))
	}
	if results[0].Record.ID != best.ID {
		t.Fatalf("expected llm rerank to move handoff doc record first, got %+v", results[0].Record)
	}
	joined := strings.Join(results[0].Preview, "\n")
	if !strings.Contains(joined, "风神交接文档") || !strings.Contains(joined, "https://handoff.example.com/doc") {
		t.Fatalf("expected llm preview lines to include handoff doc and URL, got %q", joined)
	}
	if strings.Contains(joined, "auto-mv guide") {
		t.Fatalf("expected llm preview lines to avoid the generic heading, got %q", joined)
	}
}

func TestSearchRecordsUsesLLMRerankAsFilter(t *testing.T) {
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	withFakeSearchEmbeddingProvider(t, &fakeSearchEmbeddingProvider{enabled: false})
	withSearchLLMHardFilter(t, true)
	withFakeSearchLLMProvider(t, &fakeSearchLLMProvider{
		enabled: true,
		rerankFn: func(plan searchQueryPlan, candidates []searchLLMRerankCandidate) (*searchLLMRerankResponse, error) {
			bestIndex := -1
			secondaryIndex := -1
			noiseIndex := -1
			for _, candidate := range candidates {
				joined := strings.Join(candidate.Lines, "\n")
				switch {
				case strings.Contains(joined, "handoff.example.com/doc"):
					bestIndex = candidate.Index
				case strings.Contains(joined, "handoff checklist"):
					secondaryIndex = candidate.Index
				case strings.Contains(joined, "handoff dashboard"):
					noiseIndex = candidate.Index
				}
			}
			if bestIndex < 0 || secondaryIndex < 0 || noiseIndex < 0 {
				t.Fatalf("expected best, secondary, and noise candidates to all reach rerank: %+v", candidates)
			}
			return &searchLLMRerankResponse{Results: []searchLLMRerankItem{
				{Index: bestIndex, Score: 0.99, LineNumbers: []int{1}},
				{Index: secondaryIndex, Score: 0.87, LineNumbers: []int{1}},
			}}, nil
		},
	})

	base := time.Unix(1_701_200_000, 0)
	secondary := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"doc"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "handoff checklist",
	}
	noise := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"dash"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Title:        "handoff dashboard",
	}
	best := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"links"},
		AddDate:      base.Add(2 * time.Second),
		ModifiedDate: base.Add(2 * time.Second),
		MyProblem:    true,
		Title:        "handoff doc：https://handoff.example.com/doc",
	}
	secondary.Save(true)
	noise.Save(true)
	best.Save(true)

	results := SearchRecords("handoff", 10, false, nil, false, false)
	if len(results) != 2 {
		t.Fatalf("expected llm rerank to filter out omitted candidates, got %d: %+v", len(results), results)
	}
	if results[0].Record.ID != best.ID {
		t.Fatalf("expected best handoff doc first after llm filter, got %+v", results[0].Record)
	}
	for _, result := range results {
		if result.Record.ID == noise.ID {
			t.Fatalf("expected omitted candidate to be removed by llm filter: %+v", results)
		}
	}
}
