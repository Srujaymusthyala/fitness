package app

import (
	"github.com/jovandeginste/workout-tracker/pkg/database"
	"github.com/labstack/echo/v4"
	"github.com/vorlif/spreak"
	"github.com/vorlif/spreak/humanize"
	"github.com/vorlif/spreak/humanize/locale/de"
	"github.com/vorlif/spreak/humanize/locale/nl"
	"golang.org/x/text/language"
)

const (
	BrowserLanguage   = "browser"
	DefaultTotalsShow = database.WorkoutTypeRunning
)

func (a *App) ConfigureLocalizer() error {
	var domain spreak.FsOption

	if a.Translations != nil {
		domain = spreak.WithFs(a.Translations)
	} else {
		domain = spreak.WithPath(".")
	}

	bundle, err := spreak.NewBundle(
		// Set the language used in the program code/templates
		spreak.WithSourceLanguage(language.English),
		// Set the path from which the translations should be loaded
		spreak.WithFilesystemLoader(spreak.NoDomain, domain),
		// Specify the languages you want to load
		spreak.WithLanguage(translations()...),
	)
	if err != nil {
		return err
	}

	a.translator = bundle

	a.humanizer = humanize.MustNew(
		humanize.WithLocale(humanLocales()...),
	)

	return nil
}

func translations() []interface{} {
	return []interface{}{
		language.English,
		language.Dutch,
		language.German,
	}
}

func humanLocales() []*humanize.LocaleData {
	return []*humanize.LocaleData{
		nl.New(),
		de.New(),
	}
}

func langFromContext(ctx echo.Context) []interface{} {
	return []interface{}{
		ctx.QueryParam("lang"),
		ctx.Get("user_language"),
		langFromBrowser(ctx),
	}
}

func langFromBrowser(ctx echo.Context) string {
	return ctx.Request().Header.Get("Accept-Language")
}

func (a *App) i18n(ctx echo.Context, message string, vars ...interface{}) string {
	return a.translatorFromContext(ctx).Getf(message, vars...)
}

func (a *App) translatorFromContext(ctx echo.Context) *spreak.Localizer {
	return spreak.NewLocalizer(a.translator, langFromContext(ctx)...)
}

func (a *App) humanizerFromContext(ctx echo.Context) *humanize.Humanizer {
	return a.humanizer.CreateHumanizer(langFromContext(ctx)...)
}
