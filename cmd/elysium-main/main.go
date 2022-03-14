package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
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
type ContactList struct {
	Contacts []Contacts `json:"contacts"`
	Count    int        `json:"count"`
}
type Attributes struct {
}
type Contacts struct {
	Email            string     `json:"email"`
	ID               int        `json:"id"`
	EmailBlacklisted bool       `json:"emailBlacklisted"`
	SmsBlacklisted   bool       `json:"smsBlacklisted"`
	CreatedAt        time.Time  `json:"createdAt"`
	ModifiedAt       time.Time  `json:"modifiedAt"`
	ListIds          []int      `json:"listIds"`
	Attributes       Attributes `json:"attributes"`
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
		contactsRaw := getContacts()

		// Process contacts
		contactsSlice := processContacts(contactsRaw)

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

func processContacts(rawContacts ContactList) []To {
	var slice []To
	for _, contact := range rawContacts.Contacts {
		contactStruct := To{
			Name:  strings.Split(contact.Email, "@")[0],
			Email: contact.Email,
		}
		slice = append(slice, contactStruct)
	}

	return slice
}

func getContacts() ContactList {
	reqUrl := "https://api.sendinblue.com/v3/contacts"
	res := sibGetRequest(reqUrl)

	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var body ContactList
	json.Unmarshal(bodyBytes, &body)

	return body
}

func sibGetRequest(url string) *http.Response {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("api-key", os.Getenv("SENDINBLUE_KEY"))
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	return res
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
