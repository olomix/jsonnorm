package jsonnorm

import (
	"fmt"
	"time"

	"github.com/spyzhov/ajson"
)

type Rule interface {
	Paths() []string
	Apply(*ajson.Node) error
}

// TimeRule specifies how to normalize json document. If we find nodes by
// jsonpath, we will replace
type TimeRule struct {
	JSONPaths []string
	// If TZ is not nil, then we will normalize JSON timezone to TZ.
	TZ *time.Location
	// If JSONPath's specified time is in range from Time - PeriodBefore to
	// Time + PeriodAfter, then we will replace json value with Time.
	// If Time is zero, then we will not replace json value.
	Time         time.Time
	PeriodBefore time.Duration
	PeriodAfter  time.Duration
	// Time format layout. If not specified, then time.RFC3339
	Layout string
}

func (tr TimeRule) Paths() []string { return tr.JSONPaths }

func (tr TimeRule) Apply(node *ajson.Node) error {
	tmStr, err := node.GetString()
	if err != nil {
		return fmt.Errorf("cannot get string from node: %w", err)
	}

	tm, err := time.Parse(time.RFC3339Nano, tmStr)
	if err != nil {
		return fmt.Errorf("cannot parse time from node: %w", err)
	}

	doReplace := false

	if !tr.Time.IsZero() && !tr.Time.Equal(tm) {
		if tm.After(tr.Time.Add(-tr.PeriodBefore)) &&
			tm.Before(tr.Time.Add(tr.PeriodAfter)) {

			tm = tr.Time
			doReplace = true
		}
	}

	if tr.TZ != nil {
		tm = tm.In(tr.TZ)
		doReplace = true
	}

	if doReplace {
		err = node.SetString(tm.In(tr.TZ).Format(tr.layout()))
		if err != nil {
			return fmt.Errorf("cannot set string to node: %w", err)
		}
	}

	return nil
}

func (tr TimeRule) layout() string {
	if tr.Layout == "" {
		return time.RFC3339
	}
	return tr.Layout
}

type Normalizer struct {
	Rules []Rule
	Doc   string
}

func NewNormalizer(doc string, rules ...Rule) *Normalizer {
	return &Normalizer{
		Rules: rules,
		Doc:   doc,
	}
}

func (n *Normalizer) Apply() (string, error) {
	node, err := ajson.Unmarshal([]byte(n.Doc))
	if err != nil {
		return "", err
	}

	for _, rule := range n.Rules {
		jsonPaths := rule.Paths()
		for _, jsonPath := range jsonPaths {
			ns, err := node.JSONPath(jsonPath)
			if err != nil {
				return "", err
			}
			for _, ni := range ns {
				err = rule.Apply(ni)
				if err != nil {
					return "", err
				}
			}
		}
	}

	return node.String(), nil
}

func (n *Normalizer) AddRule(r Rule) *Normalizer {
	n.Rules = append(n.Rules, r)
	return n
}
