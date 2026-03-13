package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/grewwc/go_tools/src/sortw"
	"github.com/grewwc/go_tools/src/utilsw"
)

const (
	searchLLMEnabledConfigName    = "memo.search.llm.enabled"
	searchLLMEndpointConfigName   = "memo.search.llm.endpoint"
	searchLLMModelConfigName      = "memo.search.llm.model"
	searchLLMHardFilterConfigName = "memo.search.llm.hard_filter"

	defaultSearchLLMEndpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	defaultSearchLLMModel    = "qwen-flash"

	searchLLMRequestTimeout      = 18 * time.Second
	searchLLMMaxRewrites         = 4
	searchLLMMaxRerankCandidates = 8
	searchLLMMaxRerankResults    = 6
	searchLLMMaxRecordRunes      = 2200
)

type searchLLMProvider interface {
	Enabled() bool
	RewriteQuery(query string) (*searchLLMRewriteResult, error)
	RerankResults(plan searchQueryPlan, candidates []searchLLMRerankCandidate) (*searchLLMRerankResponse, error)
}

type remoteSearchLLMProvider struct {
	enabled  bool
	apiKey   string
	endpoint string
	model    string
	client   *http.Client

	mu      sync.Mutex
	broken  bool
	lastErr error
}

type searchLLMRewriteResult struct {
	Rewrites      []string `json:"rewrites"`
	PreferredTags []string `json:"preferred_tags"`
	NeedURL       bool     `json:"need_url"`
}

type searchLLMRerankCandidate struct {
	Index     int
	BaseScore float64
	Tags      []string
	Lines     []string
}

type searchLLMRerankResponse struct {
	Results []searchLLMRerankItem `json:"results"`
}

type searchLLMRerankItem struct {
	Index       int     `json:"index"`
	Score       float64 `json:"score"`
	LineNumbers []int   `json:"line_numbers"`
}

type searchLLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type searchLLMRequest struct {
	Model    string             `json:"model"`
	Messages []searchLLMMessage `json:"messages"`
	Stream   bool               `json:"stream"`
}

type searchLLMResponse struct {
	Choices []searchLLMChoice  `json:"choices"`
	Error   *searchLLMAPIError `json:"error,omitempty"`
}

type searchLLMChoice struct {
	Message searchLLMChoiceMessage `json:"message"`
}

type searchLLMChoiceMessage struct {
	Content any `json:"content"`
}

type searchLLMAPIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

var currentSearchLLMProvider searchLLMProvider = newRemoteSearchLLMProvider()
var currentSearchLLMHardFilterEnabled = loadSearchLLMHardFilterEnabled()

func loadSearchLLMHardFilterEnabled() bool {
	config := utilsw.GetAllConfig()
	if config == nil {
		return false
	}
	return !isFalseSearchConfigValue(strings.TrimSpace(config.GetOrDefault(searchLLMHardFilterConfigName, "false").(string)))
}

func newRemoteSearchLLMProvider() searchLLMProvider {
	config := utilsw.GetAllConfig()
	if config == nil {
		return &remoteSearchLLMProvider{}
	}
	apiKey := strings.TrimSpace(config.GetOrDefault("api_key", "").(string))
	enabled := !isFalseSearchConfigValue(strings.TrimSpace(config.GetOrDefault(searchLLMEnabledConfigName, "true").(string)))
	endpoint := strings.TrimSpace(config.GetOrDefault(searchLLMEndpointConfigName, "").(string))
	if endpoint == "" {
		endpoint = strings.TrimSpace(config.GetOrDefault("ai.model.endpoint", defaultSearchLLMEndpoint).(string))
	}
	if endpoint == "" {
		endpoint = defaultSearchLLMEndpoint
	}
	model := strings.TrimSpace(config.GetOrDefault(searchLLMModelConfigName, defaultSearchLLMModel).(string))
	if model == "" {
		model = defaultSearchLLMModel
	}
	return &remoteSearchLLMProvider{
		enabled:  enabled,
		apiKey:   apiKey,
		endpoint: endpoint,
		model:    model,
		client:   &http.Client{Timeout: searchLLMRequestTimeout},
	}
}

func (p *remoteSearchLLMProvider) Enabled() bool {
	if p == nil {
		return false
	}
	if !p.enabled || p.apiKey == "" || p.endpoint == "" || p.model == "" {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return !p.broken
}

func (p *remoteSearchLLMProvider) RewriteQuery(query string) (*searchLLMRewriteResult, error) {
	if !p.Enabled() || strings.TrimSpace(query) == "" {
		return nil, nil
	}
	content, err := p.completePrompt(
		"You expand memo search queries for retrieval. Return JSON only and do not add markdown fences.",
		fmt.Sprintf(`User query: %q

Return JSON with this shape:
{"rewrites":["..."],"preferred_tags":["..."],"need_url":false}

Rules:
- Provide at most %d rewrites.
- Each rewrite should be a short retrieval phrase, not a sentence.
- Preserve entity names and add cross-lingual aliases when useful.
- Add link/url/doc/run intent only when strongly implied by the query.
- preferred_tags should only include strong hints such as links, db, doc, runbook.
- Do not repeat the original query verbatim unless adding a clearly useful intent word.
- Return valid JSON only.`, query, searchLLMMaxRewrites),
	)
	if err != nil {
		p.markBroken(err)
		return nil, err
	}
	var result searchLLMRewriteResult
	if err = json.Unmarshal([]byte(extractSearchLLMJSON(content)), &result); err != nil {
		return nil, err
	}
	result.Rewrites = normalizeSearchLLMTextList(result.Rewrites, searchLLMMaxRewrites, query)
	result.PreferredTags = normalizeSearchLLMTextList(result.PreferredTags, 3, "")
	if len(result.Rewrites) == 0 && len(result.PreferredTags) == 0 && !result.NeedURL {
		return nil, nil
	}
	return &result, nil
}

func (p *remoteSearchLLMProvider) RerankResults(plan searchQueryPlan, candidates []searchLLMRerankCandidate) (*searchLLMRerankResponse, error) {
	if !p.Enabled() || len(candidates) == 0 {
		return nil, nil
	}
	content, err := p.completePrompt(
		"You rerank memo search results and pick the most useful lines. Return JSON only and do not add markdown fences.",
		buildSearchLLMRerankPrompt(plan, candidates),
	)
	if err != nil {
		p.markBroken(err)
		return nil, err
	}
	var result searchLLMRerankResponse
	if err = json.Unmarshal([]byte(extractSearchLLMJSON(content)), &result); err != nil {
		return nil, err
	}
	result.Results = normalizeSearchLLMRerankItems(result.Results, candidates)
	if len(result.Results) == 0 {
		return nil, nil
	}
	return &result, nil
}

func (p *remoteSearchLLMProvider) completePrompt(systemPrompt, userPrompt string) (string, error) {
	requestBody := searchLLMRequest{
		Model: p.model,
		Messages: []searchLLMMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream: false,
	}
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var parsed searchLLMResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("parse llm response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		msg := strings.TrimSpace(string(body))
		if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
			msg = parsed.Error.Message
		}
		return "", fmt.Errorf("llm request failed: %s", msg)
	}
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return "", fmt.Errorf("llm request failed: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm response is empty")
	}
	text := strings.TrimSpace(extractSearchLLMContent(parsed.Choices[0].Message.Content))
	if text == "" {
		return "", fmt.Errorf("llm response content is empty")
	}
	return text, nil
}

func (p *remoteSearchLLMProvider) markBroken(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.broken = true
	p.lastErr = err
}

func buildSearchLLMRerankPrompt(plan searchQueryPlan, candidates []searchLLMRerankCandidate) string {
	var builder strings.Builder
	builder.WriteString("User query: ")
	builder.WriteString(plan.Original)
	builder.WriteString("\n")
	rewrites := make([]string, 0, len(plan.Variants))
	for _, variant := range plan.Variants {
		if compactSearchText(variant.Text) == compactSearchText(plan.Original) {
			continue
		}
		rewrites = append(rewrites, variant.Text)
	}
	if len(rewrites) > 0 {
		builder.WriteString("Expanded rewrites: ")
		builder.WriteString(strings.Join(rewrites, " | "))
		builder.WriteString("\n")
	}
	if len(plan.PreferredTags) > 0 {
		builder.WriteString("Preferred tags: ")
		builder.WriteString(strings.Join(plan.PreferredTags, ", "))
		builder.WriteString("\n")
	}
	builder.WriteString(fmt.Sprintf(`
Return JSON with this shape:
{"results":[{"index":0,"score":0.98,"line_numbers":[1,2]}]}

Rules:
- index must refer to a candidate below.
- score must be between 0 and 1.
- line_numbers must be exact 1-based line numbers from that candidate's content.
- Prefer candidates that best answer the user's real intent, not just broad token overlap.
- Generic words like 链接/url/link are not enough by themselves; omit candidates that only match generic link intent but miss the main topic or entity.
- Dataset links are not document links unless the query explicitly asks for datasets.
- If the useful content is a heading plus the next detail/URL line, include both line numbers.
- Omit irrelevant candidates.
- Return at most %d results.

Candidates:
`, searchLLMMaxRerankResults))
	for _, candidate := range candidates {
		builder.WriteString(fmt.Sprintf("[candidate %d]\n", candidate.Index))
		builder.WriteString(fmt.Sprintf("base_score: %.3f\n", candidate.BaseScore))
		if len(candidate.Tags) > 0 {
			builder.WriteString("tags: ")
			builder.WriteString(strings.Join(candidate.Tags, ", "))
			builder.WriteString("\n")
		}
		builder.WriteString("content:\n")
		for idx, line := range candidate.Lines {
			builder.WriteString(fmt.Sprintf("[line %d] %s\n", idx+1, line))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func buildSearchLLMRerankCandidate(index int, result *SearchResult) searchLLMRerankCandidate {
	return searchLLMRerankCandidate{
		Index:     index,
		BaseScore: result.Score,
		Tags:      append([]string(nil), result.Record.Tags...),
		Lines:     buildSearchLLMCandidateLines(result.Record),
	}
}

func buildSearchLLMCandidateLines(record *Record) []string {
	rawLines := strings.Split(strings.ReplaceAll(record.Title, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	totalRunes := 0
	for _, rawLine := range rawLines {
		line := strings.Join(strings.Fields(strings.TrimSpace(rawLine)), " ")
		if line == "" {
			continue
		}
		if totalRunes >= searchLLMMaxRecordRunes {
			break
		}
		remaining := searchLLMMaxRecordRunes - totalRunes
		line = truncateSearchPreview(line, minInt(remaining, 220))
		if line == "" {
			continue
		}
		lines = append(lines, line)
		totalRunes += utf8.RuneCountInString(line) + 1
		if totalRunes >= searchLLMMaxRecordRunes {
			break
		}
	}
	return lines
}

func buildSearchPreviewFromLineNumbers(record *Record, lineNumbers []int) []string {
	if len(lineNumbers) == 0 {
		return nil
	}
	lines := buildSearchLLMCandidateLines(record)
	if len(lines) == 0 {
		return nil
	}
	preview := make([]string, 0, len(lineNumbers))
	seen := map[int]struct{}{}
	for _, lineNumber := range lineNumbers {
		if lineNumber < 1 || lineNumber > len(lines) {
			continue
		}
		if _, ok := seen[lineNumber]; ok {
			continue
		}
		seen[lineNumber] = struct{}{}
		preview = append(preview, formatSearchPreviewLine(lines[lineNumber-1], 180))
	}
	return preview
}

func normalizeSearchLLMTextList(values []string, limit int, original string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	originalCompact := compactSearchText(original)
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		compact := compactSearchText(trimmed)
		if compact == "" {
			continue
		}
		if originalCompact != "" && compact == originalCompact {
			continue
		}
		if _, ok := seen[compact]; ok {
			continue
		}
		seen[compact] = struct{}{}
		result = append(result, trimmed)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

func normalizeSearchLLMRerankItems(items []searchLLMRerankItem, candidates []searchLLMRerankCandidate) []searchLLMRerankItem {
	result := make([]searchLLMRerankItem, 0, len(items))
	seen := map[int]struct{}{}
	for _, item := range items {
		if item.Index < 0 || item.Index >= len(candidates) {
			continue
		}
		if _, ok := seen[item.Index]; ok {
			continue
		}
		seen[item.Index] = struct{}{}
		item.Score = clampFloat(item.Score, 0, 1)
		item.LineNumbers = normalizeSearchLLMLineNumbers(item.LineNumbers, len(candidates[item.Index].Lines))
		result = append(result, item)
		if len(result) >= searchLLMMaxRerankResults {
			break
		}
	}
	return result
}

func normalizeSearchLLMLineNumbers(lineNumbers []int, maxLine int) []int {
	result := make([]int, 0, len(lineNumbers))
	seen := map[int]struct{}{}
	for _, lineNumber := range lineNumbers {
		if lineNumber < 1 || lineNumber > maxLine {
			continue
		}
		if _, ok := seen[lineNumber]; ok {
			continue
		}
		seen[lineNumber] = struct{}{}
		result = append(result, lineNumber)
	}
	sortw.Sort(result, nil)
	return result
}

func extractSearchLLMContent(content any) string {
	switch value := content.(type) {
	case string:
		return value
	case []any:
		parts := make([]string, 0, len(value))
		for _, item := range value {
			if obj, ok := item.(map[string]any); ok {
				if text, ok := obj["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "")
	default:
		body, _ := json.Marshal(value)
		return string(body)
	}
}

func extractSearchLLMJSON(text string) string {
	trimmed := strings.TrimSpace(text)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return strings.TrimSpace(trimmed[start : end+1])
	}
	return trimmed
}
