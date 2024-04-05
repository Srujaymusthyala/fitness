package app

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/jovandeginste/workout-tracker/pkg/database"
	"github.com/jovandeginste/workout-tracker/pkg/templatehelpers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/vorlif/spreak/humanize"
	"golang.org/x/text/language"
)

type Template struct {
	app       *App
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	r, err := t.templates.Clone()
	if err != nil {
		return err
	}

	l := langFromBrowser(ctx)
	tr := t.app.translatorFromContext(ctx)
	h := t.app.humanizerFromContext(ctx)
	u := t.app.getCurrentUser(ctx)

	units := u.Units()
	if units == templatehelpers.BrowserUnits {
		units = templatehelpers.UnitsFromBrowserLanguage(l)
	}

	r.Funcs(template.FuncMap{
		"i18n":         tr.Getf,
		"language":     tr.Language().String,
		"humanizer":    func() *humanize.Humanizer { return h },
		"RelativeDate": h.NaturalTime,
		"CurrentUser":  func() *database.User { return u },
		"LocalTime":    func(t time.Time) time.Time { return t.In(u.Timezone()) },
		"LocalDate":    func(t time.Time) string { return t.In(u.Timezone()).Format("2006-01-02 15:04") },

		"HumanDistance": templatehelpers.HumanDistance(units),
		"HumanSpeed":    templatehelpers.HumanSpeed(units),
		"HumanTempo":    templatehelpers.HumanTempo(units),
	})

	return r.ExecuteTemplate(w, name, data)
}

func echoFunc(key string, _ ...interface{}) string {
	return key
}

func (a *App) viewTemplateFunctions() template.FuncMap {
	h := a.humanizer.CreateHumanizer(language.English)

	return template.FuncMap{
		"i18n":        echoFunc,
		"language":    func() string { return BrowserLanguage },
		"humanizer":   func() *humanize.Humanizer { return h },
		"CurrentUser": func() *database.User { return nil },
		"LocalTime":   func(t time.Time) time.Time { return t.UTC() },
		"LocalDate":   func(t time.Time) string { return t.UTC().Format("2006-01-02 15:04") },

		"supportedUnits":     templatehelpers.SupportedUnits,
		"supportedLanguages": a.translator.SupportedLanguages,
		"workoutTypes":       database.WorkoutTypes,

		"NumericDuration":         templatehelpers.NumericDuration,
		"CountryCodeToFlag":       templatehelpers.CountryCodeToFlag,
		"ToKilometer":             templatehelpers.ToKilometer,
		"HumanDistance":           templatehelpers.HumanDistance(templatehelpers.MetricUnits),
		"HumanSpeed":              templatehelpers.HumanSpeed(templatehelpers.MetricUnits),
		"HumanTempo":              templatehelpers.HumanTempo(templatehelpers.MetricUnits),
		"HumanDuration":           templatehelpers.HumanDuration,
		"IconFor":                 templatehelpers.IconFor,
		"BoolToHTML":              templatehelpers.BoolToHTML,
		"BoolToCheckbox":          templatehelpers.BoolToCheckbox,
		"BuildDecoratedAttribute": templatehelpers.BuildDecoratedAttribute,
		"ToLanguageInformation":   templatehelpers.ToLanguageInformation,
		"Timezones":               templatehelpers.Timezones,
		"LocalUnit":               templatehelpers.LocalUnit,

		"RelativeDate": h.NaturalTime,

		"RouteFor": func(name string, params ...interface{}) string {
			rev := a.echo.Reverse(name, params...)
			if rev == "" {
				return "/invalid/route/#" + name
			}

			return rev
		},
	}
}

func (a *App) parseViewTemplates() *template.Template {
	templ := template.New("").Funcs(a.viewTemplateFunctions())
	if a.Views == nil {
		return templ
	}

	err := fs.WalkDir(a.Views, ".", func(path string, d fs.DirEntry, err error) error {
		if d != nil && d.IsDir() {
			return err
		}

		if strings.HasSuffix(path, ".html") {
			if _, myErr := templ.ParseFS(a.Views, path); err != nil {
				a.logger.Warn(fmt.Sprintf("Error loading template: %v", myErr))
				return myErr
			}
		}

		return err
	})
	if err != nil {
		a.logger.Warn(fmt.Sprintf("Error loading template: %v", err))
		log.Warn(fmt.Sprintf("Error loading template: %v", err))
	}

	return templ
}
