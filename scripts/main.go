package main

const (
	layout              = "2006-1-2"
	pathScript          = "/home/tom/mattermost-i18n-scoreboard/"
	weblateToken string = ""
)

func main() {
	fetchTranslators()
	fetchTranslations()

}
