package templatehelpers

import (
	"html/template"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNumericDuration(t *testing.T) {
	assert.InDelta(t, 1, NumericDuration(time.Second), 0)
	assert.InDelta(t, 3600, NumericDuration(time.Hour), 0)
}

func TestCountryCodeToFlag(t *testing.T) {
	assert.Equal(t, "🇺🇦", CountryCodeToFlag("UA"))
	assert.Equal(t, "🇧🇪", CountryCodeToFlag("BE"))
}

func TestHumanDistanceKM(t *testing.T) {
	assert.Equal(t, "0.00 km", HumanDistanceKM(1.23))
	assert.Equal(t, "1.23 km", HumanDistanceKM(1234))
	assert.Equal(t, "1234.57 km", HumanDistanceKM(1234567))
}

func TestHumanDistance(t *testing.T) {
	assert.Equal(t, "0.00 km", HumanDistanceKM(1.23))
	assert.Equal(t, "1.23 km", HumanDistanceKM(1234))
	assert.Equal(t, "1234.57 km", HumanDistanceKM(1234567))
}

func TestHumanSpeedKPH(t *testing.T) {
	assert.Equal(t, "4.43 km/h", HumanSpeedKPH(1.23))
	assert.Equal(t, "10.01 km/h", HumanSpeedKPH(2.78))
	assert.Equal(t, "17.96 km/h", HumanSpeedKPH(4.99))
}

func TestHumanTempoKM(t *testing.T) {
	assert.Equal(t, "13:33 min/km", HumanTempoKM(1.23))
	assert.Equal(t, "5:59 min/km", HumanTempoKM(2.78))
	assert.Equal(t, "3:20 min/km", HumanTempoKM(4.99))
	assert.Equal(t, "5:01 min/km", HumanTempoKM(3.32))
}

func TestBoolToHTML(t *testing.T) {
	assert.Equal(t, template.HTML("<i class=\"text-green-500 fas fa-check\"></i>"), BoolToHTML(true))
	assert.Equal(t, template.HTML("<i class=\"text-rose-500 fas fa-times\"></i>"), BoolToHTML(false))
}

func TestBoolToCheckbox(t *testing.T) {
	assert.Equal(t, template.HTML("checked"), BoolToCheckbox(true))
	assert.Equal(t, template.HTML(""), BoolToCheckbox(false))
}

func TestBuildDecoratedAttribute(t *testing.T) {
	r := BuildDecoratedAttribute("the-icon", "the-name", "the-value")

	assert.NotNil(t, r)
	assert.Equal(t, "the-icon", r.Icon)
	assert.Equal(t, "the-name", r.Name)
	assert.Equal(t, "the-value", r.Value)
}
