package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type componentStatsWeblate struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []struct {
		Language struct {
			Code       string `json:"code"`
			Name       string `json:"name"`
			Population int64  `json:"population"`
		} `json:"language"`
		Component struct {
			Name    string `json:"name"`
			Slug    string `json:"slug"`
			Project struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"project"`
		} `json:"component"`
		Total             int     `json:"total"`
		TotalWords        int     `json:"total_words"`
		Translated        int     `json:"translated"`
		TranslatedWords   int     `json:"translated_words"`
		TranslatedPercent float64 `json:"translated_percent"`
	} `json:"results"`
}

type languageStatsWeblate struct {
	LastChange             string  `json:"last_change"`
	RecentChanges          int     `json:"recent_changes"`
	TranslatedPercent      float64 `json:"translated_percent"`
	TranslatedWordsPercent float64 `json:"translated_words_percent"`
	TranslatedCharsPercent float64 `json:"translated_chars_percent"`
}

type Component struct {
	ComponentName     string
	ProjectName       string
	Total             int
	TotalWords        int
	Translated        int
	TranslatedWords   int
	TranslatedPercent float64
}

type languageStats struct {
	Name                   string
	LastChange             string
	RecentChanges          int
	TranslatedPercent      float64
	TranslatedWordsPercent float64
	TranslatedCharsPercent float64
	Components             []Component
}

func readWeblateStatsTranslations(urlRequest string) map[string]languageStats {
	languageStatistics := make(map[string]languageStats)
	weblateClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}
	var hasMorePages bool = true
	//	var weblateusers []generalUserData
	for hasMorePages == true {

		req, err := http.NewRequest(http.MethodGet, urlRequest, nil)
		if err != nil {
			fmt.Println("Could not make request")
			log.Fatal(err)
		}
		req.Header.Set("Authorization", "Token "+weblateToken)

		res, getErr := weblateClient.Do(req)
		if getErr != nil {
			fmt.Println("Failure weblateClient")
			log.Fatal(getErr)
		}
		if res.StatusCode != 200 {
			fmt.Println("weblate didn't respond 200")
			log.Fatal(getErr)
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		bodyBytes, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			fmt.Println("failure reading body")
			log.Fatal(readErr)

		} else {
			jsonData := bodyBytes
			var data componentStatsWeblate
			err = json.Unmarshal([]byte(jsonData), &data)
			if err != nil {
				fmt.Printf("could not unmarshal json: %s\n", err)
			} else {
				for _, k := range data.Results {
					if k.Language.Code != "en" {
						var tempStat languageStats
						if _, ok := languageStatistics[k.Language.Code]; !ok {
							tempStat.Name = k.Language.Name

							var languagestat languageStatsWeblate
							req, err := http.NewRequest(http.MethodGet, "https://translate.mattermost.com/api/languages/"+k.Language.Code+"/statistics/", nil)
							if err != nil {
								fmt.Println("Could not make request")
								log.Fatal(err)
							}
							req.Header.Set("Authorization", "Token "+weblateToken)
							res, getErr := weblateClient.Do(req)
							if getErr != nil {
								fmt.Println("Failure weblateClient")
								log.Fatal(getErr)
							}
							if res.StatusCode != 200 {
								fmt.Println("weblate didn't respond 200")
								fmt.Println(res.StatusCode)
								log.Fatal(getErr)
							}
							if res.Body != nil {
								defer res.Body.Close()
							}
							bodyBytesStat, readErr := ioutil.ReadAll(res.Body)
							if readErr != nil {
								fmt.Println("failure reading body")
								log.Fatal(readErr)
							} else {
								jsonDataStat := bodyBytesStat
								err = json.Unmarshal([]byte(jsonDataStat), &languagestat)
								if err != nil {
									fmt.Printf("could not unmarshal json: %s\n", err)
								} else {
									tempStat.LastChange = languagestat.LastChange
									tempStat.RecentChanges = languagestat.RecentChanges
									tempStat.TranslatedPercent = languagestat.TranslatedPercent
									tempStat.TranslatedWordsPercent = languagestat.TranslatedWordsPercent
									tempStat.TranslatedCharsPercent = languagestat.TranslatedCharsPercent

								}
							}
							languageStatistics[k.Language.Code] = tempStat
						}
						tempStat = languageStatistics[k.Language.Code]
						var tempComponent Component
						if k.Component.Project.Name == "Mattermost Plugins" {
							continue
						}
						tempComponent.ComponentName = k.Component.Name
						tempComponent.ProjectName = k.Component.Project.Name
						if k.Component.Project.Name == "Focalboard" && k.Component.Name == "webapp" {
							k.Component.Name = "focalboard-webapp"
						}
						if k.Component.Project.Name == "Playbooks" && k.Component.Name == "webapp" {
							k.Component.Name = "playbooks-webapp"
						}

						if k.Component.Project.Name == "Calls" && k.Component.Name == "webapp" {
							k.Component.Name = "calls-webapp"
						}
						tempComponent.ComponentName = k.Component.Name
						tempComponent.Total = k.Total
						tempComponent.TotalWords = k.TotalWords
						tempComponent.Translated = k.Translated
						tempComponent.TranslatedWords = k.TranslatedWords
						tempComponent.TranslatedPercent = k.TranslatedPercent
						tempStat.Components = append(tempStat.Components, tempComponent)
						languageStatistics[k.Language.Code] = tempStat
					}
				}
			}
			if string(data.Next) != "" {
				urlRequest = data.Next
			} else {
				hasMorePages = false
			}
		}
	}
	return languageStatistics
}

func fetchTranslations() {
	var statsTranslations map[string]languageStats = readWeblateStatsTranslations("https://translate.mattermost.com/api/translations/")
	writeToFileTranslations(statsTranslations)
	year, month, day := time.Now().Date()
	fromDate := fmt.Sprintf("%d-%d-%d", year, month, day)
	createPageTranslations("Current state of translations", "percentage", fromDate, fromDate)
}

func writeToFileTranslations(stats map[string]languageStats) {
	var jsondata, err = json.Marshal(stats)
	if err != nil {
		fmt.Println("Could not marshal translators")
		log.Fatal(err)
	}
	year, month, day := time.Now().Date()
	folder := fmt.Sprintf("%s/data/translations/%d/%d/", pathScript, year, month)
	os.MkdirAll(folder, os.ModePerm)
	os.WriteFile(folder+fmt.Sprint(day)+".json", jsondata, 0644)
}
