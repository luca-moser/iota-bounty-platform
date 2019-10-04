package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"net/http"

	gwb "gopkg.in/go-playground/webhooks.v5/github"
)

const (
	path = "/webhooks"
)

const (
	actionCreated = "created"
	actionOpened  = "opened"
	actionDeleted = "deleted"
	actionEdited  = "edited"
)

const token = "ac05debbe342d8f646e9c6fe6e13643cb0fd4277"
const eventNotDefinedErrMsg = "event not defined to be parsed"

const webhookURL = "http://80.218.171.223:3001/webhooks"

const ownerName = "ibp-org"
const repoName = "testbed"

func main() {
	createWebhook()
	listenToWebhook()
}

func initClient() *github.Client{
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return github.NewClient(oauth2.NewClient(ctx, ts))
}

func fetchIssue(){
	client := initClient()

	repo, _, err := client.Repositories.Get(context.Background(), ownerName, repoName)
	must(err)

	repo, _, err = client.Repositories.GetByID(context.Background(), *repo.ID)
	must(err)

	issue, _, err := client.Issues.Get(context.Background(), *repo.Owner.Name, *repo.Name, 1)
	must(err)

}

func createWebhook() {

	client := initClient()
	hooks, _, err := client.Repositories.ListHooks(ctx, ownerName, repoName, &github.ListOptions{})
	must(err)

	var installed bool
	for _, hook := range hooks {
		if hook.Config["url"] == webhookURL {
			installed = true
			break
		}
	}

	if installed {
		fmt.Println("webhook already installed")
		return
	}

	fmt.Println("installing webbhook")
	createdHook, _, err := client.Repositories.CreateHook(ctx, "ibp-org", "testbed", &github.Hook{
		Name:   github.String("web"),
		Events: []string{"meta", "issues", "issue_comment"},
		Active: github.Bool(true),
		Config: map[string]interface{}{
			"url":          webhookURL,
			"content_type": "json",
			"insecure_ssl": 1, // insecure
		},
	})
	must(err)

	fmt.Printf("webhook created with id %d, url => %s\n", *createdHook.ID, *createdHook.URL)
}

func listenToWebhook() {
	hook, _ := gwb.New()

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, gwb.IssuesEvent, gwb.IssueCommentEvent, gwb.PingEvent, gwb.MetaEvent)
		fmt.Println("new event")
		if err != nil {
			switch err.(type) {
			case *json.SyntaxError:
				fmt.Printf("please configure the webhook to send JSON payloads!\n")
			default:
				switch err.Error() {
				case eventNotDefinedErrMsg:
					fmt.Printf("please configure your webhook to solely emit events for issue, issue comments and meta (webhook deletion)\n")
				default:
					fmt.Printf("error while parsing event: %s, %T\n", err.Error(), err)
				}
			}
			return
		}

		switch t := payload.(type) {
		case gwb.PingPayload:
			fmt.Printf("webhook got registered\n")
		case gwb.MetaPayload:
			fmt.Printf("webhook got deleted\n")
		case gwb.IssuesPayload:
			switch t.Action {
			case actionOpened:
				fmt.Printf("new issue opened: '%s' by %s\n", t.Issue.Title, t.Sender.Login)
			case actionEdited:
				fmt.Printf("issue edited: '%s' by %s\n", t.Issue.Title, t.Sender.Login)
			case actionDeleted:
				fmt.Printf("issue deleted: '%s' by %s\n", t.Issue.Title, t.Sender.Login)
			}

		case gwb.IssueCommentPayload:
			switch t.Action {
			case actionCreated:
				fmt.Printf("new issue comment created: '%s' by %s\n", t.Comment.Body, t.Sender.Login)
			case actionEdited:
				fmt.Printf("issue comment edited: '%s' by %s\n", t.Comment.Body, t.Sender.Login)
			case actionDeleted:
				fmt.Printf("issue comment deleted: '%s' by %s\n", t.Comment.Body, t.Sender.Login)
			}
		}
	})

	fmt.Println("listening for webhooks events")
	http.ListenAndServe(":3000", nil)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
