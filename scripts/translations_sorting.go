package main

import (
	"sort"
)

type sortedLanguageStats struct {
	languageCode string
	stats        languageStats
}

func sortingByTranslated(m map[string]languageStats) []sortedLanguageStats {
	sortedLanguages := make([]sortedLanguageStats, 0, len(m))
	for k, v := range m {
		sortedLanguages = append(sortedLanguages, sortedLanguageStats{k, v})
	}
	sort.SliceStable(sortedLanguages, func(i, j int) bool {
		return sortedLanguages[i].stats.TranslatedPercent > sortedLanguages[j].stats.TranslatedPercent
	})
	return sortedLanguages
}

func sortingByLastModified(m map[string]languageStats) []sortedLanguageStats {
	sortedLanguages := make([]sortedLanguageStats, 0, len(m))
	for k, v := range m {
		sortedLanguages = append(sortedLanguages, sortedLanguageStats{k, v})
	}
	sort.SliceStable(sortedLanguages, func(i, j int) bool {
		return sortedLanguages[i].stats.LastChange > sortedLanguages[j].stats.LastChange
	})

	return sortedLanguages
}
