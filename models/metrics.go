package models

import (
	"encoding/json"
	"fmt"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
)

type (
	// IncomingMetric represents metrics as they come in through the API
	IncomingMetric struct {
		Key   string   `json:"key" form:"key" query:"key"`
		Value float64  `json:"value" form:"value" query:"value"`
		Tags  []string `json:"tags" form:"tags" query:"tags"`
	}
	// PlainMetric represents metrics after we've added a timestamp
	PlainMetric struct {
		Key       string   `json:"key" form:"key" query:"key"`
		Value     float64  `json:"value" form:"value" query:"value"`
		Timestamp int64    `json:"timestamp" form:"timestamp" query:"timestamp"`
		Tags      []string `json:"tags" form:"tags" query:"tags"`
	}
	// Metric is our fully-featured model. It is used to generate database tables, and is meant to be easily extensible over time.
	// If we were to use metrics in other parts of the application, we should use instances of Metric.
	Metric struct {
		ID        uint    `gorm:"primary_key"`
		Key       string  `gorm:"index;not null" json:"key"`
		Value     float64 `gorm:"not null" json:"value"`
		Timestamp int64   `gorm:"index;not null" json:"timestamp"` // store this as a number since the server time doesn't even necessary reflect timezone of origin. we can transform this later if needed.
		Tags      []Tag   `gorm:"many2many:metric_tags;" `
	}
	// Tag is our fully-featured tag model. At present, it is only used to generate database tables.
	// It is more than just a string because we could easily imagine wanting to extend this table with more metadata about each tag.
	Tag struct {
		ID   uint   `gorm:"primary_key"`
		Text string `gorm:"unique_index;not null" json:"text"`
		// We could also consider storing a first_seen time
	}

	// This is for our join table
	MetricTags struct {
		MetricID int
		TagID    int
	}
)

// validation for metrics coming in through the API
func (m IncomingMetric) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Key, validation.Required, validation.Length(1, 5000)),
		validation.Field(&m.Value, validation.Required),
		validation.Field(&m.Tags, validation.Length(0, 50)),
	)
}

// AddMetric inserts a metric into the database
func (db *DB) AddMetric(metric PlainMetric) (metricsInserted int64, err error) {
	var dbResult *gorm.DB

	if len(metric.Tags) == 0 {
		// If there are no tags, the insert is fairly simple
		dbResult = db.Exec(
			`INSERT INTO metrics("key", "value", "timestamp")
      VALUES(?,?,?)`, metric.Key, metric.Value, metric.Timestamp)
		fmt.Printf("bees")
	} else {
		// If there are tags, the insert is a bit more involved.
		// Here we are paying the price for having tags be more than just a simple string
		plainTags := metric.Tags
		placeholders := ""

		for i := range plainTags {
			placeholders += "($" + strconv.Itoa(i+1) + "), "
		}
		j := len(plainTags) + 1

		dbResult = db.Exec(
			`WITH tags_for_metric as (
          VALUES `+strings.Trim(placeholders, ", ")+`
        ), new_metric as (
          INSERT INTO metrics (key, value, timestamp)
          VALUES ($`+
				strconv.Itoa(j)+`, $`+strconv.Itoa(j+1)+`, $`+strconv.Itoa(j+2)+`)
           RETURNING id
        ), tag as (
          INSERT INTO tags (text)
    	     SELECT * from tags_for_metric
        ON CONFLICT (text)
        DO UPDATE SET text = excluded.text
        RETURNING id
      )

      INSERT INTO metric_tags(metric_id, tag_id)
      SELECT new_metric.id, tag.id
      FROM new_metric
      CROSS JOIN tag`,
			plainTags, metric.Key, metric.Value, metric.Timestamp)
	}

	metricsInserted, err = dbResult.RowsAffected, dbResult.Error

	return
}

// GetMetricsOpts are the available options for the input to GetMetrics.
// They will be automatically parsed from the query string
type (
	GetMetricsOpts struct {
		Key          string `json:"key" form:"key" query:"key"`                               // search a specific key
		Tag          string `json:"tag" form:"tag" query:"tag"`                               // search for a specific tag
		MinTimestamp int64  `json:"min_timestamp" form:"min_timestamp" query:"min_timestamp"` // inclusive
		MaxTimestamp int64  `json:"max_timestamp" form:"max_timestamp" query:"max_timestamp"` // exclusive
		// It might be nice to implement the below in a future iteration
		// Tags         []string `json:"tags" form:"tags" query:"tags"` // search a metrics containing all tags
	}
)

// GetMetrics retrieves metrics from the database
// These my be filtered by key, tag, min_timestamp, and max_timestamp
func (db *DB) GetMetrics(query GetMetricsOpts) (metrics []PlainMetric, err error) {
	var dbResult *gorm.DB

	type Result struct {
		Tags json.RawMessage
		PlainMetric
	}

	var results []Result

	// build the where clause from the options
	var whereClause string
	i := 1
	vals := make([]interface{}, 0)

	if query.Tag != "" {
		whereClause = whereClause + fmt.Sprintf("t.text=$%d AND ", i)
		i++
		vals = append(vals, query.Tag)
	}
	if query.Key != "" {
		whereClause = whereClause + fmt.Sprintf("m.key=$%d AND ", i)
		i++
		vals = append(vals, query.Key)
	}
	if query.MinTimestamp != 0 {
		whereClause = whereClause + fmt.Sprintf("m.timestamp>=$%d AND ", i)
		i++
		vals = append(vals, query.MinTimestamp)
	}
	if query.MaxTimestamp != 0 {
		whereClause = whereClause + fmt.Sprintf("m.timestamp<$%d AND ", i)
		i++
		vals = append(vals, query.MaxTimestamp)
	}
	if whereClause != "" {
		whereClause = "WHERE " + strings.Trim(whereClause, "AND ")
	}

	// Run the query

	// In production, we'd need a reasonable paging strategy
	// For now, we'll just set a fairly high upper limit so we don't crash the database
	dbResult = db.
		Raw(`SELECT m.*, COALESCE(jsonb_agg(
      t.text
    ) FILTER (WHERE t IS NOT NULL), '[]') AS tags
    FROM metrics as m
    LEFT JOIN metric_tags as mt on m.id = mt.metric_id
    LEFT JOIN tags as t on mt.tag_id = t.id `+
			whereClause+
			` GROUP BY m.id ORDER BY m.timestamp asc LIMIT 100000`, vals).Scan(&results)

	if err = dbResult.Error; err != nil {
		return
	}

	// Convert the results to the proper type
	metrics = make([]PlainMetric, len(results))

	for i, v := range results {
		metrics[i] = PlainMetric{
			Key:       v.Key,
			Value:     v.Value,
			Timestamp: v.Timestamp,
		}
		json.Unmarshal(v.Tags, &metrics[i].Tags)
	}
	return
}
