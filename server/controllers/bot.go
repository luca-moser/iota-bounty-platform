package controllers

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/guards"
	"github.com/luca-moser/iota-bounty-platform/server/misc"
	"github.com/luca-moser/iota-bounty-platform/server/models"
	"github.com/luca-moser/iota-bounty-platform/server/server/config"
	"go.mongodb.org/mongo-driver/mongo"
	gwb "gopkg.in/go-playground/webhooks.v5/github"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const newBountyMessage = `
This issue has been linked with the bounty platform.  
Help rising the incentive to solve this issue by sending iota tokens to the following address:
[%s](https://thetangle.org/address/%s)

> Please note that the tokens you send to the address can not be recovered

Important:
**If you move this repository make sure to await for the bounty platform to synchronize the repository state before releasing a bounty.**

#### Releasing the bounty (as a repository admin)
Release the bounty by issuing following comment:
` + "`release bounty to @<bounty_receiver_name>`" + `

#### Receiving the bounty (as the issue solver)
Simply create a comment with your IOTA address (+checksum, must be 90 chars long!) to which to receive the tokens to after the above 'release comment' has been posted.
`

const bountyIsReleasedMessage = `
The bounty of %d iotas has been released.
@%s please post your receiving IOTA address as a comment.
The receiver of the bounty can still be changed by issuing the bounty release command again.
`

const bountyReceiverHasBeenUpdatedMessage = `
The receiver of the bounty of %d iotas has been updated to %s.
@%s please post your receiving IOTA address as a comment.
The receiver of the bounty can still be changed by issuing the bounty release command again.
`

const bountySentMessage = `
Hey @%s, the bounty of %d iotas has been sent off. Bundle: [%s](https://thetangle.org/bundle/%s).
`

const bountyDeletedMessage = `
The bounty associated with this issue has been deleted from the bounty platform, therefore
the bounty is no longer active.
`

var srv = &http.Server{}

func ShutdownWebHookListener() {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1500)*time.Millisecond)
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println("couldn't shutdown web hook listener cleanly")
	}
}

const (
	actionCreated = "created"
	actionOpened  = "opened"
	actionDeleted = "deleted"
	actionEdited  = "edited"
)

// lets use a global lock for easy synchronisation, contention should never be a problem
var processMu = sync.Mutex{}

type Bot struct {
	Config     *config.Configuration `inject:""`
	GHClient   *github.Client        `inject:""`
	RepoCtrl   *RepoCtrl             `inject:""`
	BountyCtrl *BountyCtrl           `inject:""`
	Mongo      *mongo.Client         `inject:""`
	logger     log15.Logger
}

func (b *Bot) Init() error {
	logger, err := misc.GetLogger("bot")
	if err != nil {
		return err
	}
	b.logger = logger
	go b.Run()
	return nil
}

func (b *Bot) Run() {
	b.InstallWebHooks()
	go b.ListenToWebHooks()
	for {
		b.Sync()
		time.Sleep(time.Duration(b.Config.GitHub.SyncIntervalSeconds) * time.Second)
	}
}

func (b *Bot) Sync() {
	processMu.Lock()
	defer processMu.Unlock()
	b.RepoCtrl.SyncRepositories()
	b.BountyCtrl.SyncBounties()
}

func (b *Bot) ListenToWebHooks() {
	hook, _ := gwb.New()

	ghConf := b.Config.GitHub

	http.HandleFunc(ghConf.WebHook.URLPath, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, gwb.IssueCommentEvent)
		if err != nil {
			b.logger.Error(fmt.Sprintf("error while parsing event from GitHub web hook: %s", err.Error()))
			return
		}

		switch t := payload.(type) {
		case gwb.IssueCommentPayload:
			evRepo := t.Repository

			// check whether the repository is even known to the platform
			repo, err := b.RepoCtrl.GetByID(t.Repository.ID)
			if err != nil {
				b.logger.Warn(fmt.Sprintf("got an issue event via web hook of a repository which is "+
					"not registered on the bounty platform: %d/%s/%s", evRepo.ID, evRepo.Owner.Login, evRepo.Name))
				return
			}

			// check whether the issue is linked with a bounty
			bounty, err := b.BountyCtrl.GetByID(t.Issue.ID)
			if err != nil {
				// issue is not linked to a bounty
				b.logger.Warn(fmt.Sprintf("new issue comment event on repository %d/%s/%s which is not linked to a bounty: %d/%s", evRepo.ID, evRepo.Owner.Login, evRepo.Name, t.Issue.ID, t.Issue.Title))
				return
			}

			// don't handle anything on an already transferred bounty
			if bounty.State == models.BountyStateTransferred {
				return
			}

			switch t.Action {
			case actionCreated:
				b.logger.Info(fmt.Sprintf("new comment on repository %d/%s/%s; issue %d/%s", repo.ID, repo.Owner, repo.Name, t.Issue.ID, t.Issue.Title))
				b.HandleIssueComment(t, bounty, repo)
			}
		default:
			b.logger.Warn(fmt.Sprintf("got an event via web hook of a non wanted type: %s", payload))
		}
	})

	srv.Addr = ghConf.WebHook.ListenAddress
	b.logger.Info(fmt.Sprintf("listening for web hook events via %s%s", ghConf.WebHook.ListenAddress, ghConf.WebHook.URLPath))
	if err := srv.ListenAndServe(); err != nil {
		b.logger.Error(fmt.Sprintf("unable to setup web hooks listener: %s", err.Error()))
		os.Exit(-1)
	}
}

func (b *Bot) InstallWebHooks() {
	b.logger.Info("checking/installing web hooks on repositories...")

	repositories, err := b.RepoCtrl.GetAll()
	if err != nil {
		b.logger.Error("couldn't load repositories " + err.Error())
		panic(err)
	}

	confWebHookURL := b.Config.GitHub.WebHook.URL
	b.logger.Info(fmt.Sprintf("checking for web hook %s", confWebHookURL))

	// go over each repository and check whether the config defined hook url is still installed
	for _, repo := range repositories {
		_, res, err := b.GHClient.Repositories.GetByID(DefaultCtx(), repo.ID)
		if err != nil {
			if res != nil && res.StatusCode == 404 {
				b.logger.Error(fmt.Sprintf("couldn't load hooks of repository %d/%s/%s. did you set the repository private and forgot to add permission for the used bounty platform account?", repo.ID, repo.Owner, repo.Name))
				continue
			}
			b.logger.Error(fmt.Sprintf("couldn't load repository %d/%s/%s: %s ", repo.ID, repo.Owner, repo.Name, err.Error()))
			continue
		}

		hooks, res, err := b.GHClient.Repositories.ListHooks(DefaultCtx(), repo.Owner, repo.Name, &github.ListOptions{})
		if err != nil {
			if res != nil && res.StatusCode == 404 {
				b.logger.Error(fmt.Sprintf("couldn't load hooks of repository %d/%s/%s because the used bounty platform account has no permissions or the repository no longer exists", repo.ID, repo.Owner, repo.Name))
				continue
			}
			b.logger.Error(fmt.Sprintf("couldn't load hooks of repository %d/%s/%s: %s ", repo.ID, repo.Owner, repo.Name, err.Error()))
			continue
		}

		var installed bool
		for _, hook := range hooks {
			if hook.Config["url"] == confWebHookURL+b.Config.GitHub.WebHook.URLPath {
				installed = true
				break
			}
		}

		if installed {
			b.logger.Info(fmt.Sprintf("web hook for repository %d/%s/%s is installed", repo.ID, repo.Owner, repo.Name))
			return
		}

		b.logger.Info(fmt.Sprintf("installing web hook for repository %d/%s/%s...", repo.ID, repo.Owner, repo.Name))
		createdHook, _, err := b.GHClient.Repositories.CreateHook(DefaultCtx(), repo.Owner, repo.Name, &github.Hook{
			Name:   github.String("web"),
			Events: []string{"issue_comment"},
			Active: github.Bool(true),
			Config: map[string]interface{}{
				"url":          confWebHookURL + b.Config.GitHub.WebHook.URLPath,
				"content_type": "json",
				"insecure_ssl": func() int {
					if b.Config.GitHub.WebHook.TLS {
						return 0
					}
					return 1
				}(),
			},
		})
		if err != nil {
			b.logger.Error(fmt.Sprintf("couldn't create web hook of repository %d/%s/%s: %s ", repo.ID, repo.Owner, repo.Name, err.Error()))
			continue
		}

		b.logger.Info(fmt.Sprintf("web hook for repository %d/%s/%s with id %d successfully installed", repo.ID, repo.Owner, repo.Name, createdHook.GetID()))
	}
}

func (b *Bot) PostNewBountyMessage(owner string, repo string, bounty *models.Bounty) error {
	comment := &github.IssueComment{
		Body: github.String(fmt.Sprintf(newBountyMessage, bounty.PoolAddress, bounty.PoolAddress)),
	}
	_, _, err := b.GHClient.Issues.CreateComment(DefaultCtx(), owner, repo, bounty.IssueNumber, comment)
	if err != nil {
		return err
	}
	b.logger.Info(fmt.Sprintf("posted new bounty message on: %s/%s issue %d - %s", owner, repo, bounty.IssueNumber, bounty.Title))
	return nil
}

func (b *Bot) PostBountyDeletedFromPlatformMessage(owner string, repo string, bounty *models.Bounty) error {
	comment := &github.IssueComment{
		Body: github.String(bountyDeletedMessage),
	}
	_, _, err := b.GHClient.Issues.CreateComment(DefaultCtx(), owner, repo, bounty.IssueNumber, comment)
	if err != nil {
		return err
	}
	b.logger.Info(fmt.Sprintf("posted bounty deleted from platform on: %s/%s issue %d - %s", owner, repo, bounty.IssueNumber, bounty.Title))
	return nil
}

func (b *Bot) PostBountyReleasedMessage(owner string, repo string, bounty *models.Bounty) error {
	receiver, _, err := b.GHClient.Users.GetByID(DefaultCtx(), bounty.ReceiverID)
	if err != nil {
		return err
	}

	comment := &github.IssueComment{
		Body: github.String(fmt.Sprintf(bountyIsReleasedMessage, bounty.Balance, receiver.GetLogin())),
	}
	_, _, err = b.GHClient.Issues.CreateComment(DefaultCtx(), owner, repo, bounty.IssueNumber, comment)
	if err != nil {
		return err
	}
	b.logger.Info(fmt.Sprintf("posted bounty released message on: %s/%s issue %d - %s", owner, repo, bounty.IssueNumber, bounty.Title))
	return nil
}

func (b *Bot) PostBountyReceiverUpdatedMessage(owner string, repo string, bounty *models.Bounty) error {
	receiver, _, err := b.GHClient.Users.GetByID(DefaultCtx(), bounty.ReceiverID)
	if err != nil {
		return err
	}

	comment := &github.IssueComment{
		Body: github.String(fmt.Sprintf(bountyReceiverHasBeenUpdatedMessage, bounty.Balance, receiver.GetLogin(), receiver.GetLogin())),
	}
	_, _, err = b.GHClient.Issues.CreateComment(DefaultCtx(), owner, repo, bounty.IssueNumber, comment)
	if err != nil {
		return err
	}
	b.logger.Info(fmt.Sprintf("posted bounty updated receiver message on: %s/%s issue %d - %s", owner, repo, bounty.IssueNumber, bounty.Title))
	return nil
}

func (b *Bot) PostBountySentMessage(owner string, repo string, bounty *models.Bounty, value uint64, bundleHash string) error {
	receiver, _, err := b.GHClient.Users.GetByID(DefaultCtx(), bounty.ReceiverID)
	if err != nil {
		return err
	}

	comment := &github.IssueComment{
		Body: github.String(fmt.Sprintf(bountySentMessage, receiver.GetLogin(), value, bundleHash, bundleHash)),
	}
	_, _, err = b.GHClient.Issues.CreateComment(DefaultCtx(), owner, repo, bounty.IssueNumber, comment)
	if err != nil {
		return err
	}
	b.logger.Info(fmt.Sprintf("posted bounty sent message on: %s/%s issue %d - %s", owner, repo, bounty.IssueNumber, bounty.Title))
	return nil
}

var releaseBountyCmd = "release bounty to @"

var receiverNameInvalidMessage = `
Couldn't read out receiver name from the bounty release command.
Please make sure you use the appropriate syntax of: 
` + "`release bounty to @<bounty_receiver_name>`"

var receiverNotFoundOnGitHubMessage = `
Couldn't find the user specified in the bounty release command.
Please make sure you use the appropriate syntax of: 
` + "`release bounty to @<bounty_receiver_name>`"

var releaseCommandIssuerIsNotRepoAdminMessage = `
Only the repository admins are allowed to issue bounty release commands.
`

var failedToTransferBountyErrorMessage = `
Unfortunately an error occurred while sending the bounty to your address.
Please reinitiate the sending by posting your address again.

Error message: %s
`

var bountyAddressHasNoFunds = `
Unfortunately it seems that the bounty was released but there are no funds on the bounty address.
Please reinitiate the sending by posting **your** address again once there are funds on the address.
`

var postedAddrHasInvalidChecksumMessage = `The posted message has an invalid checksum.`

func (b *Bot) HandleIssueComment(issuePayload gwb.IssueCommentPayload, bounty *models.Bounty, repo *models.Repository) {
	processMu.Lock()
	defer processMu.Unlock()

	comment := strings.TrimSpace(issuePayload.Comment.Body)
	b.logger.Info("handling comment: " + comment)

	// transfer command/address
	if guards.IsAddressWithChecksum(comment) {
		b.HandleBountyTransfer(issuePayload, bounty, repo, comment)
		return
	}

	if !strings.HasPrefix(comment, releaseBountyCmd) {
		return
	}

	b.HandleBountyRelease(issuePayload, bounty, repo, comment)
}

func (b *Bot) HandleBountyTransfer(issuePayload gwb.IssueCommentPayload, bounty *models.Bounty, repo *models.Repository, addr string) {

	// check whether the bounty has actually been marked as released
	if bounty.State != models.BountyStateReleased {
		b.logger.Error(fmt.Sprintf("ignoring posted address as bounty has not been released"))
		return
	}

	// check whether the correct receiver has sent the message
	if issuePayload.Sender.ID != bounty.ReceiverID {
		b.logger.Error(fmt.Sprintf("ignoring posted address as the comment creator doesn't match the bounty receiver"))
		return
	}

	// check whether the address checksum is correct
	if err := address.ValidChecksum(addr[:81], addr[81:]); err != nil {
		b.logger.Error(fmt.Sprintf("posted address has an invalid checksum"))
		comment := &github.IssueComment{Body: github.String(postedAddrHasInvalidChecksumMessage)}
		_, _, err := b.GHClient.Issues.CreateComment(DefaultCtx(), repo.Owner, repo.Name, bounty.IssueNumber, comment)
		if err != nil {
			b.logger.Info(fmt.Sprintf("unable to write wrong address checksum error message: %s", err.Error()))
		}
		return
	}

	bndl, value, err := b.BountyCtrl.TransferBounty(bounty, addr)
	if err != nil {
		b.logger.Error(fmt.Sprintf("failed to send bounty: %s", err.Error()))
		// bounty address is actually empty, so we can't send anything yet
		if err == ErrBountyAddrEmpty {
			comment := &github.IssueComment{Body: github.String(bountyAddressHasNoFunds)}
			_, _, err := b.GHClient.Issues.CreateComment(DefaultCtx(), repo.Owner, repo.Name, bounty.IssueNumber, comment)
			if err != nil {
				b.logger.Info(fmt.Sprintf("unable to write bounty address empty error message: %s", err.Error()))
			}
			return
		}

		comment := &github.IssueComment{Body: github.String(fmt.Sprintf(failedToTransferBountyErrorMessage, err.Error()))}
		_, _, err := b.GHClient.Issues.CreateComment(DefaultCtx(), repo.Owner, repo.Name, bounty.IssueNumber, comment)
		if err != nil {
			b.logger.Info(fmt.Sprintf("unable to write bounty transfer failed error message: %s", err.Error()))
		}
		return
	}

	if err := b.PostBountySentMessage(repo.Owner, repo.Name, bounty, value, bndl[0].Bundle); err != nil {
		b.logger.Error(fmt.Sprintf("unable to post bounty transffered message: %s", err.Error()))
	}
}

func (b *Bot) HandleBountyRelease(issuePayload gwb.IssueCommentPayload, bounty *models.Bounty, repo *models.Repository, comment string) {

	collaborators, _, err := b.GHClient.Repositories.ListCollaborators(DefaultCtx(), repo.Owner, repo.Name, &github.ListCollaboratorsOptions{})
	if err != nil {
		b.logger.Error(fmt.Sprintf("unable to fetch repository collaborators from GitHub: %s", err.Error()))
		return
	}

	var isAdmin bool
	for _, collaborator := range collaborators {
		if collaborator.GetID() == issuePayload.Sender.ID {
			admin, has := collaborator.GetPermissions()["admin"]
			if has && admin {
				isAdmin = true
			}
			break
		}
	}

	if !isAdmin {
		b.logger.Error("release command issuer is not a repository admin")
		comment := &github.IssueComment{Body: github.String(releaseCommandIssuerIsNotRepoAdminMessage)}
		_, _, err := b.GHClient.Issues.CreateComment(DefaultCtx(), repo.Owner, repo.Name, bounty.IssueNumber, comment)
		if err != nil {
			b.logger.Info(fmt.Sprintf("unable to write wrong release command issuer error message: %s", err.Error()))
		}
		return
	}

	// bounty release command
	receiverLoginName := strings.TrimSpace(strings.TrimPrefix(comment, releaseBountyCmd))
	if receiverLoginName == "" {
		b.logger.Info("can't release funds as receiver name is invalid")
		comment := &github.IssueComment{Body: github.String(receiverNameInvalidMessage),}
		_, _, err := b.GHClient.Issues.CreateComment(DefaultCtx(), repo.Owner, repo.Name, bounty.IssueNumber, comment)
		if err != nil {
			b.logger.Info(fmt.Sprintf("unable to write receiver extraction error message: %s", err.Error()))
		}
		return
	}

	b.logger.Info(fmt.Sprintf("extracted receiver of bounty: %s", receiverLoginName))
	receiver, _, err := b.GHClient.Users.Get(DefaultCtx(), receiverLoginName)
	if err != nil {
		b.logger.Error(fmt.Sprintf("couldn't fetch bounty receiver: %s", err.Error()))
		comment := &github.IssueComment{Body: github.String(receiverNotFoundOnGitHubMessage)}
		_, _, err := b.GHClient.Issues.CreateComment(DefaultCtx(), repo.Owner, repo.Name, bounty.IssueNumber, comment)
		if err != nil {
			b.logger.Info(fmt.Sprintf("unable to write receiver not found message: %s", err.Error()))
		}
		return
	}

	// check whether bounty was already released
	bountyAlreadyReleased := bounty.ReceiverID != 0
	b.logger.Info(fmt.Sprintf("setting bounty as released to: %s - %s - ID: %d", receiverLoginName, receiver.GetName(), receiver.GetID()))

	// this also automatically updates the receiver if previously set
	if err := b.BountyCtrl.ReleaseBounty(bounty, receiver.GetID()); err != nil {
		b.logger.Error(fmt.Sprintf("couldn't update bounty state: %s", err.Error()))
		return
	}

	if bountyAlreadyReleased {
		b.logger.Error("bounty receiver id updated successfully")
	} else {
		b.logger.Error("bounty state updated successfully to released")
	}

	updatedBounty, err := b.BountyCtrl.GetByID(bounty.ID)
	if err != nil {
		b.logger.Error(fmt.Sprintf("unable to load released bounty: %s", err.Error()))
		return
	}
	if bountyAlreadyReleased {
		if err := b.PostBountyReceiverUpdatedMessage(repo.Owner, repo.Name, updatedBounty); err != nil {
			b.logger.Error(fmt.Sprintf("unable to post bounty updated receiver message: %s", err.Error()))
		}
	} else {
		if err := b.PostBountyReleasedMessage(repo.Owner, repo.Name, updatedBounty); err != nil {
			b.logger.Error(fmt.Sprintf("unable to post bounty released message: %s", err.Error()))
		}
	}

	return
}
