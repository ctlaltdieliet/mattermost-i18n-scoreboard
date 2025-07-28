package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"

	"github.com/golang-module/carbon/v2"
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
	Language   string
}

func createStatsTranslators(fromdate string, tilldate string) []translator {

	var translators []translator
	fromDate, errFromDate := time.Parse(layout, fromdate)
	tillDate, errTillDate := time.Parse(layout, tilldate)
	if errFromDate == nil && errTillDate == nil {
		var pathFromFile string = fmt.Sprintf("%s/data/translators/%d/%d/", pathScript, fromDate.Year(), fromDate.Month())
		var pathTillFile string = fmt.Sprintf("%s/data/translators/%d/%d/", pathScript, tillDate.Year(), tillDate.Month())
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
								Language:      translatorsTill[vT].Language,
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
							Language:      translatorsTill[vT].Language,
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
	fmt.Println(translators)
	return translators
}

func createPageTranslators(title string, page string, Sort string, fromDate string, tillDate string, limit int, descending bool) {
	var output string = "## " + title + " ##\n"
	if limit == 0 {
		limit = 2000000
	}
	var translators []translator = createStatsTranslators(fromDate, tillDate)
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

	output = output + "|Username|Fullname|Translated|DateJoined|Language|\n"
	output = output + "|--------|--------|----------|----------|-------|\n"

	for i, translator := range translators {
		if i < limit {
			output = output + fmt.Sprintf("|%s|%s|%d|%s|%s|\n", translator.Username, translator.FullName, translator.Translated, translator.DateJoined[0:20], translator.Language)
		}
	}

	var file string = pathScript + "/pages/" + page

	os.WriteFile(file, []byte(output), 0644)
	now := carbon.Now()
	var archivefile string = fmt.Sprint(pathScript + "/archive/" + now.ToDateString() + " - " + page)
	os.WriteFile(archivefile, []byte(output), 0644)
}

func fetchTranslators() {
	//FETCHING ALL USERS FROM WEBLATE AND STORING THEM IN weblateusers
	var weblateusers []generalUserData = fetchAllUsers("https://translate.mattermost.com/api/users/")

	// FETCHING STAT FOR EACH USER AND STORING ALL USERS AND DATA in translators
	var translators []translator = fetchTranslationsByUser(weblateusers)

	//WRITING STATS TO JSON-FILE
	writeToFileTranslators(translators)
	now := carbon.Now()
	StartOfCurrentWeek := now.SetWeekStartsAt(carbon.Monday).StartOfWeek()
	StartOfCurrentMonth := now.StartOfMonth()
	createPageTranslators("Top 20 Contributors Week Till Today", "weekly_top_contributors_till_today.md", "Translated", StartOfCurrentWeek.ToDateString(), now.ToDateString(), 20, true)
	createPageTranslators("Top 20 Contributors From Beginning Month Till Today", "monthly_top_contributors_till_today.md", "Translated", StartOfCurrentMonth.ToDateString(), now.ToDateString(), 20, true)

	if now.DayOfWeek() == 7 {
		//IT'S SUNDAY, CREATE WEEKLY STATS
		createPageTranslators("Top 20 Contributors Current Week", "weekly_top_contributors.md", "Translated", StartOfCurrentWeek.ToDateString(), now.ToDateString(), 20, true)
		createPageTranslators("Top 20 Contributors From Beginning Month Till Today", "current_month_top_contributors.md", "Translated", StartOfCurrentMonth.ToDateString(), now.ToDateString(), 20, true)
	}
	if now.DayOfMonth() == 1 {
		// IT'S THE BEGINNING OF THE MONTH, CREATE MONTHLY STATS
		EndOfPrevMonth := now.StartOfMonth().Yesterday()
		StartOfPrevMonth := EndOfPrevMonth.StartOfMonth()
		createPageTranslators("Top 20 Contributors Previous Month", "previous_month_top_contributors.md", "Translated", StartOfPrevMonth.ToDateString(), EndOfPrevMonth.ToDateString(), 20, true)
		createPageTranslators("Contributors YEAR TILL TODAY", "year_till_today_contributors.md", "Translated", now.StartOfYear().ToDateString(), now.ToDateString(), 0, true)

		createPageTranslators("Translators By Date Joined", "translators_by_date_joined.md", "DateJoined", StartOfPrevMonth.ToDateString(), EndOfPrevMonth.ToDateString(), 0, true)
		if now.MonthOfYear() == 1 || now.MonthOfYear() == 4 || now.MonthOfYear() == 8 || now.MonthOfYear() == 20 {
			//IT'S THE START OF A NEW QUARTER, CREATE QUARTERLY STATS
			EndOfPrevQuarter := now.StartOfQuarter().Yesterday()
			StartOfPrevQuarter := EndOfPrevQuarter.StartOfQuarter()
			createPageTranslators("Top  Contributors Previous Quarter", "previous_quarter_top_contributors.md", "Translated", StartOfPrevQuarter.ToDateString(), EndOfPrevQuarter.ToDateString(), 0, true)
		}
		if now.MonthOfYear() == 1 {
			// IT'S JANUARY 1ST, HAPPY NEW YEAR AND CREATE THE YEARLY STATS.
			EndOfPrevYear := now.StartOfYear().Yesterday()
			StartOfPrevYear := EndOfPrevYear.StartOfYear()
			createPageTranslators("Top Contributors Previous Year", "previous_year_top_contributors.md", "Translated", StartOfPrevYear.ToDateString(), EndOfPrevYear.ToDateString(), 0, true)
		}
	}
}

