package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/mattn/go-isatty"
)

// Format is an output encoding.
type Format string

const (
	FormatAuto  Format = "auto"
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatJSONL Format = "jsonl"
	FormatCSV   Format = "csv"
	FormatTSV   Format = "tsv"
	FormatURL   Format = "url"
	FormatRaw   Format = "raw"
)

// Row is one output record: an ordered set of named columns plus the original
// value (used by json/jsonl and templates).
type Row struct {
	Cols  []string
	Vals  []string
	Value any
}

// Output renders rows in the selected format. A single Output instance handles a
// whole command run, so streaming formats can write incrementally.
type Output struct {
	format   Format
	fields   []string
	noHeader bool
	urlField string // which column the url format prints
	template *template.Template
	w        io.Writer

	tw         *tabwriter.Writer
	csvw       *csv.Writer
	headerDone bool
	jsonFirst  bool
	jsonOpen   bool
}

func newOutput(g *globalFlags) *Output {
	o := &Output{w: cmdOut, noHeader: g.noHeader, urlField: "url"}
	o.format = resolveFormat(g.output)
	if g.fields != "" {
		o.fields = splitComma(g.fields)
	}
	if g.template != "" {
		o.template = template.Must(template.New("row").Parse(g.template + "\n"))
		o.format = FormatRaw
	}
	return o
}

func resolveFormat(s string) Format {
	switch Format(s) {
	case FormatAuto, "":
		if isatty.IsTerminal(os.Stdout.Fd()) {
			return FormatTable
		}
		return FormatJSONL
	default:
		return Format(s)
	}
}

// Format returns the resolved output format.
func (o *Output) Format() Format { return o.format }

// SetURLField selects which column the url format emits (default "url").
func (o *Output) SetURLField(f string) { o.urlField = f }

// Emit renders one row.
func (o *Output) Emit(r Row) error {
	cols, vals := o.project(r)
	switch o.format {
	case FormatTable:
		return o.emitTable(cols, vals)
	case FormatCSV, FormatTSV:
		return o.emitCSV(cols, vals)
	case FormatJSONL:
		return o.emitJSONL(r.Value)
	case FormatJSON:
		return o.emitJSON(r.Value)
	case FormatURL:
		return o.emitField(r, o.urlField)
	case FormatRaw:
		if o.template != nil {
			return o.template.Execute(o.w, templateValue(r.Value))
		}
		return o.emitField(r, "")
	default:
		return o.emitJSONL(r.Value)
	}
}

func (o *Output) project(r Row) (cols, vals []string) {
	if len(o.fields) == 0 {
		return r.Cols, r.Vals
	}
	idx := map[string]int{}
	for i, c := range r.Cols {
		idx[c] = i
	}
	for _, f := range o.fields {
		cols = append(cols, f)
		switch {
		case idxHas(idx, f, len(r.Vals)):
			vals = append(vals, r.Vals[idx[f]])
		default:
			// Fall back to the row's underlying value so --fields can project any
			// field present in the full record, not just the curated columns.
			if s, ok := rowFieldString(r.Value, f); ok {
				vals = append(vals, s)
			} else {
				vals = append(vals, "")
			}
		}
	}
	return cols, vals
}

func idxHas(idx map[string]int, key string, n int) bool {
	i, ok := idx[key]
	return ok && i < n
}

// rowFieldString looks a column up in a row's underlying value (a decoded
// map or a templater), coercing the result to a display string. It is what lets
// --fields reach fields that are not in the curated column set.
func rowFieldString(v any, key string) (string, bool) {
	var m map[string]any
	switch t := v.(type) {
	case map[string]any:
		m = t
	case templater:
		m, _ = t.TemplateValue().(map[string]any)
	}
	if m == nil {
		return "", false
	}
	x, ok := m[key]
	if !ok {
		return "", false
	}
	return scalarString(x), true
}

// scalarString renders a decoded JSON value as a single display string.
func scalarString(x any) string {
	switch t := x.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'g', -1, 64)
	case []any:
		parts := make([]string, len(t))
		for i, e := range t {
			parts[i] = scalarString(e)
		}
		return strings.Join(parts, "; ")
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func (o *Output) emitTable(cols, vals []string) error {
	if o.tw == nil {
		o.tw = tabwriter.NewWriter(o.w, 0, 0, 2, ' ', 0)
	}
	if !o.headerDone && !o.noHeader {
		if _, err := fmt.Fprintln(o.tw, strings.Join(upperAll(cols), "\t")); err != nil {
			return err
		}
		o.headerDone = true
	}
	_, err := fmt.Fprintln(o.tw, strings.Join(vals, "\t"))
	return err
}

func (o *Output) emitCSV(cols, vals []string) error {
	if o.csvw == nil {
		o.csvw = csv.NewWriter(o.w)
		if o.format == FormatTSV {
			o.csvw.Comma = '\t'
		}
	}
	if !o.headerDone && !o.noHeader {
		if err := o.csvw.Write(cols); err != nil {
			return err
		}
		o.headerDone = true
	}
	return o.csvw.Write(vals)
}

func (o *Output) emitJSONL(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(o.w, string(b))
	return err
}

func (o *Output) emitJSON(v any) error {
	if !o.jsonOpen {
		if _, err := fmt.Fprint(o.w, "["); err != nil {
			return err
		}
		o.jsonOpen = true
		o.jsonFirst = true
	}
	if !o.jsonFirst {
		if _, err := fmt.Fprint(o.w, ","); err != nil {
			return err
		}
	}
	o.jsonFirst = false
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(o.w, "\n  "+string(b))
	return err
}

func (o *Output) emitField(r Row, field string) error {
	if field == "" && len(r.Vals) > 0 {
		_, err := fmt.Fprintln(o.w, r.Vals[0])
		return err
	}
	for i, c := range r.Cols {
		if c == field && i < len(r.Vals) {
			_, err := fmt.Fprintln(o.w, r.Vals[i])
			return err
		}
	}
	return nil
}

// Flush finalises buffered formats. Call once at the end of a command.
func (o *Output) Flush() error {
	if o.tw != nil {
		return o.tw.Flush()
	}
	if o.csvw != nil {
		o.csvw.Flush()
		return o.csvw.Error()
	}
	if o.jsonOpen {
		_, err := fmt.Fprintln(o.w, "\n]")
		return err
	}
	return nil
}

// Raw writes bytes straight to the output (for --raw/page text/file bodies).
func (o *Output) Raw(b []byte) error {
	_, err := o.w.Write(b)
	return err
}

// templater lets a value expose a decoded, template-friendly view of itself
// while keeping its raw form for JSON output (where original types matter).
type templater interface{ TemplateValue() any }

// templateValue returns the value a Go template should range over: the decoded
// view when the value provides one, otherwise the value itself.
func templateValue(v any) any {
	if t, ok := v.(templater); ok {
		return t.TemplateValue()
	}
	return v
}

func splitComma(s string) []string {
	var out []string
	for p := range strings.SplitSeq(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func upperAll(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToUpper(s)
	}
	return out
}
