package merged

import (
	"regexp"
	"strings"

	"github.com/vatsimnerd/util/set"
)

const (
	runwayIdentExpr = `(\d{2}(?:[LRC]|\s(?:LEFT|RIGHT|CENTER))?)(?:\s(?:(?:AND|OR)\s)?(\d{2}(?:[LRC]|\s(?:LEFT|RIGHT|CENTER))?))?(?:\s(?:(?:AND|OR)\s)?(\d{2}(?:[LRC]|\s(?:LEFT|RIGHT|CENTER))?))?`
)

var (
	arrivalExpressions = []*regexp.Regexp{
		regexp.MustCompile(
			`(?:(?:APPROACH|ARRIVAL|LANDING|LDG)\s)+(?:RUNWAY|RWY)S?\s` +
				runwayIdentExpr,
		),
		regexp.MustCompile(
			`(?:RUNWAY|RWY)S?\s` +
				runwayIdentExpr +
				`\sFOR\s(?:ARRIVAL|LANDING|LDG|APPROACH)`,
		),
		regexp.MustCompile(
			`(?:RUNWAY|RWY)S?\s` + runwayIdentExpr + `\sIN\sUSE`,
		),
		regexp.MustCompile(
			`(?:RUNWAY|RWY)S?\sIN\sUSE\s` + runwayIdentExpr,
		),
		regexp.MustCompile(
			`(?:APPROACH|ARRIVAL|LANDING|LDG)\sAND\s(?:TAKEOFF|DEPARTURE|DEPARTING|DEP)\s(?:RUNWAY|RWY)S?\s` +
				runwayIdentExpr,
		),
	}

	departureExpressions = []*regexp.Regexp{
		regexp.MustCompile(
			`(?:TAKEOFF|DEPARTURE|DEPARTING|DEP)\s(?:RUNWAY|RWY)S?\s` + runwayIdentExpr,
		),
		regexp.MustCompile(
			`(?:RUNWAY|RWY)S?\s` + runwayIdentExpr + `\sFOR\s(?:TAKEOFF|DEPARTURE|DEP)`,
		),
		regexp.MustCompile(
			`(?:RUNWAY|RWY)S?\s` + runwayIdentExpr + `\sIN\sUSE`,
		),
		regexp.MustCompile(
			`(?:RUNWAY|RWY)S?\sIN\sUSE\s` + runwayIdentExpr,
		),
		regexp.MustCompile(
			`(?:APPROACH|ARRIVAL|LANDING|LDG)\sAND\s(?:TAKEOFF|DEPARTURE|DEPARTING|DEP)\s(?:RUNWAY|RWY)S?\s` +
				runwayIdentExpr,
		),
	}

	exprSpecial    = regexp.MustCompile(`[^A-Z0-9\s]`)
	exprWhitespace = regexp.MustCompile(`\s+`)
)

func normalizeIdent(ident string) string {
	ident = regexp.MustCompile(`\s`).ReplaceAllString(ident, "")
	if len(ident) > 3 {
		ident = ident[0:3]
	}
	return ident
}

func normalizeAtisText(text string) string {
	text = strings.ToUpper(text)
	text = exprSpecial.ReplaceAllString(text, "")
	text = exprWhitespace.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func detectArrivalRunways(atisText string) *set.Set[string] {
	if atisText != "" {
		for _, re := range arrivalExpressions {
			match := re.FindAllStringSubmatch(atisText, -1)
			if len(match) > 0 {
				results := set.New[string]()
				for _, m := range match[0][1:] {
					if m != "" {
						results.Add(normalizeIdent(m))
					}
				}
				return results
			}
		}
	}
	return nil
}

func detectDepartureRunways(atisText string) *set.Set[string] {
	if atisText != "" {
		for _, re := range departureExpressions {
			match := re.FindAllStringSubmatch(atisText, -1)
			if len(match) > 0 {
				results := set.New[string]()
				for _, m := range match[0][1:] {
					if m != "" {
						results.Add(normalizeIdent(m))
					}
				}
				return results
			}
		}
	}
	return nil
}

func (a *Airport) setActiveRunways() {
	if a.Controllers.ATIS == nil {
		for _, rwy := range a.Runways {
			rwy.ActiveLnd = false
			rwy.ActiveTO = false
		}
		return
	}

	atisText := normalizeAtisText(a.Controllers.ATIS.TextAtis)

	runways := detectArrivalRunways(atisText)
	for ident, rwy := range a.Runways {
		rwy.ActiveLnd = runways.Has(ident)
	}

	runways = detectDepartureRunways(atisText)
	for ident, rwy := range a.Runways {
		rwy.ActiveTO = runways.Has(ident)
	}
}