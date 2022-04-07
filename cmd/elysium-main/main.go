package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"github.com/robfig/cron"
	sendinblue "github.com/sendinblue/APIv3-go-library/lib"
	"google.golang.org/api/iterator"
)

var day int
var hour int

var previousUrls []string

type Poem []struct {
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	Lines     []string `json:"lines"`
	Linecount string   `json:"linecount"`
}
type Email struct {
	Sender      Sender   `json:"sender"`
	To          []string `json:"to"`
	Subject     string   `json:"subject"`
	HTMLContent string   `json:"htmlContent"`
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
		rand.Seed(time.Now().UnixNano())
		addJobForWeek(
			randomNumExcludeParameter(day, 1, 5),
			randomNumExcludeParameter(hour, 9, 17),
			timezone,
		)
	})

	weeklyJob.Start()
}

func addJobForWeek(day int, hour int, timezone *time.Location) {
	cronString := fmt.Sprintf("0 0 %d * * %d", hour, day)
	fmt.Println(cronString)
	oneOffJob := cron.NewWithLocation(timezone)
	oneOffJob.AddFunc(cronString, func() {
		sendEmail()
		oneOffJob.Stop()
	})

	oneOffJob.Start()
}

func sendEmail() {
	// Get all contacts
	contactsRaw := getContacts()

	// Process contacts
	contactsSlice := processContacts(contactsRaw)

	// Get image from bucket
	imageUrl := randStringSlice(getGcsUrls())

	// Get poem from somewhere
	poem := getRandomPoem()

	// Construct HTML
	html := constructHtml(imageUrl, poem)

	// Construct transactional query
	// email := constructSibQuery(html, contactsSlice)

	// Fire transactional email
	sendTransactionalEmail(contactsSlice, html)
}

func sendTransactionalEmail(
	contactsSlice []sendinblue.SendSmtpEmailTo,
	html string,
) {
	var ctx context.Context
	cfg := sendinblue.NewConfiguration()
	cfg.AddDefaultHeader("api-key", os.Getenv("SENDINBLUE_KEY"))
	body := sendinblue.SendSmtpEmail{
		Sender: &sendinblue.SendSmtpEmailSender{
			Name:  "Elly",
			Email: "elly@thecaninecosmos.com",
		},
		To:          contactsSlice,
		HtmlContent: html,
		Subject:     "Woof",
	}

	sib := sendinblue.NewAPIClient(cfg)
	_, _, err := sib.TransactionalEmailsApi.SendTransacEmail(ctx, body)
	if err != nil {
		panic(err)
	}
	fmt.Println("Email sent!")
}

func constructSibQuery(html string, toSlice []string) Email {
	emailJson := Email{
		Sender: Sender{
			Name:  "Elly",
			Email: "elly@thecaninecosmos.dev",
		},
		To:          toSlice,
		Subject:     "Woof",
		HTMLContent: html,
	}

	return emailJson
}

func constructHtml(url string, poemStruc Poem) string {
	dat, err := os.ReadFile("./html/email_template.txt")
	if err != nil {
		panic(err)
	}

	html := strings.Replace(string(dat), "HTTPREPLACE", url, 1)
	poem := "<p><strong>" + poemStruc[0].Title + "</strong></p><p>" + strings.Join(poemStruc[0].Lines[:], "</p><p>") + "</p><p>- " + poemStruc[0].Author + "</p>"
	html = strings.Replace(string(html), "TEXTREPLACE", poem, 1)

	return html
}

func getGcsUrls() []string {
	bucket := os.Getenv("GCS_BUCKET")
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var urls []string
	it := client.Bucket(bucket).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(err)
		}
		urls = append(urls, attrs.MediaLink)
	}
	return urls
}

func processContacts(rawContacts ContactList) []sendinblue.SendSmtpEmailTo {
	var slice []sendinblue.SendSmtpEmailTo
	for _, contact := range rawContacts.Contacts {
		contactStruct := sendinblue.SendSmtpEmailTo{
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

func getRandomPoem() Poem {
	req, _ := http.NewRequest("GET", "https://poetrydb.org/random/1", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var body Poem
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
	unique := false
	var newNum int
	for !unique {
		newNum = randInt(min, max)
		if newNum != exclude {
			unique = true
		}
	}
	return newNum
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randStringSlice(slice []string) string {
	var url string
	if len(slice) == len(previousUrls) {
		previousUrls = nil
	}
	for {
		url = slice[rand.Intn(len(slice))]
		if !contains(previousUrls, url) {
			break
		}
		rand.Seed(time.Now().UnixNano())
	}
	previousUrls = append(previousUrls, url)
	return url
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Init the .env file if not running in production
func init() {
	env := os.Getenv("ENVIRONMENT")
	if env != "production" && env != "docker" {
		err := godotenv.Load(".env")
		if err != nil {
			panic("Error loading .env file")
		}
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func testMail(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-API-KEY") == os.Getenv("API_KEY") {
		sendEmail()
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["message"] = "Email test sent"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}
		w.Write(jsonResp)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["message"] = "Unauthorized"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}
		w.Write(jsonResp)
	}
	return
}

func main() {
	fmt.Println("Starting crons")
	StartCrons()
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/test-email", testMail)
	http.ListenAndServe(":7070", nil)
}
