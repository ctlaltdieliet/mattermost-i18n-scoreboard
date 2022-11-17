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

var translators []translator
var weblateusers []generalUserData
var url string = "https://translate.mattermost.com/api/users/"
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

func fetchAllUsers(urlRequest string) {
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
			url = res.Next
		} else {
			url = "STOP"
		}
	}
}

func fetchTranslationsByUser() {
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
			var res userStatsWeblate
			json.Unmarshal(bodyBytes, &res)
			var translator translator
			translator.FullName = weblateusers[k].FullName
			translator.Username = weblateusers[k].Username
			translator.DateJoined = weblateusers[k].DateJoined
			translator.Translated = res.Translated
			translator.Commented = res.Commented
			translator.Suggested = res.Suggested
			translator.Total = res.Suggested + res.Commented + res.Translated
			translator.Languages = res.Languages
			translators = append(translators, translator)
		}
	}
}
func writeToFile() {
	var jsondata, err = json.Marshal(translators)
	if err != nil {
		fmt.Println("Could not marshal translators")
		log.Fatal(err)
	}
	year, month, day := time.Now().Date()
	if day == 0 {
		fmt.Println("d")
	}
	var folder string = "/home/tomdemoor/mattermost/i18n/scripts/mattermost-i18n-scoreboard/data/" + fmt.Sprint(year) + "/" + fmt.Sprint(month) + "/"
	fmt.Println(folder)
	os.MkdirAll(folder, os.ModePerm)
	os.WriteFile(folder+fmt.Sprint(day)+".json", jsondata, 0644)

}

func main() {

	//FETCHING ALL USERS FROM WEBLATE AND STORING THEM IN weblateusers
	for url != "STOP" {
		fetchAllUsers(url)
	}
	// FETCHING STAT FOR EACH USER AND STORING ALL USERS AND DATA in translators
	fetchTranslationsByUser()

	//WRITING STATS TO JSON-FILE
	writeToFile()
}
