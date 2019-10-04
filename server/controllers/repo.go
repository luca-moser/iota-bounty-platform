package controllers

import (
	"fmt"
	"github.com/google/go-github/github"
	"github.com/luca-moser/iota-bounty-platform/server/misc"
	"github.com/luca-moser/iota-bounty-platform/server/models"
	"github.com/luca-moser/iota-bounty-platform/server/server/config"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"gopkg.in/inconshreveable/log15.v2"
	"strings"
	"time"
)

const repoCollection = "repos"
const deletedRepoCollection = "deleted_repos"

type RepoCtrl struct {
	Config     *config.Configuration `inject:""`
	BountyCtrl *BountyCtrl           `inject:""`
	Bot        *Bot                  `inject:""`
	GHClient   *github.Client        `inject:""`
	Mongo      *mongo.Client         `inject:""`
	Coll       *mongo.Collection
	DelColl    *mongo.Collection
	logger     log15.Logger
}

func (rc *RepoCtrl) Init() error {
	logger, err := misc.GetLogger("repo-ctrl")
	if err != nil {
		return err
	}
	rc.logger = logger

	// init db collections and indexes
	dbName := rc.Config.DB.DBName
	rc.Coll = rc.Mongo.Database(dbName).Collection(repoCollection)
	rc.DelColl = rc.Mongo.Database(dbName).Collection(deletedRepoCollection)

	t := true
	f := false
	urlIndexName := "url"
	urlIndex := mongo.IndexModel{
		Keys: bsonx.Doc{{Key: "url", Value: bsonx.Int32(int32(1))}},
		Options: &options.IndexOptions{
			Name: &urlIndexName, Background: &f,
			Unique: &t, Sparse: &t,
			Collation: &options.Collation{
				Locale:   "en",
				Strength: 2,
			},
		},
	}

	indexes := []mongo.IndexModel{urlIndex}
	if _, err := rc.Coll.Indexes().CreateMany(DefaultCtx(), indexes); err != nil {
		return err
	}

	return nil
}

func (rc *RepoCtrl) GetAll() ([]models.Repository, error) {
	repos := []models.Repository{}
	res, err := rc.Coll.Find(DefaultCtx(), bson.D{})
	if err != nil {
		return nil, err
	}
	for res.Next(DefaultCtx()) {
		var repo models.Repository
		if err := res.Decode(&repo); err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, errors.Wrap(err, "(repo) couldn't load all repos")
}

func (rc *RepoCtrl) GetByID(id int64) (*models.Repository, error) {
	res := rc.Coll.FindOne(DefaultCtx(), bson.D{
		{"_id", id},
	})
	if res.Err() != nil {
		return nil, res.Err()
	}
	repo := &models.Repository{}
	err := res.Decode(repo)
	return repo, errors.Wrapf(err, "(repo) couldn't load repo '%d'", id)
}

func (rc *RepoCtrl) GetByOwnerAndName(owner string, name string) (*models.Repository, error) {
	res := rc.Coll.FindOne(DefaultCtx(), bson.D{
		{"owner", owner},
		{"name", name},
	})
	if res.Err() != nil {
		return nil, res.Err()
	}
	repo := &models.Repository{}
	err := res.Decode(repo)
	return repo, errors.Wrapf(err, "(repo) couldn't load repo '%s/%s'", owner, name)
}

func (rc *RepoCtrl) Add(owner string, name string) (*models.Repository, error) {
	// fetch repo from Github
	repo, _, err := rc.GHClient.Repositories.Get(DefaultCtx(), owner, name)
	if err != nil {
		return nil, err
	}

	// check that issues are enabled
	if !repo.GetHasIssues() {
		return nil, ErrIssuesDeactivated
	}

	repoModel := &models.Repository{
		Model: models.Model{
			CreatedOn: time.Now(),
		},
		ID:          *repo.ID,
		Owner:       strings.TrimSpace(owner),
		Name:        strings.TrimSpace(name),
		URL:         strings.TrimSpace(repo.GetHTMLURL()),
		Description: strings.TrimSpace(repo.GetDescription()),
	}

	if _, err := rc.Coll.InsertOne(DefaultCtx(), repoModel); err != nil {
		return nil, errors.Wrap(err, "(repo) couldn't insert repo")
	}

	// trigger the bot to re-check webhooks
	rc.Bot.InstallWebHooks()

	return repoModel, nil
}

func (rc *RepoCtrl) SyncRepositories() {
	repos, err := rc.GetAll()
	if err != nil {
		rc.logger.Error(fmt.Sprintf("can't load all repo for sync: %s", err.Error()))
		return
	}

	for i := range repos {
		repo := &repos[i]
		rc.logger.Info(fmt.Sprintf("syncing repo: %d/%s/%s", repo.ID, repo.Owner, repo.Name))
		if err := rc.SyncRepo(repo); err != nil {
			rc.logger.Error(fmt.Sprintf("can't sync repo %d/%s/%s: %s", repo.ID, repo.Owner, repo.Name, err.Error()))
		}
	}
}

func (rc *RepoCtrl) SyncRepo(repo *models.Repository) error {
	ghRepo, res, err := rc.GHClient.Repositories.GetByID(DefaultCtx(), repo.ID)
	if err != nil {
		if res != nil && res.StatusCode == 404 {
			// delete the repository automatically and all its associated bounties
			// as it no longer exists
			if err := rc.Delete(repo.ID); err != nil {
				return err
			}
			return err
		}
		return err
	}

	// check that issues are enabled
	if !ghRepo.GetHasIssues() {
		return ErrIssuesDeactivated
	}

	mut := bson.D{{"$set", bson.D{
		{"owner", strings.TrimSpace(ghRepo.GetOwner().GetLogin())},
		{"name", strings.TrimSpace(ghRepo.GetName())},
		{"url", strings.TrimSpace(ghRepo.GetHTMLURL())},
		{"description", strings.TrimSpace(ghRepo.GetDescription())},
		{"model.updated_on", time.Now()},
	}}}

	_, err = rc.Coll.UpdateOne(DefaultCtx(), bson.D{{"_id", repo.ID}}, mut)
	return errors.Wrapf(err, "(repo) couldn't update repo '%d'", repo.ID)
}

func (rc *RepoCtrl) AddViaURL(url string) (*models.Repository, error) {

	owner, name, err := misc.ExtractOwnerAndNameFromGitHubURL(url)
	if err != nil {
		return nil, err
	}

	return rc.Add(owner, name)
}

func (rc *RepoCtrl) Delete(id int64) error {
	repo, err := rc.GetByID(id)
	if err != nil {
		return err
	}

	bounties, err := rc.BountyCtrl.GetOfRepository(repo.Owner, repo.Name)
	if err != nil {
		return err
	}

	for i := range bounties {
		if err := rc.BountyCtrl.Delete(bounties[i].ID, repo); err != nil {
			rc.logger.Error(fmt.Sprintf("couldn't delete associated bounty: %s", err.Error()))
		}
	}

	if _, err := rc.Coll.DeleteOne(DefaultCtx(), bson.D{{"_id", id}}); err != nil {
		return errors.Wrapf(err, "(repo) couldn't delete repo '%s'", id)
	}
	_, err = rc.DelColl.InsertOne(DefaultCtx(), models.DeletedModel{Object: repo})
	return errors.Wrapf(err, "(repo) couldn't move repo '%s' to deleted collection", id)
}
