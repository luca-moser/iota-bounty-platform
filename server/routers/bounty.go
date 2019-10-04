package routers

import (
	"github.com/labstack/echo/middleware"
	"github.com/luca-moser/iota-bounty-platform/server/controllers"
	"github.com/luca-moser/iota-bounty-platform/server/server/config"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

type BountyRouter struct {
	R         *echo.Echo              `inject:""`
	BC        *controllers.BountyCtrl `inject:""`
	Dev       bool                    `inject:"dev"`
	Config    *config.Configuration   `inject:""`
	JWTConfig middleware.JWTConfig    `inject:"jwt_config_user"`
}

func (br *BountyRouter) Init() {

	routeGroup := br.R.Group("/api/bounties")

	routeGroup.GET("/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return err
		}

		repo, err := br.BC.GetByID(id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, repo)
	})

	routeGroup.GET("/:owner/:name", func(c echo.Context) error {
		owner := c.Param("owner")
		name := c.Param("name")

		bounties, err := br.BC.GetOfRepository(owner, name)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, bounties)
	})

	routeGroup.POST("", func(c echo.Context) error {
		issueIDStr := c.QueryParam("issue_id")
		owner := c.QueryParam("owner")
		name := c.QueryParam("name")
		issueID, err := strconv.Atoi(issueIDStr)
		if err != nil {
			return err
		}

		bounty, err := br.BC.Add(owner, name, issueID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, bounty)
	})

	routeGroup.DELETE("/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}
		if err := br.BC.Delete(int64(id)); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, SimpleMsg{"ok"})
	})

}
