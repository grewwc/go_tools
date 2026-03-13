package internal

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/grewwc/go_tools/src/sortw"
	"github.com/grewwc/go_tools/src/utilsw"
	"golang.org/x/text/unicode/norm"
)

const (
	searchEmbeddingEnabledConfigName  = "memo.search.embedding.enabled"
	searchEmbeddingEndpointConfigName = "memo.search.embedding.endpoint"
	searchEmbeddingModelConfigName    = "memo.search.embedding.model"
	searchEmbeddingCacheConfigName    = "memo.search.embedding.cache"

	defaultSearchEmbeddingEndpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings"
	defaultSearchEmbeddingModel    = "text-embedding-v4"
	defaultSearchEmbeddingCache    = "~/.go_tools_memo_embedding_cache.json"

	searchEmbeddingBatchSize      = 16
	searchEmbeddingRequestTimeout = 45 * time.Second
	searchEmbeddingCacheVersion   = 1

	searchSemanticOnlyThreshold    = 0.55
	searchSemanticAssistThreshold  = 0.24
	searchPreviewSemanticThreshold = 0.42
	searchEmbeddingMaxRecordRunes  = 4096
	searchEmbeddingMaxQueryRunes   = 512
	searchEmbeddingMaxLineRunes    = 512
)

type searchEmbeddingProvider interface {
	Enabled() bool
	EmbedTexts(texts []string) ([][]float64, error)
}

type remoteSearchEmbeddingProvider struct {
	enabled   bool
	apiKey    string
	endpoint  string
	model     string
	cachePath string
	client    *http.Client

	mu         sync.Mutex
	cache      searchEmbeddingCacheFile
	cacheReady bool
	cacheDirty bool
	broken     bool
	lastErr    error
}

type searchEmbeddingCacheFile struct {
	Version int                             `json:"version"`
	Texts   map[string]searchEmbeddingEntry `json:"texts"`
}

type searchEmbeddingEntry struct {
	Vector      []float64 `json:"vector"`
	UpdatedUnix int64     `json:"updated_unix"`
}

type searchEmbeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

type searchEmbeddingResponse struct {
	Data  []searchEmbeddingResponseData `json:"data"`
	Error *searchEmbeddingResponseError `json:"error,omitempty"`
}

type searchEmbeddingResponseData struct {
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type searchEmbeddingResponseError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

var currentSearchEmbeddingProvider searchEmbeddingProvider = newRemoteSearchEmbeddingProvider()

func newRemoteSearchEmbeddingProvider() searchEmbeddingProvider {
	config := utilsw.GetAllConfig()
	if config == nil {
		return &remoteSearchEmbeddingProvider{}
	}
	apiKey := strings.TrimSpace(config.GetOrDefault("api_key", "").(string))
	enabled := !isFalseSearchConfigValue(strings.TrimSpace(config.GetOrDefault(searchEmbeddingEnabledConfigName, "true").(string)))
	endpoint := strings.TrimSpace(config.GetOrDefault(searchEmbeddingEndpointConfigName, "").(string))
	if endpoint == "" {
		endpoint = deriveSearchEmbeddingEndpoint(strings.TrimSpace(config.GetOrDefault("ai.model.endpoint", "").(string)))
	}
	model := strings.TrimSpace(config.GetOrDefault(searchEmbeddingModelConfigName, defaultSearchEmbeddingModel).(string))
	cachePath := strings.TrimSpace(config.GetOrDefault(searchEmbeddingCacheConfigName, defaultSearchEmbeddingCache).(string))
	return &remoteSearchEmbeddingProvider{
		enabled:   enabled,
		apiKey:    apiKey,
		endpoint:  endpoint,
		model:     model,
		cachePath: utilsw.ExpandUser(cachePath),
		client:    &http.Client{Timeout: searchEmbeddingRequestTimeout},
	}
}

func (p *remoteSearchEmbeddingProvider) Enabled() bool {
	if p == nil {
		return false
	}
	if !p.enabled || p.apiKey == "" || p.model == "" || p.endpoint == "" {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return !p.broken
}

func (p *remoteSearchEmbeddingProvider) EmbedTexts(texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	if len(texts) == 0 {
		return result, nil
	}
	if !p.Enabled() {
		return result, nil
	}

	preparedTexts := make([]string, len(texts))
	cacheKeys := make([]string, len(texts))
	missingTextByKey := map[string]string{}

	p.mu.Lock()
	cache := p.loadCacheLocked()
	for i, text := range texts {
		prepared := prepareSearchEmbeddingText(text)
		preparedTexts[i] = prepared
		if prepared == "" {
			continue
		}
		key := searchEmbeddingCacheKey(p.model, prepared)
		cacheKeys[i] = key
		if entry, ok := cache.Texts[key]; ok && len(entry.Vector) > 0 {
			result[i] = append([]float64(nil), entry.Vector...)
			continue
		}
		missingTextByKey[key] = prepared
	}
	p.mu.Unlock()

	if len(missingTextByKey) > 0 {
		keys := make([]string, 0, len(missingTextByKey))
		missingTexts := make([]string, 0, len(missingTextByKey))
		for key, text := range missingTextByKey {
			keys = append(keys, key)
			missingTexts = append(missingTexts, text)
		}
		fetched, err := p.fetchAndCacheEmbeddings(keys, missingTexts)
		if err != nil {
			p.markBroken(err)
			return nil, err
		}
		for i, key := range cacheKeys {
			if len(result[i]) > 0 || key == "" {
				continue
			}
			if vector, ok := fetched[key]; ok {
				result[i] = append([]float64(nil), vector...)
			}
		}
	}

	p.mu.Lock()
	cache = p.loadCacheLocked()
	for i, key := range cacheKeys {
		if len(result[i]) > 0 || key == "" {
			continue
		}
		if entry, ok := cache.Texts[key]; ok && len(entry.Vector) > 0 {
			result[i] = append([]float64(nil), entry.Vector...)
		}
	}
	p.mu.Unlock()
	return result, nil
}

func (p *remoteSearchEmbeddingProvider) fetchAndCacheEmbeddings(keys, texts []string) (map[string][]float64, error) {
	all := make(map[string][]float64, len(keys))
	for start := 0; start < len(texts); start += searchEmbeddingBatchSize {
		end := minInt(start+searchEmbeddingBatchSize, len(texts))
		batchKeys := keys[start:end]
		batchTexts := texts[start:end]
		vectors, err := p.fetchEmbeddings(batchTexts)
		if err != nil {
			return nil, err
		}
		if len(vectors) != len(batchTexts) {
			return nil, fmt.Errorf("embedding batch size mismatch: want %d got %d", len(batchTexts), len(vectors))
		}
		for i := range batchKeys {
			all[batchKeys[i]] = vectors[i]
		}
	}

	p.mu.Lock()
	cache := p.loadCacheLocked()
	for key, vector := range all {
		cache.Texts[key] = searchEmbeddingEntry{Vector: append([]float64(nil), vector...), UpdatedUnix: time.Now().Unix()}
		p.cacheDirty = true
	}
	_ = p.saveCacheLocked()
	p.mu.Unlock()

	return all, nil
}

func (p *remoteSearchEmbeddingProvider) fetchEmbeddings(texts []string) ([][]float64, error) {
	requestBody := searchEmbeddingRequest{
		Model:          p.model,
		Input:          texts,
		EncodingFormat: "float",
	}
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var parsed searchEmbeddingResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse embedding response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		msg := strings.TrimSpace(string(body))
		if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
			msg = parsed.Error.Message
		}
		return nil, fmt.Errorf("embedding request failed: %s", msg)
	}
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return nil, fmt.Errorf("embedding request failed: %s", parsed.Error.Message)
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("embedding response is empty")
	}
	sortw.StableSort(parsed.Data, func(d1, d2 searchEmbeddingResponseData) int {
		return d1.Index - d2.Index
	})
	// sort.SliceStable(parsed.Data, func(i, j int) bool {
	// 	return parsed.Data[i].Index < parsed.Data[j].Index
	// })
	result := make([][]float64, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		result = append(result, append([]float64(nil), item.Embedding...))
	}
	return result, nil
}

func (p *remoteSearchEmbeddingProvider) loadCacheLocked() *searchEmbeddingCacheFile {
	if p.cacheReady {
		return &p.cache
	}
	p.cache = searchEmbeddingCacheFile{Version: searchEmbeddingCacheVersion, Texts: map[string]searchEmbeddingEntry{}}
	p.cacheReady = true
	if p.cachePath == "" {
		return &p.cache
	}
	body, err := os.ReadFile(p.cachePath)
	if err != nil {
		return &p.cache
	}
	var parsed searchEmbeddingCacheFile
	if err = json.Unmarshal(body, &parsed); err != nil || parsed.Version != searchEmbeddingCacheVersion {
		return &p.cache
	}
	if parsed.Texts == nil {
		parsed.Texts = map[string]searchEmbeddingEntry{}
	}
	p.cache = parsed
	return &p.cache
}

func (p *remoteSearchEmbeddingProvider) saveCacheLocked() error {
	if !p.cacheDirty || p.cachePath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p.cachePath), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(p.cache, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := p.cachePath + ".tmp"
	if err = os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return err
	}
	if err = os.Rename(tmpPath, p.cachePath); err != nil {
		return err
	}
	p.cacheDirty = false
	return nil
}

func (p *remoteSearchEmbeddingProvider) markBroken(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.broken = true
	p.lastErr = err
}

func deriveSearchEmbeddingEndpoint(chatEndpoint string) string {
	chatEndpoint = strings.TrimSpace(chatEndpoint)
	if chatEndpoint == "" {
		return defaultSearchEmbeddingEndpoint
	}
	if strings.HasSuffix(chatEndpoint, "/chat/completions") {
		return strings.TrimSuffix(chatEndpoint, "/chat/completions") + "/embeddings"
	}
	return strings.TrimRight(chatEndpoint, "/") + "/embeddings"
}

func isFalseSearchConfigValue(val string) bool {
	val = strings.TrimSpace(strings.ToLower(val))
	switch val {
	case "0", "false", "no", "off", "disabled":
		return true
	default:
		return false
	}
}

func searchEmbeddingCacheKey(model, text string) string {
	sum := sha256.Sum256([]byte(model + "\x00" + text))
	return hex.EncodeToString(sum[:])
}

func prepareSearchEmbeddingText(text string) string {
	text = norm.NFKC.String(strings.TrimSpace(text))
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "<sep>", "\n")
	text = searchURLPattern.ReplaceAllString(text, " link ")
	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(strings.TrimSpace(line)), " ")
		if line == "" {
			continue
		}
		cleaned = append(cleaned, line)
	}
	return strings.Join(cleaned, "\n")
}

func buildSearchEmbeddingTextForQuery(query string) string {
	return truncateSearchEmbeddingText(prepareSearchEmbeddingText(query), searchEmbeddingMaxQueryRunes)
}

func buildSearchEmbeddingTextForRecord(record *Record) string {
	parts := make([]string, 0, len(record.Tags)+8)
	if len(record.Tags) > 0 {
		parts = append(parts, "tags: "+strings.Join(record.Tags, ", "))
	}
	for _, rawLine := range strings.Split(strings.ReplaceAll(record.Title, "\r\n", "\n"), "\n") {
		line := prepareSearchEmbeddingText(rawLine)
		if line == "" {
			continue
		}
		parts = append(parts, line)
	}
	return truncateSearchEmbeddingText(strings.Join(parts, "\n"), searchEmbeddingMaxRecordRunes)
}

func buildSearchEmbeddingTextForPreview(text string) string {
	return truncateSearchEmbeddingText(prepareSearchEmbeddingText(text), searchEmbeddingMaxLineRunes)
}

func truncateSearchEmbeddingText(text string, maxRunes int) string {
	if maxRunes <= 0 || text == "" {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return string(runes)
	}
	return string(runes[:maxRunes])
}

func searchEmbeddingScores(queryText string, candidateTexts []string) ([]float64, bool) {
	provider := currentSearchEmbeddingProvider
	if provider == nil || !provider.Enabled() {
		return nil, false
	}
	queryText = buildSearchEmbeddingTextForQuery(queryText)
	if queryText == "" || len(candidateTexts) == 0 {
		return nil, false
	}
	inputs := make([]string, 0, len(candidateTexts)+1)
	inputs = append(inputs, queryText)
	for _, text := range candidateTexts {
		inputs = append(inputs, text)
	}
	vectors, err := provider.EmbedTexts(inputs)
	if err != nil || len(vectors) != len(inputs) || len(vectors[0]) == 0 {
		return nil, false
	}
	queryVector := vectors[0]
	scores := make([]float64, len(candidateTexts))
	for i := range candidateTexts {
		scores[i] = cosineSimilarity(queryVector, vectors[i+1])
	}
	return scores, true
}

func cosineSimilarity(left, right []float64) float64 {
	if len(left) == 0 || len(left) != len(right) {
		return 0
	}
	dot := 0.0
	leftNorm := 0.0
	rightNorm := 0.0
	for i := range left {
		dot += left[i] * right[i]
		leftNorm += left[i] * left[i]
		rightNorm += right[i] * right[i]
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}

func combineHybridSearchScore(query searchDocument, lexicalScore, semanticScore float64) float64 {
	semanticScore = clampFloat(semanticScore, 0, 1)
	if lexicalScore <= 0 {
		if hasStrongShortQueryToken(query.tokens) || semanticScore < searchSemanticOnlyThreshold {
			return 0
		}
		return 1.1 + 2.2*(semanticScore-searchSemanticOnlyThreshold)
	}
	if semanticScore < searchSemanticAssistThreshold {
		return lexicalScore
	}
	lexicalNorm := normalizeLexicalSearchScore(lexicalScore)
	combined := lexicalScore + 1.45*semanticScore + 0.55*lexicalNorm
	if semanticScore > 0.58 {
		combined += 0.35
	}
	if hasStrongShortQueryToken(query.tokens) && lexicalScore < 0.85 && semanticScore < 0.75 {
		return lexicalScore
	}
	return maxFloat(lexicalScore, combined)
}

func normalizeLexicalSearchScore(score float64) float64 {
	if score <= 0 {
		return 0
	}
	return math.Min(score/3.25, 1.25)
}

func hasStrongShortQueryToken(tokens []string) bool {
	for _, token := range tokens {
		if requiresStrongSearchTokenMatch(token) {
			return true
		}
	}
	return false
}

func augmentSearchPreviewCandidatesWithSemanticScores(query string, candidates []searchPreviewCandidate) []searchPreviewCandidate {
	if len(candidates) == 0 {
		return candidates
	}
	texts := make([]string, len(candidates))
	for i, candidate := range candidates {
		texts[i] = buildSearchEmbeddingTextForPreview(candidate.text)
	}
	scores, ok := searchEmbeddingScores(query, texts)
	if !ok {
		return candidates
	}
	for i, semantic := range scores {
		semantic = clampFloat(semantic, 0, 1)
		if semantic < searchPreviewSemanticThreshold {
			continue
		}
		boost := 0.92 + 1.18*semantic
		if candidates[i].score <= 0 {
			candidates[i].score = boost
		} else {
			candidates[i].score = maxFloat(candidates[i].score, boost)
			if semantic > 0.5 {
				candidates[i].score += 0.22 * semantic
			}
		}
		if candidates[i].hasURL && semantic > 0.6 {
			candidates[i].score += 0.06
		}
	}
	return candidates
}

func clampFloat(value, lower, upper float64) float64 {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}
