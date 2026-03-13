package internal

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/sortw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/unicode/norm"
)

type SearchResult struct {
	Record  *Record
	Score   float64
	Preview []string
}

type searchQueryVariant struct {
	Text   string
	Doc    searchDocument
	Weight float64
}

type searchQueryPlan struct {
	Original      string
	Variants      []searchQueryVariant
	PreferredTags []string
}

type searchPreviewCandidate struct {
	text   string
	score  float64
	order  int
	isTag  bool
	hasURL bool
}

var searchHeadingPattern = regexp.MustCompile(`^\s*(\d+[.)、]|[-*•]|[一二三四五六七八九十]+[.)、])\s*`)

type searchDocument struct {
	normalized  string
	compact     string
	tokens      []string
	ngrams      map[string]int
	hasURL      bool
	requestsURL bool
}

var (
	searchSynonymGroups = [][]string{
		{"链接", "地址", "网址", "url", "uri", "link", "endpoint", "host", "域名", "官网", "site", "http", "https"},
		{"数据库", "database", "db"},
	}
	searchGenericSuffixes = []string{"相关", "有关", "内容", "方面", "资料", "信息", "情况", "问题", "事项", "记录"}
	searchIgnoredTokens   = map[string]struct{}{
		"相关": {},
		"有关": {},
		"内容": {},
		"方面": {},
		"资料": {},
		"信息": {},
		"情况": {},
		"问题": {},
		"事项": {},
		"记录": {},
	}
	searchURLPattern   = regexp.MustCompile(`(?i)https?://\S+`)
	searchSynonymIndex = map[string]int{}
	searchKnownTerms   []string
)

const (
	searchPreviewMaxLines               = 2
	searchMinInformativeCoverage        = 0.2
	searchMinInformativeCoverageWithURL = 0.26
)

const searchFloatCompareEpsilon = 1e-9

var searchDatabaseAccessHints = []string{
	"host",
	"port",
	"user",
	"passwd",
	"password",
	"mysql",
	"postgres",
	"postgresql",
	"redis",
	"mongo",
	"mongodb",
	"jdbc",
	"dsn",
	"consul",
	"访问",
	"连接",
	"账号",
	"用户名",
	"密码",
}

func init() {
	seen := map[string]struct{}{}
	for idx, group := range searchSynonymGroups {
		for _, term := range group {
			normalized := compactSearchText(term)
			if normalized == "" {
				continue
			}
			searchSynonymIndex[normalized] = idx
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			searchKnownTerms = append(searchKnownTerms, normalized)
		}
	}
}

func SearchRecords(query string, limit int64, includeFinished bool, tags []string, useAnd bool, prefix bool) []*SearchResult {
	plan := buildSearchQueryPlan(query)
	if len(plan.Variants) == 0 {
		return []*SearchResult{}
	}
	records, _ := listRecords(math.MaxInt64, false, includeFinished, tags, useAnd, "", prefix, false)
	results := rankSearchRecords(plan, records)
	results = filterSearchResults(results)
	results = rerankSearchResultsWithLLM(plan, results)
	if limit > 0 && len(results) > int(limit) {
		results = results[:limit]
	}
	writeSearchInfo(results)
	return results
}

func SearchHighlightPattern(query string) *regexp.Regexp {
	tokens := tokenizeSearchText(query)
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if utf8.RuneCountInString(token) < 2 {
			continue
		}
		parts = append(parts, regexp.QuoteMeta(token))
	}
	if len(parts) == 0 {
		compact := compactSearchText(query)
		if compact == "" {
			return nil
		}
		parts = append(parts, regexp.QuoteMeta(compact))
	}
	sortw.StableSort(parts, func(s1, s2 string) int {
		return utf8.RuneCountInString(s1) - utf8.RuneCountInString(s2)
	})
	pattern, err := regexp.Compile("(?i)" + strings.Join(parts, "|"))
	if err != nil {
		return nil
	}
	return pattern
}

func SearchPreview(record *Record, query string) []string {
	queryDoc := newSearchQueryDocument(query)
	if queryDoc.compact == "" {
		return fallbackSearchPreview(record)
	}
	candidates := buildSearchPreviewCandidates(record, queryDoc)
	candidates = augmentSearchPreviewCandidatesWithSemanticScores(query, candidates)
	selected := pickSearchPreviewCandidates(record, candidates)
	if len(selected) == 0 {
		return fallbackSearchPreview(record)
	}
	rawLines := strings.Split(strings.ReplaceAll(record.Title, "\r\n", "\n"), "\n")
	selected = expandSearchPreviewContext(rawLines, selected, queryDoc)
	sortw.StableSort(selected, func(s1, s2 searchPreviewCandidate) int {
		return s1.order - s2.order
	})
	previews := make([]string, 0, len(selected))
	for _, candidate := range selected {
		if candidate.isTag {
			continue
		}
		text := strings.TrimSpace(candidate.text)
		if text == "" {
			continue
		}
		previews = append(previews, formatSearchPreviewLine(text, 180))
	}
	if len(previews) == 0 {
		return fallbackSearchPreview(record)
	}
	return previews
}

func newSearchDocument(text string) searchDocument {
	return newSearchDocumentWithOptions(text, true)
}

func newSearchQueryDocument(text string) searchDocument {
	return newSearchDocumentWithOptions(text, !looksLikeExplicitURL(text))
}

func newSearchDocumentWithOptions(text string, trimURLs bool) searchDocument {
	normalized := normalizeSearchTextWithOptions(text, trimURLs)
	compact := compactSearchText(normalized)
	return searchDocument{
		normalized:  normalized,
		compact:     compact,
		tokens:      tokenizeSearchText(normalized),
		ngrams:      buildSearchNgrams(compact),
		hasURL:      containsURL(strings.ToLower(text)),
		requestsURL: searchRequestsURL(normalized),
	}
}

func buildSearchQueryPlan(query string) searchQueryPlan {
	plan := searchQueryPlan{Original: strings.TrimSpace(query)}
	plan.addVariant(plan.Original, 1)
	provider := currentSearchLLMProvider
	if provider == nil || !provider.Enabled() || strings.TrimSpace(query) == "" || !shouldUseLLMQueryRewrite(plan.Original) {
		return plan
	}
	rewrite, err := provider.RewriteQuery(plan.Original)
	if err != nil || rewrite == nil {
		return plan
	}
	for idx, rewriteText := range rewrite.Rewrites {
		weight := 0.96 - 0.06*float64(idx)
		if weight < 0.78 {
			weight = 0.78
		}
		plan.addVariant(rewriteText, weight)
	}
	if rewrite.NeedURL && !searchRequestsURL(plan.Original) {
		plan.addVariant(strings.TrimSpace(plan.Original+" 链接"), 0.97)
	}
	plan.PreferredTags = dedupeSearchTags(rewrite.PreferredTags)
	return plan
}

func shouldUseLLMQueryRewrite(query string) bool {
	queryDoc := newSearchQueryDocument(query)
	if queryDoc.compact == "" {
		return false
	}
	if len(queryDoc.tokens) <= 1 {
		return true
	}
	hasHan := false
	for _, r := range query {
		if unicode.In(r, unicode.Han) {
			hasHan = true
			break
		}
	}
	if !hasHan {
		return true
	}
	return !queryDoc.requestsURL && len(queryDoc.tokens) <= 2
}

func (plan *searchQueryPlan) addVariant(text string, weight float64) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return
	}
	doc := newSearchQueryDocument(trimmed)
	if doc.compact == "" {
		return
	}
	for idx, variant := range plan.Variants {
		if variant.Doc.compact != doc.compact {
			continue
		}
		if weight > variant.Weight {
			plan.Variants[idx].Text = trimmed
			plan.Variants[idx].Weight = weight
			plan.Variants[idx].Doc = doc
		}
		return
	}
	plan.Variants = append(plan.Variants, searchQueryVariant{Text: trimmed, Doc: doc, Weight: weight})
}

func dedupeSearchTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, tag := range tags {
		compact := compactSearchText(tag)
		if compact == "" {
			continue
		}
		if _, ok := seen[compact]; ok {
			continue
		}
		seen[compact] = struct{}{}
		result = append(result, strings.TrimSpace(tag))
	}
	return result
}

func rankSearchRecords(plan searchQueryPlan, records []*Record) []*SearchResult {
	semanticTexts := make([]string, len(records))
	for i, record := range records {
		semanticTexts[i] = buildSearchEmbeddingTextForRecord(record)
	}
	semanticScoresByVariant := make([][]float64, len(plan.Variants))
	semanticOKByVariant := make([]bool, len(plan.Variants))
	for idx, variant := range plan.Variants {
		semanticScoresByVariant[idx], semanticOKByVariant[idx] = searchEmbeddingScores(variant.Text, semanticTexts)
	}
	results := make([]*SearchResult, 0, len(records))
	for i, record := range records {
		bestScore := 0.0
		matchedVariants := 0
		for idx, variant := range plan.Variants {
			lexicalScore := scoreSearchRecord(variant.Doc, record)
			score := lexicalScore
			if semanticOKByVariant[idx] {
				score = combineHybridSearchScore(variant.Doc, lexicalScore, semanticScoresByVariant[idx][i])
			}
			if score <= 0 {
				continue
			}
			weighted := score * variant.Weight
			if weighted > bestScore {
				bestScore = weighted
			}
			if score >= 0.92 {
				matchedVariants++
			}
		}
		if bestScore <= 0 {
			continue
		}
		if matchedVariants > 1 {
			bestScore += 0.08 * float64(matchedVariants-1)
		}
		if recordMatchesSearchTags(record, plan.PreferredTags) {
			bestScore += 0.18
		}
		results = append(results, &SearchResult{Record: record, Score: bestScore})
	}
	sortSearchResults(results)
	return results
}

func scoreSearchRecord(query searchDocument, record *Record) float64 {
	titleDoc := newSearchDocument(record.Title)
	tagsText := strings.Join(record.Tags, " ")
	tagsDoc := newSearchDocument(tagsText)
	titleAndTagsDoc := newSearchDocument(strings.TrimSpace(record.Title + " " + tagsText))
	tagsAndTitleDoc := newSearchDocument(strings.TrimSpace(tagsText + " " + record.Title))

	titleScore := scoreSearchDocument(query, titleDoc)
	tagScore := scoreSearchDocument(query, tagsDoc)
	combinedScore := maxFloat(
		scoreSearchDocument(query, titleAndTagsDoc),
		scoreSearchDocument(query, tagsAndTitleDoc),
	)
	if titleScore <= 0 && tagScore <= 0 && combinedScore <= 0 {
		return 0
	}
	if searchNeedsDatabaseLinkContext(query) && !recordHasSearchDatabaseLinkContext(record) {
		return 0
	}
	best := maxFloat(titleScore, combinedScore, tagScore*0.94)
	if titleScore > 0 && tagScore > 0 {
		best += 0.35 * minFloat(titleScore, tagScore)
	}
	if tagScore > 0 {
		best += 0.12 * tagScore
	}
	lineTopScore, strongLineCount, strongURLLineCount := searchStrongLineStats(buildSearchPreviewCandidates(record, query))
	if lineTopScore > 0 {
		best = maxFloat(best, lineTopScore*0.98)
		if strongLineCount > 1 {
			best += 0.26 * float64(strongLineCount-1)
		}
		if strongURLLineCount > 1 {
			best += 0.08 * float64(strongURLLineCount-1)
		}
		if hasSearchTag(record, "links") && (strongLineCount > 0 || strongURLLineCount > 0) {
			best += 0.24
		}
	}
	return best
}

func recordMatchesSearchTags(record *Record, targetTags []string) bool {
	if len(targetTags) == 0 {
		return false
	}
	targetSet := map[string]struct{}{}
	for _, tag := range targetTags {
		compact := compactSearchText(tag)
		if compact == "" {
			continue
		}
		targetSet[compact] = struct{}{}
	}
	if len(targetSet) == 0 {
		return false
	}
	for _, tag := range record.Tags {
		if _, ok := targetSet[compactSearchText(tag)]; ok {
			return true
		}
	}
	return false
}

func filterSearchResults(results []*SearchResult) []*SearchResult {
	if len(results) == 0 {
		return results
	}
	topScore := results[0].Score
	if topScore < 0.55 {
		return []*SearchResult{}
	}
	if topScore < 1.15 {
		return results[:1]
	}
	minScore := math.Max(0.75, topScore*0.65)
	filtered := make([]*SearchResult, 0, len(results))
	for _, result := range results {
		if result.Score >= minScore {
			filtered = append(filtered, result)
		}
	}
	if len(filtered) == 0 {
		return results[:1]
	}
	return filtered
}

func rerankSearchResultsWithLLM(plan searchQueryPlan, results []*SearchResult) []*SearchResult {
	provider := currentSearchLLMProvider
	if provider == nil || !provider.Enabled() || len(results) == 0 {
		return results
	}
	limit := minInt(searchLLMMaxRerankCandidates, len(results))
	candidates := make([]searchLLMRerankCandidate, 0, limit)
	for idx := 0; idx < limit; idx++ {
		candidates = append(candidates, buildSearchLLMRerankCandidate(idx, results[idx]))
	}
	reranked, err := provider.RerankResults(plan, candidates)
	if err != nil || reranked == nil || len(reranked.Results) == 0 {
		return results
	}
	boostBase := results[0].Score + 3
	filtered := make([]*SearchResult, 0, len(reranked.Results))
	kept := map[int]struct{}{}
	for order, item := range reranked.Results {
		if item.Index < 0 || item.Index >= limit {
			continue
		}
		if _, ok := kept[item.Index]; ok {
			continue
		}
		preview := buildSearchPreviewFromLineNumbers(results[item.Index].Record, item.LineNumbers)
		if len(preview) > 0 {
			results[item.Index].Preview = preview
		}
		results[item.Index].Score = maxFloat(results[item.Index].Score, boostBase-float64(order)*0.01+clampFloat(item.Score, 0, 1)*0.001)
		filtered = append(filtered, results[item.Index])
		kept[item.Index] = struct{}{}
	}
	if len(filtered) == 0 {
		return results
	}
	if !currentSearchLLMHardFilterEnabled {
		sortSearchResults(results)
		return results
	}
	sortSearchResults(filtered)
	return filtered
}

func sortSearchResults(results []*SearchResult) {
	sortw.StableSort(results, func(r1, r2 *SearchResult) int {
		if scoreCmp := compareSearchFloatDesc(r1.Score, r2.Score); scoreCmp != 0 {
			return scoreCmp
		}
		if !r1.Record.ModifiedDate.Equal(r2.Record.ModifiedDate) {
			if r1.Record.ModifiedDate.After(r2.Record.ModifiedDate) {
				return -1
			}
			return 1
		}
		if r1.Record.AddDate.After(r2.Record.AddDate) {
			return -1
		}
		if r1.Record.AddDate.Before(r2.Record.AddDate) {
			return 1
		}
		return 0
	})
}

func writeSearchInfo(results []*SearchResult) {
	if len(results) == 0 {
		return
	}
	ids := make([]*primitive.ObjectID, len(results))
	titles := make([]string, len(results))
	for i, result := range results {
		ids[i] = &result.Record.ID
		titles[i] = result.Record.Title
	}
	WriteInfo(ids, titles)
}

func buildSearchPreviewCandidates(record *Record, query searchDocument) []searchPreviewCandidate {
	candidates := make([]searchPreviewCandidate, 0, len(record.Tags)+4)
	if len(record.Tags) > 0 {
		tagsLine := "tags: " + strings.Join(record.Tags, ", ")
		tagScore := scoreSearchDocument(query, newSearchDocument(tagsLine))
		if tagScore > 0 {
			candidates = append(candidates, searchPreviewCandidate{
				text:  tagsLine,
				score: tagScore + 0.08,
				order: -1,
				isTag: true,
			})
		}
	}
	lines := strings.Split(strings.ReplaceAll(record.Title, "\r\n", "\n"), "\n")
	for idx, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		score := scoreSearchDocument(query, newSearchDocument(line))
		hasURL := containsURL(strings.ToLower(line))
		if hasURL {
			score += 0.05
			if query.requestsURL {
				score += 0.12
			}
		}
		candidates = append(candidates, searchPreviewCandidate{
			text:   line,
			score:  score,
			order:  idx,
			hasURL: hasURL,
		})
	}
	return candidates
}

func pickSearchPreviewCandidates(record *Record, candidates []searchPreviewCandidate) []searchPreviewCandidate {
	if len(candidates) == 0 {
		return nil
	}
	sortw.StableSort(candidates, func(c1, c2 searchPreviewCandidate) int {
		if scoreCmp := compareSearchFloatDesc(c1.score, c2.score); scoreCmp != 0 {
			return scoreCmp
		}
		if c1.hasURL != c2.hasURL {
			if c1.hasURL {
				return -1
			}
			return 1
		}
		if c1.isTag != c2.isTag {
			if !c1.isTag {
				return -1
			}
			return 1
		}
		if c1.order < c2.order {
			return -1
		}
		if c1.order > c2.order {
			return 1
		}
		return 0
	})
	minScore := math.Max(0.85, candidates[0].score*0.72)
	limit := searchPreviewMaxLines
	if hasSearchTag(record, "links") {
		limit = len(candidates)
		minScore = math.Max(0.78, candidates[0].score*0.6)
	}
	selected := make([]searchPreviewCandidate, 0, limit)
	seen := map[string]struct{}{}
	hasTitle := false
	for _, candidate := range candidates {
		text := strings.TrimSpace(candidate.text)
		if text == "" {
			continue
		}
		if candidate.score < minScore {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		selected = append(selected, candidate)
		if !candidate.isTag {
			hasTitle = true
		}
		if len(selected) >= limit {
			break
		}
	}
	if len(selected) == 1 && selected[0].isTag {
		for _, candidate := range candidates {
			if candidate.isTag {
				continue
			}
			text := strings.TrimSpace(candidate.text)
			if text == "" {
				continue
			}
			if _, ok := seen[text]; ok {
				continue
			}
			selected = append(selected, candidate)
			hasTitle = true
			break
		}
	}
	if !hasTitle {
		for _, candidate := range candidates {
			if candidate.isTag {
				continue
			}
			text := strings.TrimSpace(candidate.text)
			if text == "" {
				continue
			}
			if _, ok := seen[text]; ok {
				continue
			}
			selected = append(selected, candidate)
			break
		}
	}
	return selected
}

func expandSearchPreviewContext(rawLines []string, selected []searchPreviewCandidate, query searchDocument) []searchPreviewCandidate {
	if len(selected) == 0 || len(rawLines) == 0 {
		return selected
	}
	result := make([]searchPreviewCandidate, 0, len(selected)+2)
	seen := map[string]struct{}{}
	appendCandidate := func(candidate searchPreviewCandidate) {
		text := strings.TrimSpace(candidate.text)
		if text == "" {
			return
		}
		key := fmt.Sprintf("%d:%s", candidate.order, text)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		result = append(result, candidate)
	}
	for _, candidate := range selected {
		appendCandidate(candidate)
		if candidate.isTag || candidate.hasURL || !looksLikeSearchHeadingLine(candidate.text) {
			continue
		}
		nextIdx := nextMeaningfulSearchLine(rawLines, candidate.order+1)
		if nextIdx < 0 {
			continue
		}
		nextText := strings.TrimSpace(rawLines[nextIdx])
		if nextText == "" || looksLikeSearchHeadingLine(nextText) {
			continue
		}
		nextScore := scoreSearchDocument(query, newSearchDocument(nextText))
		if nextScore <= 0 && !looksLikeSearchDetailLine(nextText) {
			continue
		}
		appendCandidate(searchPreviewCandidate{
			text:   nextText,
			score:  maxFloat(nextScore, candidate.score*0.7),
			order:  nextIdx,
			hasURL: containsURL(strings.ToLower(nextText)),
		})
	}
	return result
}

func fallbackSearchPreview(record *Record) []string {
	if strings.TrimSpace(record.Title) == "" {
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(record.Title, "\r\n", "\n"), "\n")
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		return []string{formatSearchPreviewLine(line, 180)}
	}
	return nil
}

func searchStrongLineStats(candidates []searchPreviewCandidate) (float64, int, int) {
	lineTopScore := 0.0
	for _, candidate := range candidates {
		if candidate.isTag {
			continue
		}
		if candidate.score > lineTopScore {
			lineTopScore = candidate.score
		}
	}
	if lineTopScore == 0 {
		return 0, 0, 0
	}
	strongLineCount := 0
	strongURLLineCount := 0
	minStrongScore := math.Max(0.95, lineTopScore*0.62)
	for _, candidate := range candidates {
		if candidate.isTag || candidate.score < minStrongScore {
			continue
		}
		strongLineCount++
		if candidate.hasURL {
			strongURLLineCount++
		}
	}
	return lineTopScore, strongLineCount, strongURLLineCount
}

func hasSearchTag(record *Record, target string) bool {
	target = compactSearchText(target)
	if target == "" {
		return false
	}
	for _, tag := range record.Tags {
		if compactSearchText(tag) == target {
			return true
		}
	}
	return false
}

func looksLikeSearchHeadingLine(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	if containsURL(strings.ToLower(text)) {
		return false
	}
	if searchHeadingPattern.MatchString(text) {
		return true
	}
	if strings.HasSuffix(text, ":") || strings.HasSuffix(text, "：") {
		return true
	}
	runes := []rune(text)
	if len(runes) > 40 {
		return false
	}
	for _, keyword := range []string{"run", "运行", "启动", "部署", "地址", "入口", "命令", "配置", "链接", "url"} {
		if strings.Contains(strings.ToLower(text), keyword) {
			return true
		}
	}
	return false
}

func looksLikeSearchDetailLine(text string) bool {
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" {
		return false
	}
	if containsURL(text) {
		return true
	}
	for _, keyword := range []string{"doas", "python", "go ", "npm ", "pnpm ", "yarn ", "curl ", "bash ", "ssh ", "kubectl", "make ", "./", "url", "命令", "运行", "步骤", "入口", "地址"} {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func nextMeaningfulSearchLine(lines []string, start int) int {
	for idx := start; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) != "" {
			return idx
		}
	}
	return -1
}

func isASCIIAlphaNumSearchToken(token string) bool {
	if token == "" {
		return false
	}
	for _, r := range token {
		if r >= utf8.RuneSelf {
			return false
		}
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func truncateSearchPreview(text string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}

func formatSearchPreviewLine(text string, maxRunes int) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if containsURL(strings.ToLower(text)) {
		return text
	}
	return truncateSearchPreview(text, maxRunes)
}

func scoreSearchDocument(query, candidate searchDocument) float64 {
	if query.compact == "" || candidate.compact == "" {
		return 0
	}
	coverage := searchCoverageScore(query.tokens, candidate.tokens)
	if coverage == 0 {
		return 0
	}
	if informativeTokens := filterSearchInformativeTokens(query.tokens); len(informativeTokens) > 0 {
		informativeCoverage := searchCoverageScore(informativeTokens, candidate.tokens)
		minInformativeCoverage := searchMinInformativeCoverage
		if query.requestsURL {
			minInformativeCoverage = searchMinInformativeCoverageWithURL
		}
		if informativeCoverage < minInformativeCoverage {
			return 0
		}
		coverage = maxFloat(coverage, 0.72*coverage+0.38*informativeCoverage)
	}
	ngram := diceCoefficient(query.ngrams, candidate.ngrams)
	whole := searchTokenSimilarity(query.compact, candidate.compact)
	if utf8.RuneCountInString(candidate.compact) > 96 {
		whole *= 0.4
	}
	exact := 0.0
	if strings.Contains(candidate.compact, query.compact) || strings.Contains(candidate.normalized, query.normalized) {
		exact = 1.25
	}
	prefix := 0.0
	if strings.HasPrefix(candidate.compact, query.compact) {
		prefix = 0.22
	}
	urlBonus := 0.0
	if query.requestsURL && candidate.hasURL {
		urlBonus = 0.18
	}
	return 1.35*coverage + 0.75*ngram + 0.35*whole + exact + prefix + urlBonus
}

func searchCoverageScore(queryTokens, candidateTokens []string) float64 {
	if len(queryTokens) == 0 || len(candidateTokens) == 0 {
		return 0
	}
	totalWeight := 0.0
	totalScore := 0.0
	for _, queryToken := range queryTokens {
		weight := float64(clampInt(utf8.RuneCountInString(queryToken), 1, 6))
		best := 0.0
		for _, candidateToken := range candidateTokens {
			score := searchTokenSimilarity(queryToken, candidateToken)
			if score > best {
				best = score
			}
		}
		if requiresStrongSearchTokenMatch(queryToken) && best < 0.85 {
			return 0
		}
		totalWeight += weight
		totalScore += best * weight
	}
	if totalWeight == 0 {
		return 0
	}
	return totalScore / totalWeight
}

func filterSearchInformativeTokens(tokens []string) []string {
	result := make([]string, 0, len(tokens))
	seen := map[string]struct{}{}
	appendToken := func(token string) {
		token = compactSearchText(token)
		if token == "" {
			return
		}
		if _, ignored := searchIgnoredTokens[token]; ignored {
			return
		}
		if _, ok := seen[token]; ok {
			return
		}
		seen[token] = struct{}{}
		result = append(result, token)
	}
	for _, token := range tokens {
		token = compactSearchText(token)
		if token == "" {
			continue
		}
		if _, ignored := searchIgnoredTokens[token]; ignored {
			continue
		}
		if group, ok := searchSynonymIndex[token]; ok && group == 0 {
			continue
		}
		stripped := stripSearchURLAffixes(token)
		if stripped != "" && stripped != token {
			appendToken(stripped)
			continue
		}
		appendToken(token)
	}
	return result
}

func searchNeedsDatabaseLinkContext(query searchDocument) bool {
	return query.requestsURL && searchContainsIntentGroupToken(query.tokens, 1)
}

func searchContainsIntentGroupToken(tokens []string, group int) bool {
	for _, token := range tokens {
		compact := compactSearchText(token)
		if compact == "" {
			continue
		}
		if synonymGroup, ok := searchSynonymIndex[compact]; ok && synonymGroup == group {
			return true
		}
	}
	return false
}

func recordHasSearchDatabaseLinkContext(record *Record) bool {
	rawLines := strings.Split(strings.ReplaceAll(record.Title, "\r\n", "\n"), "\n")
	hasURLLine := false
	for idx, rawLine := range rawLines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		lineDoc := newSearchDocument(line)
		if lineDoc.hasURL {
			hasURLLine = true
		}
		if !searchContainsIntentGroupToken(lineDoc.tokens, 1) {
			continue
		}
		if lineDoc.hasURL || looksLikeSearchDatabaseAccessLine(lineDoc.normalized) {
			return true
		}
		nextIdx := nextMeaningfulSearchLine(rawLines, idx+1)
		if nextIdx >= 0 && containsURL(strings.ToLower(strings.TrimSpace(rawLines[nextIdx]))) {
			return true
		}
	}
	if !hasURLLine {
		return false
	}
	tagsDoc := newSearchDocument(strings.Join(record.Tags, " "))
	return searchContainsIntentGroupToken(tagsDoc.tokens, 1)
}

func looksLikeSearchDatabaseAccessLine(text string) bool {
	text = normalizeSearchText(text)
	if text == "" {
		return false
	}
	for _, hint := range searchDatabaseAccessHints {
		if strings.Contains(text, compactSearchText(hint)) {
			return true
		}
	}
	return false
}

func stripSearchURLAffixes(token string) string {
	token = compactSearchText(token)
	if token == "" {
		return ""
	}
	best := token
	for _, term := range searchSynonymGroups[0] {
		affix := compactSearchText(term)
		if affix == "" || affix == token {
			continue
		}
		if strings.HasPrefix(token, affix) {
			candidate := strings.TrimPrefix(token, affix)
			if utf8.RuneCountInString(candidate) >= 2 && utf8.RuneCountInString(candidate) < utf8.RuneCountInString(best) {
				best = candidate
			}
		}
		if strings.HasSuffix(token, affix) {
			candidate := strings.TrimSuffix(token, affix)
			if utf8.RuneCountInString(candidate) >= 2 && utf8.RuneCountInString(candidate) < utf8.RuneCountInString(best) {
				best = candidate
			}
		}
	}
	return best
}

func searchTokenSimilarity(left, right string) float64 {
	left = compactSearchText(left)
	right = compactSearchText(right)
	if left == "" || right == "" {
		return 0
	}
	if left == right {
		return 1
	}
	leftGroup, leftOK := searchSynonymIndex[left]
	rightGroup, rightOK := searchSynonymIndex[right]
	if leftOK && rightOK && leftGroup == rightGroup {
		return 0.92
	}
	if strings.Contains(right, left) || strings.Contains(left, right) {
		shorter := minInt(utf8.RuneCountInString(left), utf8.RuneCountInString(right))
		longer := maxInt(utf8.RuneCountInString(left), utf8.RuneCountInString(right))
		if isASCIIAlphaNumSearchToken(left) && isASCIIAlphaNumSearchToken(right) {
			ratio := float64(shorter) / float64(longer)
			if ratio < 0.75 {
				return 0.18 + 0.62*ratio
			}
			return 0.42 + 0.5*ratio
		}
		return 0.82 + 0.18*float64(shorter)/float64(longer)
	}
	ngram := diceCoefficient(buildSearchNgrams(left), buildSearchNgrams(right))
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	if len(leftRunes) > 64 || len(rightRunes) > 64 {
		return ngram
	}
	distance := algow.EditDistance(leftRunes, rightRunes, nil)
	maxLen := maxInt(len(leftRunes), len(rightRunes))
	if maxLen == 0 {
		return 0
	}
	edit := 1 - float64(distance)/float64(maxLen)
	if edit < 0 {
		edit = 0
	}
	return 0.68*edit + 0.32*ngram
}

func requiresStrongSearchTokenMatch(token string) bool {
	token = compactSearchText(token)
	if token == "" {
		return false
	}
	return isASCIIAlphaNumSearchToken(token) && utf8.RuneCountInString(token) <= 2
}

func normalizeSearchText(text string) string {
	return normalizeSearchTextWithOptions(text, true)
}

func normalizeSearchTextWithOptions(text string, trimURLs bool) string {
	text = sanitizeSearchSource(text, trimURLs)
	text = strings.TrimSpace(strings.ToLower(norm.NFKC.String(text)))
	if text == "" {
		return ""
	}
	var builder strings.Builder
	lastSpace := false
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.In(r, unicode.Han) {
			builder.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			builder.WriteByte(' ')
			lastSpace = true
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func sanitizeSearchSource(text string, trimURLs bool) string {
	if !trimURLs {
		return text
	}
	return searchURLPattern.ReplaceAllString(text, " url ")
}

func compactSearchText(text string) string {
	return strings.ReplaceAll(normalizeSearchText(text), " ", "")
}

func tokenizeSearchText(text string) []string {
	normalized := normalizeSearchText(text)
	if normalized == "" {
		return nil
	}
	tokens := make([]string, 0)
	seen := map[string]struct{}{}
	appendToken := func(token string) {
		token = compactSearchText(token)
		if token == "" {
			return
		}
		if _, ignored := searchIgnoredTokens[token]; ignored {
			return
		}
		if _, ok := seen[token]; ok {
			return
		}
		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}
	for _, chunk := range strings.Fields(normalized) {
		parts := splitMixedSearchToken(chunk)
		for _, part := range parts {
			appendToken(part)
			appendKnownSearchTerms(part, appendToken)
			appendDerivedSearchTerms(part, appendToken)
		}
		appendKnownSearchTerms(chunk, appendToken)
		appendDerivedSearchTerms(chunk, appendToken)
	}
	return tokens
}

func splitMixedSearchToken(token string) []string {
	parts := make([]string, 0, 2)
	var builder strings.Builder
	lastClass := 0
	flush := func() {
		if builder.Len() == 0 {
			return
		}
		parts = append(parts, builder.String())
		builder.Reset()
	}
	for _, r := range token {
		class := searchRuneClass(r)
		if class == 0 {
			flush()
			lastClass = 0
			continue
		}
		if lastClass != 0 && class != lastClass {
			flush()
		}
		builder.WriteRune(r)
		lastClass = class
	}
	flush()
	if len(parts) == 0 && token != "" {
		parts = append(parts, token)
	}
	return parts
}

func appendKnownSearchTerms(token string, appendToken func(string)) {
	compact := compactSearchText(token)
	if compact == "" {
		return
	}
	for _, term := range searchKnownTerms {
		if term == compact {
			continue
		}
		if strings.Contains(compact, term) {
			appendToken(term)
		}
	}
}

func appendDerivedSearchTerms(token string, appendToken func(string)) {
	compact := compactSearchText(token)
	if compact == "" {
		return
	}
	stripped := stripGenericSearchSuffixes(compact)
	if stripped != "" && stripped != compact {
		appendToken(stripped)
		appendKnownSearchTerms(stripped, appendToken)
	}
	if isHanSearchToken(compact) {
		appendHanSearchPrefixes(compact, appendToken)
	}
}

func stripGenericSearchSuffixes(token string) string {
	stripped := token
	for {
		updated := stripped
		for _, suffix := range searchGenericSuffixes {
			if strings.HasSuffix(updated, suffix) {
				candidate := strings.TrimSuffix(updated, suffix)
				if utf8.RuneCountInString(candidate) >= 2 {
					updated = candidate
				}
			}
		}
		if updated == stripped {
			break
		}
		stripped = updated
	}
	return stripped
}

func appendHanSearchPrefixes(token string, appendToken func(string)) {
	runes := []rune(token)
	maxPrefix := minInt(4, len(runes)-1)
	for size := 2; size <= maxPrefix; size++ {
		appendToken(string(runes[:size]))
	}
}

func isHanSearchToken(token string) bool {
	if token == "" {
		return false
	}
	for _, r := range token {
		if !unicode.In(r, unicode.Han) {
			return false
		}
	}
	return true
}

func looksLikeExplicitURL(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	return strings.Contains(text, "http://") || strings.Contains(text, "https://")
}

func searchRequestsURL(text string) bool {
	for _, token := range tokenizeSearchText(text) {
		if group, ok := searchSynonymIndex[token]; ok && group == 0 {
			return true
		}
	}
	return false
}

func containsURL(text string) bool {
	return strings.Contains(text, "http://") || strings.Contains(text, "https://")
}

func buildSearchNgrams(text string) map[string]int {
	text = compactSearchText(text)
	if text == "" {
		return nil
	}
	runes := []rune(text)
	ngrams := map[string]int{}
	if len(runes) == 1 {
		ngrams[text] = 1
		return ngrams
	}
	for _, size := range []int{2, 3} {
		if len(runes) < size {
			continue
		}
		for i := 0; i+size <= len(runes); i++ {
			ngrams[string(runes[i:i+size])]++
		}
	}
	if len(ngrams) == 0 {
		ngrams[text] = 1
	}
	return ngrams
}

func diceCoefficient(left, right map[string]int) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	intersection := 0
	total := 0
	for _, count := range left {
		total += count
	}
	for _, count := range right {
		total += count
	}
	for token, leftCount := range left {
		rightCount, ok := right[token]
		if !ok {
			continue
		}
		intersection += minInt(leftCount, rightCount)
	}
	if total == 0 {
		return 0
	}
	return 2 * float64(intersection) / float64(total)
}

func searchRuneClass(r rune) int {
	switch {
	case unicode.In(r, unicode.Han):
		return 1
	case r < utf8.RuneSelf && (unicode.IsLetter(r) || unicode.IsNumber(r)):
		return 2
	case unicode.IsLetter(r) || unicode.IsNumber(r):
		return 3
	default:
		return 0
	}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func clampInt(value, lower, upper int) int {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}

func compareSearchFloat(left, right float64) int {
	if math.IsNaN(left) {
		if math.IsNaN(right) {
			return 0
		}
		return -1
	}
	if math.IsNaN(right) {
		return 1
	}
	diff := left - right
	if math.Abs(diff) <= searchFloatCompareEpsilon {
		return 0
	}
	if diff < 0 {
		return -1
	}
	return 1
}

func compareSearchFloatDesc(left, right float64) int {
	return -compareSearchFloat(left, right)
}

func minFloat(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	minVal := values[0]
	for _, value := range values[1:] {
		if value < minVal {
			minVal = value
		}
	}
	return minVal
}

func maxFloat(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, value := range values[1:] {
		if value > maxVal {
			maxVal = value
		}
	}
	return maxVal
}
