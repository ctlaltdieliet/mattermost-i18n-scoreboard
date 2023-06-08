package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/golang-module/carbon/v2"
)

func getComponents() ([]string, []string) {
	componentsShipped := []string{"mattermost-webapp-monorepo", "mattermost-server-monorepo", "mattermost-mobile-v2", "mattermost-desktop", "focalboard-webapp", "playbooks-webapp", "calls-webapp"}
	componentsWIP := []string{"mattermost-webapp-wip", "mattermost-server-wip", "mattermost-webapp-wip"}
	return componentsShipped, componentsWIP
}
func splitShippedWippedLanguages(languageStatistics map[string]languageStats) (map[string]languageStats, map[string]languageStats) {
	var shippedLanguages = make(map[string]languageStats)
	var WIPLanguages = make(map[string]languageStats)
	for language, stats := range languageStatistics {
		for _, component := range stats.Components {
			if component.ComponentName == "mattermost-webapp-monorepo" {
				shippedLanguages[language] = stats
			}
			if component.ComponentName == "mattermost-webapp-wip" {
				WIPLanguages[language] = stats
			}
		}
	}
	return shippedLanguages, WIPLanguages
}

func readJSONLanguages(startdate string) map[string]languageStats {
	fromDate, errFromDate := time.Parse(layout, startdate)
	if errFromDate == nil {
		var path string = fmt.Sprintf("%s/data/translations/%d/%d/", pathScript, fromDate.Year(), fromDate.Month())
		//fmt.Println("HET PAD IS " + path + fmt.Sprint(fromDate.Day()) + ".json")
		jsonFile, err := ioutil.ReadFile(path + fmt.Sprint(fromDate.Day()) + ".json")
		if err != nil {
			log.Fatal("error reading json languages file " + path + fmt.Sprint(fromDate.Day()) + ".json")
		} else {
			var languageStats map[string]languageStats
			errUnmarshal := json.Unmarshal([]byte(jsonFile), &languageStats)
			if errUnmarshal != nil {
				log.Fatal("error unmarshalling json languages file ")
			} else {
				return languageStats
			}
		}
	} else {
		fmt.Println(errFromDate)
	}
	return nil
}

func createPageTranslations(page string, sort string, fromDate string, tillDate string) {
	var sortedLanguageStatsShippedStart []sortedLanguageStats
	var sortedLanguageStatsWIPStart []sortedLanguageStats
	var outputShippedTitles string = "|---|---|"
	var outputWIPTitles string = "|---|---|"
	var outputShipped string = "###  Shipped languages  ###\n"
	var outputWIP string = "###  WIP languages  ###\n"

	componentsShipped := []string{"mattermost-webapp-monorepo", "mattermost-server-monorepo", "mattermost-mobile-v2", "mattermost-desktop", "focalboard-webapp", "playbooks-webapp","calls-webapp"}
	componentsWIP := []string{"mattermost-webapp-wip", "mattermost-server-wip", "mattermost-webapp-wip"}

	outputShipped = outputShipped + "|Language|Code|"
	for _, component := range componentsShipped {
		outputShipped = outputShipped + component + "|"
		outputShippedTitles = outputShippedTitles + "---|"
	}
	outputShippedTitles = outputShippedTitles + "---|---|\n"
	outputShipped = outputShipped + "Total|Last Modified|\n" + outputShippedTitles

	outputWIP = outputWIP + "|Language|Code|"
	for _, component := range componentsWIP {
		outputWIP = outputWIP + component + "|"
		outputWIPTitles = outputWIPTitles + "---|"
	}

	outputWIPTitles = outputWIPTitles + "---|--|\n"
	outputWIP = outputWIP + "Total|Last Modified|\n" + outputWIPTitles

	var translationsStart map[string]languageStats = readJSONLanguages(fromDate)
	/*if tillDate != fromDate {
		var translationsEnd map[string]languageStats = readJSONLanguages(tillDate)
	}*/

	languagesShippedStart, languagesWIPStart := splitShippedWippedLanguages(translationsStart)

	if sort == "percentage" {
		sortedLanguageStatsShippedStart = sortingByTranslated((languagesShippedStart))
		sortedLanguageStatsWIPStart = sortingByTranslated((languagesWIPStart))
	}
	if sort == "lastmodified" {
		sortedLanguageStatsShippedStart = sortingByLastModified((languagesShippedStart))
		sortedLanguageStatsWIPStart = sortingByLastModified((languagesWIPStart))
	}
	for _, stats := range sortedLanguageStatsShippedStart {
		outputShipped = outputShipped + "|" + stats.stats.Name + "|" + stats.languageCode + "|"
		for _, componentName := range componentsShipped {
			var TranslatedPercent int64 = 0
			for _, component := range stats.stats.Components {
				if component.ComponentName == componentName {
					TranslatedPercent = int64(component.TranslatedPercent)
				}
			}
			outputShipped = fmt.Sprintf("%s %d|", outputShipped, TranslatedPercent)
		}
		outputShipped = fmt.Sprintf("%s %d|%s|\n", outputShipped, int64(stats.stats.TranslatedPercent), stats.stats.LastChange)
	}
	for _, stats := range sortedLanguageStatsWIPStart {
		outputWIP = outputWIP + "|" + stats.stats.Name + "|" + stats.languageCode + "|"
		for _, componentName := range componentsWIP {
			var TranslatedPercent int64 = 0
			for _, component := range stats.stats.Components {
				if component.ComponentName == componentName {
					TranslatedPercent = int64(component.TranslatedPercent)
				}
			}
			outputWIP = fmt.Sprintf("%s %d|", outputWIP, TranslatedPercent)
		}
		outputWIP = fmt.Sprintf("%s %d|%s|\n", outputWIP, int64(stats.stats.TranslatedPercent), stats.stats.LastChange)
	}
	output := "## Statistics Languages ##\n" + outputShipped + outputWIP

	var file string = pathScript + "/pages/language_statistics.md"
	os.WriteFile(file, []byte(output), 0644)
	now := carbon.Now()
	var archivefile string = fmt.Sprint(pathScript + "/archive/" + now.ToDateString() + " - language_statistics.md")
	os.WriteFile(archivefile, []byte(output), 0644)
}
