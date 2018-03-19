package tickets

import (
	"fmt"
	"github.com/haisum/jirawork/pkg/jirahttp/activity"
	"github.com/haisum/jirawork/pkg/jirahttp/api"
	"strings"
	"time"
)

const (
	endpoint = "/rest/api/2/issue/%s/worklog"
)

type JIRATime time.Time

func (t JIRATime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s:00:00.000+0000\"", time.Time(t).UTC().Format("2006-01-02T15"))
	return []byte(stamp), nil
}

type WorkLog struct {
	TimeSpent string   `json:"timeSpent"`
	Comment   string   `json:"comment"`
	Started   JIRATime `json:"started"`
}

type Ticket struct {
	Title   string
	Summary string
	Updated time.Time
}

type Tickets []Ticket

func (ts Tickets) Append(t Ticket) Tickets {
	present := false
	for _, v := range ts {
		if v.Title == t.Title {
			present = true
		}
	}
	if !present {
		return append(ts, t)
	} else {
		return ts
	}
}

func (ts Tickets) Get(title string) *Ticket {
	for _, t := range ts {
		if t.Title == title {
			return &t
		}
	}
	return nil
}

func (t *Ticket) Format(format string) string {
	replacer := strings.NewReplacer("{title}", t.Title, "{summary}", t.Summary, "{date}", t.Updated.Format("02-01-2006"))
	return replacer.Replace(format)
}

func (t *Ticket) LogWork(client api.Client, w WorkLog) error {
	ep := fmt.Sprintf(endpoint, t.Title)
	return client.Post(ep, w)
}

type feed struct {
	Entry []struct {
		Updated        string `xml:"updated"`
		ActivityObject struct {
			Title   string `xml:"title"`
			Summary string `xml:"summary"`
		} `xml:"object"`
		ActivityTarget struct {
			Title   string `xml:"title"`
			Summary string `xml:"summary"`
		} `xml:"target"`
	} `xml:"entry"`
}

func Get(activityClient activity.Client, ticketPrefix string, after, before int64) (Tickets, error) {
	f := feed{}
	t := Tickets{}
	err := activityClient.Get(after, before, &f)
	if err != nil {
		return t, err
	}
	for _, entry := range f.Entry {
		if strings.HasPrefix(entry.ActivityObject.Title, ticketPrefix) {
			updated, _ := time.Parse(time.RFC3339, entry.Updated)
			t = t.Append(Ticket{entry.ActivityObject.Title, entry.ActivityObject.Summary, updated})
		} else if strings.HasPrefix(entry.ActivityTarget.Title, ticketPrefix) {
			updated, _ := time.Parse(time.RFC3339, entry.Updated)
			t = t.Append(Ticket{entry.ActivityTarget.Title, entry.ActivityTarget.Summary, updated})
		}
	}
	return t, nil
}
