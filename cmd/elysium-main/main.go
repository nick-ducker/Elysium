package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/robfig/cron"
)

var day int
var hour int

type Email struct {
	Sender      Sender `json:"sender"`
	To          []To   `json:"to"`
	Subject     string `json:"subject"`
	HTMLContent string `json:"htmlContent"`
}
type Sender struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type To struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func StartCrons() {
	localTimezone := "Australia/Adelaide"
	timezone, err := time.LoadLocation(localTimezone)
	if err != nil {
		panic(err)
	}
	weeklyJob := cron.NewWithLocation(timezone)
	weeklyJob.AddFunc("0 0 0 * * 0", func() {
		addJobForWeek(
			randomNumExcludeParameter(day, 1, 5),
			randomNumExcludeParameter(hour, 9, 17),
			timezone,
		)
		weeklyJob.Stop()
	})

	weeklyJob.Start()
}

func addJobForWeek(day int, hour int, timezone *time.Location) {
	cronString := fmt.Sprintf("0 0 %d * * %d", hour, day)
	fmt.Println(cronString)
	oneOffJob := cron.NewWithLocation(timezone)
	oneOffJob.AddFunc(cronString, func() {

		fmt.Println("Ding I did a thing!")

		// Get all contacts
		contactSlice := getContacts()

		// Get image from bucket
		imageUrl := getRandomImage()

		// Get poem from somewhere
		poem := getRandomPoem()
		// Contruct HTML
		html := constructHtml(imageUrl, poem)

		// dat, err := os.ReadFile("./html/email_template.html")
		// if err != nil {
		// 	panic(err)
		// }

		// html := strings.Replace(string(dat), HTTPREPLACE, )

		// Search and replace TEXTREPLACE and HTTPREPLACE

		// Construct transactional query
		// emailJson := &Email{
		// 	Sender: Sender{
		// 		Name: "Elly",
		// 		Email: "elly@thecaninecosmos.dev",
		// 	},
		// 	To: []To{nick},
		// 	Subject: "Woof",
		// 	HTMLContent:
		// }

		oneOffJob.Stop()
	})
}

func getEmailTemplate() string {
	dat, err := os.ReadFile("./html/email_template.html")
	if err != nil {
		panic(err)
	}

	html := string(dat)
	return html
}

func randomNumExcludeParameter(
	exclude int,
	min int,
	max int,
) int {
	rand.Seed(time.Now().UnixNano())

	unique := false
	var newNum int
	for !unique {
		newNum = randInt(min, max)
		fmt.Println(newNum)
		if newNum != exclude {
			unique = true
		}
	}
	return newNum
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func main() {
	fmt.Println("Starting crons")
	StartCrons()
	for {

	}
}
