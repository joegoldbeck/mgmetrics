# MGMetrics - A Metrics Store

I wrote this in response to a prompt for a prospective employer, and decided to use it as an opportunity to learn Go. I may or may not add to this project over time.

## Running The Project Locally

In order to run locally, the mgmetrics needs a Postgres server with port 5432 open, a user "metrics_user", and a password "dev_only". The easiest way to do this is via `docker-compose` with the command

`docker-compose up`

in the root directory of this project.

The run the server, you'll first need to install and build the go dependencies with

`dep ensure`

and

`go install`

and then run the executable with

`$GOPATH/bin/mgmetrics-api` (assuming you have your GOPATH configured correctly)

This will start a server running at `localhost:8080`

### Populating the database

You can use an included utility to populate the database with some random metric streams

`$GOPATH/bin/mgmetrics-populate-test-db`

You can control the number of different metric streams and how many datapoints are in each stream with the command line arguments `streams` and `length` respectively. For example,

`$GOPATH/bin/mgmetrics-populate-test-db -streams 5 -length 100`

will clear the database and then insert 5 streams of metrics with 100 datapoints each

### API

#### `POST /api/metrics`: Adding Metrics

This route accepts a JSON-formatted metric structured like

```
{
  key: String,
  value: Float,
  tags: []String
}
```

for example

```
{
  key: 'heartrate',
  value: 82.4,
  tags: ['icu', 'ward', 'bed22']
}
```

##### Example usage

`curl -H "Content-Type: application/json" -X POST -d '{"key": "heartrate", "value": 52.4, "tags": ["icu", "ward-4", "bed-22"]}' http://localhost:8080/api/metrics`

#### `GET /api/metrics`: Retrieving Metrics

This route accepts the following (optional) query parameters

- `key: String` retrieve metrics with this key
- `tag: String` retrieve metrics containing this tag
- `minTimestamp: Double` retrieve metrics received after this timestamp (inclusive)
- `maxTimestamp: Double` retrieve metrics received before this timestamp (exclusive)

There is currently no way to limit results through the api, or indicate whether additional pages would be available, but we would expect to add this in a production version. There is currently a hardcoded upper limit of 100000 metrics just to prevent things from falling over.

This route returns a JSON-formatted array of metrics of the form

```
[]{
  key: String,
  value: Float,
  tags: []String,
  timestamp: Double // milliseconds since unix epoch
}
```

for example

```
[{
  key: 'heartrate',
  value: 82.4,
  tags: ['icu', 'ward', 'bed-22'],
  timestamp: 1534895313651
}]
```

##### Example usage

`curl -X GET http://localhost:8080/api/metrics\?tag\="bed-22"\&key\="heartrate"`

`curl -X GET http://localhost:8080/api/metrics\?&key\="heartrate"&minTimestamp\=1505879975574&maxTimestamp\=1537895313651`
