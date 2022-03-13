package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/robfig/cron"
)

var day int
var hour int

func StartCrons() {
	localTimezone := "Australia/Adelaide"
	timezone, err := time.LoadLocation(localTimezone)
	if err != nil {
		log.Fatal(err)
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
		oneOffJob.Stop()
	})
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
