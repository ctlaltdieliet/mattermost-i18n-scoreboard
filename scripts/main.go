package main

const (
	layout              = "2006-1-2"
	pathScript          = "/home/tom/mattermost-i18n-scoreboard/"
	weblateToken string = "jn8fCVl8oSRZ9mei7iUopftK4g3s2uZAIccL66PC"
)

func main() {
	fetchTranslators()
	fetchTranslations()

}
