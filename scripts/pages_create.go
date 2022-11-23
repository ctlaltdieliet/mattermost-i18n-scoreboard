package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"
)

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

const (
	layout = "2006-01-02"
)

func createStats(fromdate string, tilldate string) []translator {

	var translators []translator
	fromDate, errFromDate := time.Parse(layout, fromdate)
	tillDate, errTillDate := time.Parse(layout, tilldate)
	if errFromDate == nil && errTillDate == nil {
		var pathFromFile string = fmt.Sprintf("/home/tomdemoor/mattermost/i18n/scripts/mattermost-i18n-scoreboard/data/%d/%d/", fromDate.Year(), fromDate.Month())
		var pathTillFile string = fmt.Sprintf("/home/tomdemoor/mattermost/i18n/scripts/mattermost-i18n-scoreboard/data/%d/%d/", tillDate.Year(), tillDate.Month())
		fromJson, errFromFile := ioutil.ReadFile(pathFromFile + fmt.Sprint(fromDate.Day()) + ".json")
		tillJson, errTillFile := ioutil.ReadFile(pathTillFile + fmt.Sprint(tillDate.Day()) + ".json")
		if errFromFile == nil && errTillFile == nil {

			var translatorsFrom []translator
			var translatorsTill []translator
			errFrom := json.Unmarshal([]byte(fromJson), &translatorsFrom)
			errTill := json.Unmarshal([]byte(tillJson), &translatorsTill)

			if errFrom != nil && errTill != nil {
				log.Fatal(errFrom, errTill)
			} else {
				/* from parsing json*/
				for vT := range translatorsTill {
					var found bool = false

					for vF := range translatorsFrom {
						if translatorsTill[vT].Username == translatorsFrom[vF].Username {
							translators = append(translators, translator{
								FullName:   translatorsTill[vT].FullName,
								Username:   translatorsTill[vT].Username,
								DateJoined: translatorsTill[vT].DateJoined,
								Translated: translatorsTill[vT].Translated - translatorsFrom[vF].Translated,
								Commented:  translatorsTill[vT].Commented - translatorsFrom[vF].Commented,
								Suggested:  translatorsTill[vT].Suggested - translatorsFrom[vF].Suggested,
								Total:      translatorsTill[vT].Total - translatorsFrom[vF].Total,
							})
							found = true
							continue
						}

					}
					if !found {
						translators = append(translators, translator{
							FullName:   translatorsTill[vT].FullName,
							Username:   translatorsTill[vT].Username,
							DateJoined: translatorsTill[vT].DateJoined,
							Translated: translatorsTill[vT].Translated,
							Commented:  translatorsTill[vT].Commented,
							Suggested:  translatorsTill[vT].Suggested,
							Total:      translatorsTill[vT].Total,
						})
					}
				}
			}
		} else {
			log.Fatalf(fmt.Sprintf("Error reading file  %d or %d %d or %d", errFromFile, errTillFile, fromJson, tillJson))
		}
	} else {
		log.Fatal("Error parsing one of the dates " + tilldate + " " + tilldate)
	}
	return translators
}

func createPage(title string, page string, Sort string, fromDate string, tillDate string, limit int, descending bool) {
	var output string = "## title ##\n"
	var translators []translator = createStats(fromDate, tillDate)
	sort.Slice(translators, func(i, j int) bool {
		switch {
		case Sort == "Total":
			if descending {
				return translators[i].Total > translators[j].Total
			} else {
				return translators[i].Total < translators[j].Total
			}
		case Sort == "Username":
			if descending {
				return translators[i].Username > translators[j].Username
			} else {
				return translators[i].Username < translators[j].Username
			}
		case Sort == "DateJoined":
			if descending {
				return translators[i].DateJoined > translators[j].DateJoined
			} else {
				return translators[i].DateJoined < translators[j].DateJoined
			}
		case Sort == "FullName":
			if descending {
				return translators[i].FullName > translators[j].FullName
			} else {
				return translators[i].FullName < translators[j].FullName
			}
		case Sort == "Translated":
			if descending {
				return translators[i].Translated > translators[j].Translated
			} else {
				return translators[i].Translated < translators[j].Translated
			}
		default:
			if descending {
				return translators[i].Translated > translators[j].Translated
			} else {
				return translators[i].Translated < translators[j].Translated
			}
		}
	})

	output = output + "|Username|Fullname|Translated|DateJoined|\n"
	output = output + "|--------|--------|----------|----------|\n"

	for i, translator := range translators {
		if i < limit {
			output = output + fmt.Sprintf("|%s|%s|%d|%s|\n", translator.Username, translator.FullName, translator.Translated, translator.DateJoined[0:10])
		}
	}
	var file string = "/home/tomdemoor/mattermost/i18n/scripts/mattermost-i18n-scoreboard/pages/" + page
	os.WriteFile(file, []byte(output), 0644)

}

func main() {
	var today time.Time = time.Now()
	var currentMonth string = fmt.Sprintf("%d", today.Month())
	var currentDay string = fmt.Sprintf("%d", today.Day())
	var currentYear string = fmt.Sprintf("%d", today.Year())
	var todayString = fmt.Sprintf("%s-%s-%s", currentYear, currentMonth, currentDay)
	//var startCurrentMonth string = fmt.Sprintf("%s-%s-01", currentYear, currentMonth)
	//var startPreviousMonth string = fmt.Sprintf("%s-%s-01", currentYear, currentMonth-1)
	//var EndCurrentMonth string = fmt.Sprintf("%s-%s-01", currentYear, currentMonth)
	//fmt.Println

	createPage("Top 10 contributors this week", "weekly.md", "Translations", "2022-11-01", todayString, 10, true)
	createPage("Translators sorted by date joined", "new_translators.md", "Translations", "2022-11-01", todayString, 0, true)
}
