package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/gormstore"
	"github.com/alexedwards/scs/v2"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	slogecho "github.com/samber/slog-echo"

	session "github.com/spazzymoto/echo-scs-session"
)

func newEcho(logger *slog.Logger) *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(slogecho.New(logger.With("module", "webserver")))
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())
	e.Pre(middleware.RemoveTrailingSlash())

	return e
}

func (a *App) ConfigureWebserver() error {
	var err error

	e := newEcho(a.rawLogger)
	e.Debug = a.Config.Debug

	a.sessionManager = scs.New()
	a.sessionManager.Cookie.Path = "/"
	a.sessionManager.Cookie.HttpOnly = true
	a.sessionManager.Lifetime = 24 * time.Hour

	if a.sessionManager.Store, err = gormstore.New(a.db); err != nil {
		return err
	}

	e.Use(session.LoadAndSave(a.sessionManager))

	e.Renderer = &Template{
		app:       a,
		templates: a.parseViewTemplates(),
	}

	publicGroup := e.Group("")

	publicGroup.StaticFS("/assets", a.Assets)

	publicGroup.GET("/assets", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, a.echo.Reverse("dashboard"))
	}).Name = "assets"

	userGroup := publicGroup.Group("/user")
	userGroup.GET("/signin", a.userLoginHandler).Name = "user-login"
	userGroup.POST("/signin", a.userSigninHandler).Name = "user-signin"
	userGroup.POST("/register", a.userRegisterHandler).Name = "user-register"
	userGroup.GET("/signout", a.userSignoutHandler).Name = "user-signout"

	sec := a.secureRoutes(publicGroup)
	a.adminRoutes(sec)

	a.echo = e

	return nil
}

func (a *App) ValidateAdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		u := a.getCurrentUser(ctx)
		if u == nil || !u.IsActive() {
			log.Warn("User is not found")
			return ctx.Redirect(http.StatusFound, a.echo.Reverse("user-signout"))
		}

		if !u.Admin {
			log.Warn("User is not an admin")
			return ctx.Redirect(http.StatusFound, a.echo.Reverse("dashboard"))
		}

		return next(ctx)
	}
}

func (a *App) ValidateUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if err := a.setUser(ctx); err != nil {
			log.Warn(err.Error())
			return ctx.Redirect(http.StatusFound, a.echo.Reverse("user-signout"))
		}

		return next(ctx)
	}
}

func (a *App) secureRoutes(e *echo.Group) *echo.Group {
	secureGroup := e.Group("")

	secureGroup.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:  a.jwtSecret(),
		TokenLookup: "cookie:token",
		ErrorHandler: func(c echo.Context, err error) error {
			log.Warn(err.Error())
			return c.Redirect(http.StatusFound, a.echo.Reverse("user-signout"))
		},
	}))
	secureGroup.Use(a.ValidateUserMiddleware)

	secureGroup.GET("/", a.dashboardHandler).Name = "dashboard"
	secureGroup.GET("/statistics", a.statisticsHandler).Name = "statistics"

	selfGroup := secureGroup.Group("/user")
	selfGroup.GET("/profile", a.userProfileHandler).Name = "user-profile"
	selfGroup.POST("/profile", a.userProfileUpdateHandler).Name = "user-profile-update"
	selfGroup.POST("/refresh", a.userRefreshHandler).Name = "user-refresh"
	selfGroup.POST("/reset-api-key", a.userProfileResetAPIKeyHandler).Name = "user-profile-reset-api-key"

	usersGroup := secureGroup.Group("/users")
	usersGroup.GET("/:id", a.userShowHandler).Name = "user-show"

	workoutsGroup := secureGroup.Group("/workouts")
	workoutsGroup.GET("", a.workoutsHandler).Name = "workouts"
	workoutsGroup.POST("", a.addWorkout).Name = "workouts-create"
	workoutsGroup.GET("/:id", a.workoutsShowHandler).Name = "workout-show"
	workoutsGroup.POST("/:id", a.workoutsUpdateHandler).Name = "workout-update"
	workoutsGroup.GET("/:id/edit", a.workoutsEditHandler).Name = "workout-edit"
	workoutsGroup.POST("/:id/delete", a.workoutsDeleteHandler).Name = "workout-delete"
	workoutsGroup.POST("/:id/refresh", a.workoutsRefreshHandler).Name = "workout-refresh"
	workoutsGroup.GET("/add", a.workoutsAddHandler).Name = "workout-add"

	return secureGroup
}

func (a *App) adminRoutes(e *echo.Group) *echo.Group {
	adminGroup := e.Group("/admin")
	adminGroup.Use(a.ValidateAdminMiddleware)

	adminGroup.GET("", a.adminRootHandler).Name = "admin"

	adminUsersGroup := adminGroup.Group("/users")
	adminUsersGroup.GET("/:id/edit", a.adminUserEditHandler).Name = "admin-user-edit"
	adminUsersGroup.POST("/:id", a.adminUserUpdateHandler).Name = "admin-user-update"
	adminUsersGroup.POST("/:id/delete", a.adminUserDeleteHandler).Name = "admin-user-delete"
	adminUsersGroup.GET("/:id", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, a.echo.Reverse("admin-user-edit", c.Param("id")))
	}).Name = "admin-user-show"

	return adminGroup
}
