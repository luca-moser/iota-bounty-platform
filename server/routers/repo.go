package routers

import (
	"github.com/luca-moser/iota-bounty-platform/server/controllers"
	"github.com/luca-moser/iota-bounty-platform/server/server/config"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

type RepoRouter struct {
	R      *echo.Echo              `inject:""`
	RC     *controllers.RepoCtrl   `inject:""`
	BC     *controllers.BountyCtrl `inject:""`
	Dev    bool                    `inject:"dev"`
	Config *config.Configuration   `inject:""`
}

func (rr *RepoRouter) Init() {

	routeGroup := rr.R.Group("/api/repos")

	routeGroup.GET("", func(c echo.Context) error {
		repos, err := rr.RC.GetAll()
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, repos)
	})

	routeGroup.GET("/of/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return err
		}

		bounty, err := rr.BC.GetByID(int64(id))
		if err != nil {
			return err
		}

		repo, err := rr.RC.GetByID(bounty.RepositoryID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, repo)
	})

	routeGroup.GET("/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return err
		}

		repo, err := rr.RC.GetByID(int64(id))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, repo)
	})

	routeGroup.GET("/:owner/:name", func(c echo.Context) error {
		owner := c.Param("owner")
		name := c.Param("name")

		repo, err := rr.RC.GetByOwnerAndName(owner, name)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, repo)
	})

	routeGroup.POST("", func(c echo.Context) error {
		url := c.QueryParam("url")
		if url == "" {
			return ErrBadRequest
		}

		repo, err := rr.RC.AddViaURL(url)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, repo)
	})

	routeGroup.POST("/:owner/:name", func(c echo.Context) error {
		owner := c.Param("owner")
		name := c.Param("name")

		repo, err := rr.RC.Add(owner, name)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, repo)
	})

	routeGroup.DELETE("/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return err
		}

		if err := rr.RC.Delete(int64(id)); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, SimpleMsg{"ok"})
	})

}
