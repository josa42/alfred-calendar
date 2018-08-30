package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/deanishe/awgo"
)

const (
	updateJobName = "checkForUpdate"
	repo          = "josa42/alfred-calendar"
)

var (
	flagCheck     bool
	wf            *aw.Workflow
	iconAvailable = &aw.Icon{Value: "icon/update.png"}
)

// Event :
type Event struct {
	Title    string
	Calendar string
	From     string
	To       string
	Values   map[string]string
}

// IsFullDay :
func (e Event) IsFullDay() bool {
	return e.From == ""
}

// IsPast :
func (e Event) IsPast() bool {
	t, _ := time.Parse("15:04", e.To)
	n, _ := time.Parse("15:04", time.Now().Format("15:04"))
	return t.Unix() < n.Unix()
}

func init() {
	wf = aw.New()
}

// Your workflow starts here
func run() {
	wf.Args()
	flag.Parse()

	flag.Parsed()

	if flagCheck {
		runCheck()
		return
	}

	runTriggerCheck()

	icon := &aw.Icon{Value: "icon/event.png"}

	for _, event := range EventsToday() {

		if event.IsFullDay() || event.IsPast() {
			continue
		}

		wf.NewItem(event.Title).
			Subtitle(fmt.Sprintf("%s - %s (%s)", event.From, event.To, event.Calendar)).
			Icon(icon).
			Valid(true)
	}

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}

// EventsToday :
func EventsToday() []*Event {
	out, err := exec.Command("./icalBuddy", "-uid", "eventsToday").Output()
	if err != nil {
		log.Fatal(err)
	}

	expTitle := regexp.MustCompile(`^• (.+) \((.+)\)$`)
	expKey := regexp.MustCompile(`^    ([^:\s]+): (.+)$`)
	expValue := regexp.MustCompile(`^    \s+(.+)$`)
	expTime := regexp.MustCompile(`^    (\d\d:\d\d) - (\d\d:\d\d|...)$`)

	var event *Event
	events := []*Event{}
	key := ""

	for _, line := range strings.Split(string(out), "\n") {

		if expTitle.MatchString(line) {
			parts := expTitle.FindStringSubmatch(line)

			event = &Event{Title: parts[1], Calendar: parts[2], Values: map[string]string{}}
			events = append(events, event)

			continue
		}

		if expKey.MatchString(line) {
			parts := expKey.FindStringSubmatch(line)
			key = parts[1]
			event.Values[key] = parts[2]
			continue
		}

		if expValue.MatchString(line) {
			parts := expValue.FindStringSubmatch(line)
			event.Values[key] += "\n" + parts[1]
			continue
		}

		if expTime.MatchString(line) {
			parts := expTime.FindStringSubmatch(line)
			event.From = parts[1]
			event.To = parts[2]
			if event.To == "..." {
				event.To = "24:00"
			}
			continue
		}
	}

	return events
}

func runCheck() {
	wf.Configure(aw.TextErrors(true))
	log.Println("Checking for updates...")
	if err := wf.CheckForUpdate(); err != nil {
		wf.FatalError(err)
	}
}

func runTriggerCheck() {
	if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
		log.Println("Running update check in background...")

		cmd := exec.Command(os.Args[0], "-check")
		if err := wf.RunInBackground(updateJobName, cmd); err != nil {
			log.Printf("Error starting update check: %s", err)
		}
	}
}
