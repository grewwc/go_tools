package internal

import "testing"

func disableSearchLLMProvider(t *testing.T) {
	t.Helper()
	withFakeSearchLLMProvider(t, &fakeSearchLLMProvider{enabled: false})
}

func disableRemoteSearchProviders(t *testing.T) {
	t.Helper()
	withFakeSearchEmbeddingProvider(t, &fakeSearchEmbeddingProvider{enabled: false})
	disableSearchLLMProvider(t)
}
