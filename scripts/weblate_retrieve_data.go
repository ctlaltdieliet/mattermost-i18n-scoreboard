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

var weblateToken string = ""

type translator struct {
	FullName   string
	Username   string
	DateJoined string
	Translated int
	Suggested  int
	Commented  int
	Total      int
	Languages  int
}

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
			fmt.Println("weblate didn't respond 200")
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
			translators = append(translators, translator{
				FullName:   weblateusers[k].FullName,
				Username:   weblateusers[k].Username,
				DateJoined: weblateusers[k].DateJoined,
				Translated: res.Translated,
				Commented:  res.Commented,
				Suggested:  res.Suggested,
				Total:      res.Suggested + res.Commented + res.Translated,
				Languages:  res.Languages,
			})
		}
	}
	return translators
}
func writeToFile(translators []translator) {
	var jsondata, err = json.Marshal(translators)
	if err != nil {
		fmt.Println("Could not marshal translators")
		log.Fatal(err)
	}
	year, month, day := time.Now().Date()
	if day == 0 {
		fmt.Println("d")
	}
	folder := fmt.Sprintf("/home/tomdemoor/mattermost/i18n/scripts/mattermost-i18n-scoreboard/data/%d/%d/", year, month)
	os.MkdirAll(folder, os.ModePerm)
	os.WriteFile(folder+fmt.Sprint(day)+".json", jsondata, 0644)

}

func main() {
	//FETCHING ALL USERS FROM WEBLATE AND STORING THEM IN weblateusers
	var weblateusers []generalUserData = fetchAllUsers("https://translate.mattermost.com/api/users/")

	// FETCHING STAT FOR EACH USER AND STORING ALL USERS AND DATA in translators
	var translators []translator = fetchTranslationsByUser(weblateusers)

	//WRITING STATS TO JSON-FILE
	writeToFile(translators)
}
