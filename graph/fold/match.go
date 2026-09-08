package fold

import (
	"math"
	"strconv"
)

// FoldShape is true when rows look like FoldCypher / GET /fold output.
func FoldShape(rows []map[string]any) bool {
	if len(rows) == 0 {
		return false
	}
	_, company := rows[0]["company"]
	_, ticker := rows[0]["ticker_now"]
	_, qty := rows[0]["qty_now"]
	return company || ticker || qty
}

// GoldFold is the six Notion schema rows Gap A must confirm on a live cluster.
func GoldFold() []Expectation {
	keep := map[string]bool{"nflx": true, "mnts": true, "apge": true, "lpsn": true, "square": true, "ftel": true}
	var out []Expectation
	for _, want := range Expectations() {
		if keep[want.Q] {
			out = append(out, want)
		}
	}
	return out
}

// MatchRows compares one FoldCypher / HTTP fold row set to a Notion gold expectation.
func MatchRows(rows []map[string]any, want Expectation) Check {
	check := Check{Expectation: want}
	if len(rows) == 0 {
		check.Error = "unresolved"
		return check
	}
	row := rows[0]
	check.GotQty = as_float(row["qty_now"])
	check.GotCash = as_float(first(row["cash_received"], row["cash"]))
	company := as_string(row["company"])
	ticker := as_string(first(row["ticker_now"], row["ticker"]))
	series := series_len(row["series"])
	if company != want.Company || ticker != want.Ticker {
		check.Error = "identity"
	} else if want.Series > 0 && series > 0 && series < want.Series {
		check.Error = "series"
	} else if math.Abs(check.GotQty-want.QtyNow) > 0.005 {
		check.Error = "qty"
	} else if math.Abs(check.GotCash-want.Cash) > 0.02 {
		check.Error = "cash"
	} else {
		check.OK = true
	}
	return check
}

func first(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func as_string(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func as_float(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case int32:
		return float64(typed)
	case json_number:
		f, _ := typed.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(typed, 64)
		return f
	default:
		return 0
	}
}

type json_number interface {
	Float64() (float64, error)
}

func series_len(value any) int {
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	default:
		return 0
	}
}
