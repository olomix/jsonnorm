package jsonnorm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizer_Apply(t *testing.T) {
	doc := `{
    "campaigns": [
        {
            "status_modified": "2012-01-04T00:00:00+02:00",
            "date_modified": "2012-01-04T00:00:00+02:00",
            "date_created": "2012-01-03T00:00:00+02:00",
            "status": 0,
            "period_end": "2012-01-02T00:00:00+02:00",
            "period_start": "2012-01-01T00:00:00+02:00",
            "campaign_id": 100
        },
        {
            "status_modified": "2012-01-03T23:59:55+02:00",
            "date_modified": "2012-01-04T00:00:00+02:00",
            "date_created": "2012-01-03T00:00:00+02:00",
            "status": 0,
            "period_end": "2012-01-02T00:00:00+02:00",
            "period_start": "2012-01-01T00:00:00+02:00",
            "campaign_id": 100
        }
    ]
}`

	kyivTZ, err := time.LoadLocation("Europe/Kiev")
	require.NoError(t, err)
	norm := NewNormalizer(doc, TimeRule{
		JSONPaths: []string{
			`$.campaigns[*].['period_end','period_start']`,
			`$.campaigns[*].['date_created']`,
		},
		TZ: time.UTC,
	})
	norm.AddRule(TimeRule{
		JSONPaths:    []string{`$..['status_modified']`},
		TZ:           kyivTZ,
		Time:         time.Date(2012, 1, 3, 22, 0, 0, 0, time.UTC),
		PeriodBefore: 10 * time.Second,
	})

	want := `{
  "campaigns": [
    {
      "status_modified": "2012-01-04T00:00:00+02:00",
      "date_modified": "2012-01-04T00:00:00+02:00",
      "date_created": "2012-01-02T22:00:00Z",
      "status": 0,
      "period_end": "2012-01-01T22:00:00Z",
      "period_start": "2011-12-31T22:00:00Z",
      "campaign_id": 100
    },
    {
      "status_modified": "2012-01-04T00:00:00+02:00",
      "date_modified": "2012-01-04T00:00:00+02:00",
      "date_created": "2012-01-02T22:00:00Z",
      "status": 0,
      "period_end": "2012-01-01T22:00:00Z",
      "period_start": "2011-12-31T22:00:00Z",
      "campaign_id": 100
    }
  ]
}`
	result, err := norm.Apply()
	require.NoError(t, err)
	require.JSONEq(t, want, result)
}
