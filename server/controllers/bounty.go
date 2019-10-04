package controllers

import (
	"fmt"
	"github.com/google/go-github/github"
	"github.com/iotaledger/iota.go/account"
	"github.com/iotaledger/iota.go/account/builder"
	"github.com/iotaledger/iota.go/account/deposit"
	"github.com/iotaledger/iota.go/account/plugins/promoter"
	"github.com/iotaledger/iota.go/account/store"
	mongostore "github.com/iotaledger/iota.go/account/store/mongo"
	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/pow"
	"github.com/luca-moser/iota-bounty-platform/server/misc"
	"github.com/luca-moser/iota-bounty-platform/server/models"
	"github.com/luca-moser/iota-bounty-platform/server/server/config"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"time"
)

const bountyCollection = "bounties"
const deletedBountyCollection = "deleted_bounties"

type BountyCtrl struct {
	Config   *config.Configuration `inject:""`
	GHClient *github.Client        `inject:""`
	RepoCtrl *RepoCtrl             `inject:""`
	Mongo    *mongo.Client         `inject:""`
	Coll     *mongo.Collection
	DelColl  *mongo.Collection
	Bot      *Bot `inject:""`
	logger   log15.Logger
	store    store.Store
	iotaAPI  *api.API
}

func (bc *BountyCtrl) Init() error {
	logger, err := misc.GetLogger("bounty-ctrl")
	if err != nil {
		return err
	}
	bc.logger = logger

	conf := bc.Config

	// init account module
	// init api
	_, powFunc := pow.GetFastestProofOfWorkImpl()
	bc.iotaAPI, err = api.ComposeAPI(api.HTTPClientSettings{
		URI: conf.Account.Node, LocalProofOfWorkFunc: powFunc,
		Client: &http.Client{Timeout: time.Duration(10) * time.Second},
	})
	if err != nil {
		return errors.Wrap(err, "unable to init IOTA API")
	}

	bc.store, err = mongostore.NewMongoStore(conf.DB.URI, &mongostore.Config{
		DBName: conf.DB.DBName, CollName: conf.Account.Collection,
	})

	// init db collections and indexes
	dbName := bc.Config.DB.DBName
	bc.Coll = bc.Mongo.Database(dbName).Collection(bountyCollection)
	bc.DelColl = bc.Mongo.Database(dbName).Collection(deletedBountyCollection)

	return nil
}

func (bc *BountyCtrl) GetAll() ([]models.Bounty, error) {
	bounties := []models.Bounty{}
	res, err := bc.Coll.Find(DefaultCtx(), bson.D{})
	if err != nil {
		return nil, err
	}
	for res.Next(DefaultCtx()) {
		var bounty models.Bounty
		if err := res.Decode(&bounty); err != nil {
			return nil, err
		}
		bounties = append(bounties, bounty)
	}
	return bounties, errors.Wrap(err, "(bounties) couldn't load all bounties")
}

func (bc *BountyCtrl) GetByID(id int64) (*models.Bounty, error) {
	res := bc.Coll.FindOne(DefaultCtx(), bson.D{
		{"_id", id},
	})
	if res.Err() != nil {
		return nil, res.Err()
	}
	bounty := &models.Bounty{}
	err := res.Decode(bounty)
	return bounty, errors.Wrapf(err, "(bounty) couldn't load bounty '%s'", id)
}

func (bc *BountyCtrl) GetByIssueNumber(repoID int64, issueID int) (*models.Bounty, error) {
	res := bc.Coll.FindOne(DefaultCtx(), bson.D{
		{"repository_id", repoID},
		{"issue_number", issueID},
	})
	if res.Err() != nil {
		return nil, res.Err()
	}
	bounty := &models.Bounty{}
	err := res.Decode(bounty)
	return bounty, errors.Wrapf(err, "(bounty) couldn't load bounty via repo id '%d' and issue number '%d'", repoID, issueID)
}

func (bc *BountyCtrl) GetOfRepository(owner string, name string) ([]models.Bounty, error) {
	repo, err := bc.RepoCtrl.GetByOwnerAndName(owner, name)
	if err != nil {
		fmt.Println("didn't find repository")
		return nil, err
	}

	bounties := []models.Bounty{}
	cursor, err := bc.Coll.Find(DefaultCtx(), bson.D{
		{"repository_id", repo.ID},
	})
	if err != nil {
		return nil, err
	}
	for cursor.Next(DefaultCtx()) {
		var bounty models.Bounty
		if err := cursor.Decode(&bounty); err != nil {
			return nil, err
		}
		bounties = append(bounties, bounty)
	}
	return bounties, errors.Wrapf(err, "(bounty) couldn't load bounties of repository %d", repo.ID)
}

func (bc *BountyCtrl) Add(owner string, repoName string, issueID int) (*models.Bounty, error) {

	repo, err := bc.RepoCtrl.GetByOwnerAndName(owner, repoName)
	if err != nil {
		return nil, ErrRepositoryNotInPlatform
	}

	issue, res, err := bc.GHClient.Issues.Get(DefaultCtx(), owner, repoName, issueID)
	if err != nil {
		if res != nil && res.StatusCode == 404 {
			return nil, ErrIssueDoesntExist
		}
		return nil, err
	}

	seed, err := misc.GenerateSeed()
	if err != nil {
		return nil, err
	}
	bounty := &models.Bounty{
		Model: models.Model{
			CreatedOn: time.Now(),
		},
		ID:           issue.GetID(),
		IssueNumber:  issue.GetNumber(),
		RepositoryID: repo.ID,
		ReceiverID:   0,
		Seed:         seed,
		URL:          issue.GetHTMLURL(),
		Title:        issue.GetTitle(),
		Body:         issue.GetBody(),
		State:        models.BountyStateOpen,
	}

	// initialize a new account for this issue
	acc, err := bc.LoadAccount(bounty.Seed)
	if err != nil {
		return nil, err
	}
	if err := acc.Start(); err != nil {
		return nil, err
	}

	// generate pool address, extra short timeout so it will be selected
	timeout := time.Now().Add(time.Duration(3) * time.Minute)
	cda, err := acc.AllocateDepositAddress(&deposit.Conditions{
		TimeoutAt: &timeout,
	})
	if err != nil {
		return nil, err
	}
	bounty.PoolAddress = cda.Address
	if err := acc.Shutdown(); err != nil {
		return nil, err
	}

	if _, err := bc.Coll.InsertOne(DefaultCtx(), bounty); err != nil {
		return nil, errors.Wrap(err, "(bounty) couldn't insert bounty")
	}

	// post message to the issue
	if err := bc.Bot.PostNewBountyMessage(owner, repoName, bounty); err != nil {
		return nil, err
	}

	return bounty, nil
}

func (bc *BountyCtrl) LoadAccount(seed string) (account.Account, error) {
	return builder.NewBuilder().
		WithSeed(seed).
		WithAPI(bc.iotaAPI).
		WithStore(bc.store).
		WithDepth(bc.Config.Account.GTTADepth).
		WithMWM(bc.Config.Account.MWM).
		WithSecurityLevel(consts.SecurityLevel(bc.Config.Account.SecurityLevel)).
		Build()
}

func (bc *BountyCtrl) LoadAccountForSending(seed string) (account.Account, error) {
	build := builder.NewBuilder().
		WithSeed(seed).
		WithAPI(bc.iotaAPI).
		WithStore(bc.store).
		WithDepth(bc.Config.Account.GTTADepth).
		WithMWM(bc.Config.Account.MWM).
		WithSecurityLevel(consts.SecurityLevel(bc.Config.Account.SecurityLevel))
	return build.Build(promoter.NewPromoter(build.Settings(), time.Duration(30)*time.Second))
}

func (bc *BountyCtrl) GetAccountBalance(seed string) (uint64, error) {
	acc, err := builder.NewBuilder().
		WithSeed(seed).
		WithAPI(bc.iotaAPI).
		WithStore(bc.store).
		WithDepth(bc.Config.Account.GTTADepth).
		WithMWM(bc.Config.Account.MWM).
		WithSecurityLevel(consts.SecurityLevel(bc.Config.Account.SecurityLevel)).
		Build()
	if err != nil {
		return 0, err
	}
	if err := acc.Start(); err != nil {
		return 0, err
	}
	balance, err := acc.AvailableBalance()
	if err != nil {
		return 0, err
	}
	if err := acc.Shutdown(); err != nil {
		return 0, err
	}
	return balance, nil
}

func (bc *BountyCtrl) ReleaseBounty(bounty *models.Bounty, receiverID int64) error {
	// load up account balance
	availBalance, err := bc.GetAccountBalance(bounty.Seed)
	if err != nil {
		return err
	}

	t := time.Now()
	mut := bson.D{{"$set", bson.D{
		{"state", models.BountyStateReleased},
		{"receiver_id", receiverID},
		{"balance", availBalance},
		{"model.updated_on", t},
	}}}
	_, err = bc.Coll.UpdateOne(DefaultCtx(), bson.D{{"_id", bounty.ID}}, mut)
	return errors.Wrapf(err, "(bounty) couldn't update bounty state '%s'", bounty.ID)
}

var ErrBountyAddrEmpty = errors.New("the bounty address has no funds")

func (bc *BountyCtrl) TransferBounty(bounty *models.Bounty, addr string) (bundle.Bundle, uint64, error) {
	// note, since TransferBounty is only called from within a issue comment handling
	// which is synchronized globally, it is safe to load the account and starting it
	acc, err := bc.LoadAccountForSending(bounty.Seed)
	if err != nil {
		return nil, 0, err
	}
	if err := acc.Start(); err != nil {
		return nil, 0, err
	}

	availBalance, err := acc.AvailableBalance()
	if err != nil {
		return nil, 0, err
	}

	if availBalance == 0 {
		return nil, 0, ErrBountyAddrEmpty
	}

	trsnf := account.Recipient{
		Address: addr,
		Value:   availBalance,
		Tag:     bundle.PadTag("IOTABOUNTY"),
	}

	bndl, err := acc.Send(trsnf)
	if err != nil {
		return nil, 0, err
	}

	t := time.Now()
	mut := bson.D{{"$set", bson.D{
		{"state", models.BountyStateTransferred},
		{"receiver_address", addr},
		{"bundle_hash", bndl[0].Bundle},
		{"balance", availBalance},
		{"model.updated_on", t},
	}}}
	if _, err = bc.Coll.UpdateOne(DefaultCtx(), bson.D{{"_id", bounty.ID}}, mut); err != nil {
		return nil, 0, errors.Wrapf(err, "(bounty) couldn't update bounty state '%d'", bounty.ID)
	}

	return bndl, availBalance, nil
}

func (bc *BountyCtrl) SyncBounties() {
	bounties, err := bc.GetAll()
	if err != nil {
		bc.logger.Error(fmt.Sprintf("can't load all bounties for sync: %s", err.Error()))
		return
	}

	for i := range bounties {
		bounty := &bounties[i]
		bc.logger.Info(fmt.Sprintf("syncing bounty: %d/%s", bounty.ID, bounty.Title))
		if err := bc.SyncBounty(bounty); err != nil {
			bc.logger.Error(fmt.Sprintf("can't sync bounty %d/%s: %s", bounty.ID, bounty.Title, err.Error()))
		}
	}
}

func (bc *BountyCtrl) SyncBounty(bounty *models.Bounty) error {

	repo, err := bc.RepoCtrl.GetByID(bounty.RepositoryID)
	if err != nil {
		return ErrRepositoryNotInPlatform
	}

	issue, res, err := bc.GHClient.Issues.Get(DefaultCtx(), repo.Owner, repo.Name, bounty.IssueNumber)
	if err != nil {
		if res != nil && res.StatusCode == 404 {
			// delete the bounty as the associated issue no longer exists
			if err := bc.Delete(bounty.ID); err != nil {
				return err
			}
			return ErrIssueDoesntExist
		}
		return err
	}

	balance := bounty.Balance
	// only updated bounty balance if it wasn't transferred yet
	if bounty.State != models.BountyStateTransferred {
		balance, err = bc.GetAccountBalance(bounty.Seed)
		if err != nil {
			return err
		}
	}

	t := time.Now()
	mut := bson.D{{"$set", bson.D{
		{"title", issue.GetTitle()},
		{"body", issue.GetBody()},
		{"url", issue.GetHTMLURL()},
		{"balance", balance},
		{"model.updated_on", t},
	}}}

	_, err = bc.Coll.UpdateOne(DefaultCtx(), bson.D{{"_id", bounty.ID}}, mut)
	return errors.Wrapf(err, "(bounty) couldn't update bounty '%d'", bounty.ID)
}

func (bc *BountyCtrl) Delete(id int64, repo ...*models.Repository) error {
	bounty, err := bc.GetByID(id)
	if err != nil {
		return err
	}

	if bounty.State != models.BountyStateTransferred {
		// load the repo if not given
		var r *models.Repository
		if len(repo) > 0 {
			r = repo[0]
		} else {
			r, err = bc.RepoCtrl.GetByID(bounty.RepositoryID)
			if err != nil {
				return err
			}
		}

		// ignore error as we want to delete the bounty whether we fail posting or not
		bc.Bot.PostBountyDeletedFromPlatformMessage(r.Owner, r.Name, bounty)
	}

	if _, err = bc.Coll.DeleteOne(DefaultCtx(), bson.D{{"_id", id}}); err != nil {
		return errors.Wrapf(err, "(bounty) couldn't delete bounty '%d'", id)
	}
	_, err = bc.DelColl.InsertOne(DefaultCtx(), models.DeletedModel{Object: bounty})
	return errors.Wrapf(err, "(bounty) couldn't move bounty '%s' to deleted collection", id)
}
