package main

import (
	"fmt"
	"github.com/haisum/jirawork/pkg/jirahttp"
	"github.com/haisum/jirawork/pkg/jirahttp/activity"
	"github.com/haisum/jirawork/pkg/tickets"
	"log"
	"os"
	"time"
	"strings"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/haisum/jirawork/pkg/jirahttp/api"
)


var (
	days = []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	app = kingpin.New("jirawork", "jirawork is tool for finding and logging time on work done on JIRA tickets")
	list     = app.Command("list", "finds and lists tickets which have been done some activity on by user within given week. See \"./jirawork list --help\" for help about parameters.")
	listURL = list.Flag("url", "jira url such as https://jira.myorg.com/jira. You can use JIRAWORK_URL environment variable to define URL instead of passing it in this flag.").Required().Envar("JIRAWORK_URL").URL()
	listUsername = list.Flag("username", "jira username. You can use JIRAWORK_USERNAME environment variable to define username instead of passing it in this flag.").Required().Envar("JIRAWORK_USERNAME").String()
	listPassword = list.Flag("password", "jira password. You can use JIRAWORK_PASSWORD environment variable to define password instead of passing it in this flag.").Required().Envar("JIRAWORK_PASSWORD").String()
	listTicketPrefix = list.Flag("prefix", "jira ticket prefix such as JIRA- or ENGR-. . You can use JIRAWORK_TICKET_PREFIX environment variable to define prefix instead of passing it in this flag.").Required().Envar("JIRAWORK_TICKET_PREFIX").String()
	listFormat  = list.Flag("format", "you can customize the way this command outputs tickets. Default format is {title} - {summary} - {date}. You can use JIRAWORK_FORMAT environment variable to define format instead of passing it in this flag").Envar("JIRAWORK_FORMAT").Default("{title} - {summary} - {date}").String()
	listDate  = list.Arg("date", "any date within week in dd-mm-yyyy format. You will get list of all tickets between Sunday and Saturday of week").Required().String()

	workLog = app.Command("log", "logs time on given ticket. See \"./jirawork log --help\" for help about parameters.")
	workLogURL = workLog.Flag("url", "jira url such as https://jira.myorg.com/jira. You can use JIRAWORK_URL environment variable to define URL instead of passing it in this flag.").Required().Envar("JIRAWORK_URL").URL()
	workLogUsername = workLog.Flag("username", "jira username. You can use JIRAWORK_USERNAME environment variable to define username instead of passing it in this flag.").Required().Envar("JIRAWORK_USERNAME").String()
	workLogPassword = workLog.Flag("password", "jira password. You can use JIRAWORK_PASSWORD environment variable to define password instead of passing it in this flag.").Required().Envar("JIRAWORK_PASSWORD").String()
	workLogDate  = workLog.Arg("date", "any date within week in dd-mm-yyyy format. You will get list of all tickets between Sunday and Saturday of week").Required().String()
	workLogTicket = workLog.Arg("ticket", "ticket title in PREFIX-xxxxx format").Required().String()
	workLogDay = workLog.Arg("day", "day of the week to log work on. Possible values: " + strings.Join(days, "|")).Required().Enum(days...)
	workLogTs = workLog.Arg("timeSpent", "time spent in XXm or XXh format. Examples: 10m or 5h").Required().String()
	workLogComment = workLog.Arg("comment", "description of work done").Required().String()

	)

func main() {
	kingpin.Version("0.0.1")
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// List tickets
	case list.FullCommand():
		jURL := *listURL
		config, err := jirahttp.NewConfig(jURL.String(), *listUsername, *listPassword, func(_ ...interface{}){})
		if err != nil {
			log.Fatal("error in config", err)
		}
		activityCl := activity.NewClient(config)
		givenDate,err  := time.ParseInLocation("02-01-2006", *listDate, time.Local)
		if err != nil {
			log.Fatal(err)
		}
		daysTillPrevSunday := -int(givenDate.Weekday())
		after := givenDate.AddDate(0, 0, daysTillPrevSunday)
		before := after.AddDate(0,0, int(time.Saturday))
		ts, err := tickets.Get(activityCl, *listTicketPrefix, makeTimestamp(after), makeTimestamp(before))
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range ts {
			fmt.Println(t.Format(*listFormat))
		}
	// Log work on ticket
	case workLog.FullCommand():
		jURL := *workLogURL
		config, err := jirahttp.NewConfig(jURL.String(), *workLogUsername, *workLogPassword, func(_ ...interface{}){})
		if err != nil {
			log.Fatal("error in config", err)
		}
		apiCl := api.NewClient(config)
		givenDate,err  := time.ParseInLocation("02-01-2006", *workLogDate, time.Local)
		if err != nil {
			log.Fatal(err)
		}
		daysTillPrevSunday := -int(givenDate.Weekday())
		after := givenDate.AddDate(0, 0, daysTillPrevSunday)
		fmt.Println("start of week:", after)
		logDate := after.AddDate(0,0, indexOf(*workLogDay, days))
		workLog := tickets.WorkLog{
			Started: tickets.JIRATime(logDate),
			TimeSpent: *workLogTs,
			Comment: *workLogComment,
		}
		fmt.Println("log date: ", logDate)
		t := &tickets.Ticket{Title: *workLogTicket}
		err = t.LogWork(apiCl, workLog)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Done. URL: %s/browse/%s\n",*workLogURL, *workLogTicket)
	}
}

func makeTimestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func indexOf(val string, list []string) int {
	for i, v := range list {
		if v == val {
			return i
		}
	}
	return -1
}