package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"strconv"
	"time"
)

type userStatsWeblate struct {
	Translated int `json:"translated"`
	Suggested  int `json:"suggested"`
	Commented  int `json:"commented"`
	Languages  int `json:"languages"`
}

type generalUserData struct {
	FullName      string `json:"full_name"`
	Username      string `json:"username"`
	StatisticsURL string `json:"statistics_url"`
	DateJoined    string `json:"date_joined"`
}
type usersOverview struct {
	Count   int    `json:"count"`
	Next    string `json:"next"`
	Results []generalUserData
}

type userChangesResults struct {
	Translation string `json:"translation"`
}
type userChanges struct {
	Results []userChangesResults
}

func fetchAllUsers(urlRequest string) []generalUserData {
	var hasMorePages bool = true
	var weblateusers []generalUserData
	//var  string = ""
	for hasMorePages == true {
		weblateClient := http.Client{
			Timeout: time.Second * 2, // Timeout after 2 seconds
		}

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
			fmt.Println("weblate didn't respond 200 -1  but"+strconv.Itoa(res.StatusCode))
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
			var res usersOverview
			json.Unmarshal(bodyBytes, &res)
			for k := range res.Results {
				if res.Results[k].FullName != "Deleted User" {
					weblateusers = append(weblateusers, res.Results[k])
				}
			}
			if string(res.Next) != "" {
				urlRequest = res.Next
			} else {
				hasMorePages = false
			}
		}
	}
	return weblateusers
}

func fetchTranslationsByUser(weblateusers []generalUserData) []translator {
	var translators []translator
	var language string = ""
	var urlStats string = ""
	for k := range weblateusers {
		urlStats = weblateusers[k].StatisticsURL

		weblateClient := http.Client{
			Timeout: time.Second * 2, // Timeout after 2 seconds
		}

		req, err := http.NewRequest(http.MethodGet, urlStats, nil)
		if err != nil {
			fmt.Println("Could not make request")
			log.Fatal(err)
		}

		req.Header.Set("Authorization", "Token "+weblateToken)

		res, getErr := weblateClient.Do(req)
		if res.StatusCode != 200 {
			fmt.Println("weblate didn't respond 200 -2 but "+strconv.Itoa(res.StatusCode))
			log.Fatal(getErr)
		}
		if getErr != nil {
			fmt.Println("Failure weblateClient")
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
			res := new(userStatsWeblate) // new() returns a pointer to an initialized struct
			err = json.Unmarshal(bodyBytes, res)
			if err != nil {
				log.Fatal(err)
			}
			if (res.Suggested + res.Commented + res.Translated) > 0 {
				urlLanguage := "https://translate.mattermost.com/api/changes/?action=5&user=" + weblateusers[k].Username
				weblateClient := http.Client{
					Timeout: time.Second * 2, // Timeout after 2 seconds
				}

				reqLang, err := http.NewRequest(http.MethodGet, urlLanguage, nil)
				if err != nil {
					fmt.Println("Could not make request")
					log.Fatal(err)
				}

				reqLang.Header.Set("Authorization", "Token "+weblateToken)

				resLang, getErr := weblateClient.Do(reqLang)
				if resLang.StatusCode != 200 {
					fmt.Println("weblate didn't respond 200 -3 but "+strconv.Itoa(resLang.StatusCode))
					log.Fatal(getErr)
				}
				if getErr != nil {
					fmt.Println("Failure weblateClient")
					log.Fatal(getErr)
				}
				if resLang.Body != nil {
					defer resLang.Body.Close()
				}

				bodyBytes, readErr := ioutil.ReadAll(resLang.Body)
				if readErr != nil {
					fmt.Println("failure reading body")
					log.Fatal(readErr)
				} else {
					var resLang userChanges
					json.Unmarshal(bodyBytes, &resLang)
					for l := range resLang.Results {
						s := strings.Split(resLang.Results[l].Translation, "/")
						language = s[len(s)-2]
					}
				}
			}
			

			translators = append(translators, translator{
				FullName:   weblateusers[k].FullName,
				Username:   weblateusers[k].Username,
				DateJoined: weblateusers[k].DateJoined,
				Translated: res.Translated,
				Commented:  res.Commented,
				Suggested:  res.Suggested,
				Total:      res.Suggested + res.Commented + res.Translated,
				Languages:  res.Languages,
				Language:   language,
			})
		}
	}
	return translators
}
func writeToFileTranslators(translators []translator) {
	var jsondata, err = json.Marshal(translators)
	if err != nil {
		fmt.Println("Could not marshal translators")
		log.Fatal(err)
	}
	year, month, day := time.Now().Date()
	folder := fmt.Sprintf("%s/data/translators/%d/%d/", pathScript, year, month)
	os.MkdirAll(folder, os.ModePerm)
	os.WriteFile(folder+fmt.Sprint(day)+".json", jsondata, 0644)
}

