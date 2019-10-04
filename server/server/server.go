package server

import (
	"context"
	"fmt"
	"github.com/dpapathanasiou/go-recaptcha"
	"github.com/facebookgo/inject"
	"github.com/google/go-github/github"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/luca-moser/iota-bounty-platform/server/controllers"
	"github.com/luca-moser/iota-bounty-platform/server/misc"
	"github.com/luca-moser/iota-bounty-platform/server/models"
	"github.com/luca-moser/iota-bounty-platform/server/routers"
	"github.com/luca-moser/iota-bounty-platform/server/server/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"golang.org/x/oauth2"
	"gopkg.in/inconshreveable/log15.v2"
	"html/template"
	"io"
	"os"
	"time"
)

type TemplateRendered struct {
	templates *template.Template
}

func (t *TemplateRendered) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Server struct {
	Config    *config.Configuration
	WebEngine *echo.Echo
	Logger    log15.Logger
}

func (server *Server) Start() {
	start := time.Now().UnixNano()

	// load config
	conf, err := config.LoadConfig()
	must(err)
	server.Config = conf
	httpConfig := server.Config.HTTP

	// init logger
	misc.Debug = conf.Verbose
	logger, err := misc.GetLogger("app")
	must(err)
	logger.Info("booting up app...")
	server.Logger = logger

	// init web server
	e := echo.New()
	e.HideBanner = true
	server.WebEngine = e
	if httpConfig.LogRequests {
		requestLogFile, err := os.Create(fmt.Sprintf("./logs/requests.log"))
		if err != nil {
			panic(err)
		}
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Output: requestLogFile}))
		e.Logger.SetLevel(3)
	}

	// load html files
	e.Renderer = &TemplateRendered{
		templates: template.Must(template.ParseGlob(fmt.Sprintf("%s/*.html", httpConfig.Assets.HTML))),
	}

	// asset paths
	e.Static("/assets", httpConfig.Assets.Static)
	e.File("/favicon.ico", httpConfig.Assets.Favicon)

	// init github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.GitHub.AuthToken})
	githubClient := github.NewClient(oauth2.NewClient(ctx, ts))

	// verify that the connection to GitHub actually works and the user
	// is correctly authenticated
	zenMsg, _, err := githubClient.Zen(context.Background())
	must(err)

	ownUser, _, err := githubClient.Users.Get(controllers.DefaultCtx(), "")
	must(err)
	logger.Info(fmt.Sprintf("connected to GitHub as '%s'", ownUser.GetName()))
	logger.Info(fmt.Sprintf("GitHub Zen message: %s", zenMsg))

	// create controllers
	appCtrl := &controllers.AppCtrl{}
	userCtrl := &controllers.UserCtrl{}
	repoCtrl := &controllers.RepoCtrl{}
	bountyCtrl := &controllers.BountyCtrl{}
	bot := &controllers.Bot{}
	ctrls := []controllers.Controller{appCtrl, userCtrl, repoCtrl, bountyCtrl, bot}

	// create routers
	indexRouter := &routers.IndexRouter{}
	userRouter := &routers.UserRouter{}
	repoRouter := &routers.RepoRouter{}
	bountyRouter := &routers.BountyRouter{}
	rters := []routers.Router{indexRouter, userRouter, repoRouter, bountyRouter}

	// init mongo db conn
	mongoClient, err := mongo.NewClient([]*options.ClientOptions{
		{
			WriteConcern: writeconcern.New(writeconcern.J(true), writeconcern.WMajority(), writeconcern.WTimeout(5*time.Second)),
			ReadConcern:  readconcern.Majority(),
		},
		options.Client().ApplyURI(server.Config.DB.URI),
	}...)
	must(err)

	mongoConnCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	must(mongoClient.Connect(mongoConnCtx))
	must(mongoClient.Ping(mongoConnCtx, nil))
	logger.Info("connected to MongoDB")

	// load mail templates
	mailTemplates := template.Must(template.New("mails.html").ParseGlob("./mails.html"))

	// jwt
	authJWTConf := middleware.JWTConfig{
		Claims:     &models.UserJWTClaims{},
		SigningKey: []byte(conf.JWT.PrivateKey),
	}

	// recaptcha
	recaptcha.Init(conf.ReCaptcha.PrivateKey)

	// create injection graph for automatic dependency injection
	g := inject.Graph{}

	// add various objects to the graph
	must(g.Provide(
		&inject.Object{Value: e},
		&inject.Object{Value: mongoClient},
		&inject.Object{Value: githubClient},
		&inject.Object{Value: conf},
		&inject.Object{Value: conf.Dev, Name: "dev"},
		&inject.Object{Value: conf.ReCaptcha.PublicKey, Name: "recaptcha_public_key"},
		&inject.Object{Value: conf.ReCaptcha.PrivateKey, Name: "recaptcha_private_key"},
		&inject.Object{Value: authJWTConf, Name: "jwt_config_user"},
		&inject.Object{Value: mailTemplates, Name: "mail_templates"},
	))

	// add controllers to graph
	for _, controller := range ctrls {
		must(g.Provide(&inject.Object{Value: controller}))
	}

	// add routers to graph
	for _, router := range rters {
		must(g.Provide(&inject.Object{Value: router}))
	}

	// run dependency injection
	must(g.Populate())

	// init controllers
	for _, controller := range ctrls {
		must(controller.Init())
	}
	logger.Info("initialised controllers")

	// init routers
	for _, router := range rters {
		router.Init()
	}
	logger.Info("initialised routers")

	// boot up server
	go must(e.Start(httpConfig.ListenAddress))

	// finish
	delta := (time.Now().UnixNano() - start) / 1000000
	logger.Info(fmt.Sprintf("%s ready", conf.Name), "startup", delta)
}

func (server *Server) Shutdown(ctx context.Context) {
	server.Logger.Info("shutting down server...")
	must(server.WebEngine.Shutdown(ctx))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
