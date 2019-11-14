# Double-Team

[![Go Report Card](https://goreportcard.com/badge/github.com/msales/double-team)](https://goreportcard.com/report/github.com/msales/double-team)
[![Build Status](https://travis-ci.org/msales/double-team.svg?branch=master)](https://travis-ci.org/msales/double-team)
[![Docker build](https://img.shields.io/docker/automated/msales/double-team.svg)](https://hub.docker.com/r/msales/double-team/)
[![Coverage Status](https://coveralls.io/repos/github/msales/double-team/badge.svg?branch=master)](https://coveralls.io/github/msales/double-team?branch=master)
[![GitHub release](https://img.shields.io/github/release/msales/double-team.svg)](https://github.com/msales/double-team/releases)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/msales/double-team/master/LICENSE)

## Synopsis

A HTTP Kafka producer that handles outages.

## Usage

Double-Team can be used in two different modes: `server` and `restore`

### Server

Server mode accepts HTTP post requests and publishes them to Kafka.

### Restore

Restore mode sends messages from S3 to Kafka.

## Configuration

### Server
The Double-Team server `./double-team server` can be configured with the following options:

| Flag | Description | Environment Variable |
| ---- | ----------- | -------------------- |
| --log.level | The log level to use (options: debug, info, warn, error). | DOUBLE_TEAM_LOG_LEVEL |
| --stats | The statistics service to send metrics to. | DOUBLE_TEAM_STATS |
| --kafka.brokers | The kafka seed brokers connect to. Format: 'ip:port' (multiple allowed). | DOUBLE_TEAM_KAFKA_BROKERS |
| --kafka.version | Version of Kafka for producing messages: '2.3.0'. | DOUBLE_TEAM_KAFKA_VERSION |
| --kafka.retry | The number of times to retry sending to Kafka. | DOUBLE_TEAM_KAFKA_RETRY |
| --s3.endpoint | The S3 endpoint to use. This is mainly for debugging. | DOUBLE_TEAM_S3_ENDPOINT |
| --s3.region | The S3 region the bucket exists in. | DOUBLE_TEAM_S3_REGION |
| --s3.bucket | The S3 bucket to write messages to. | DOUBLE_TEAM_S3_BUCKET |
| --port | The address to bind to for the http server. | DOUBLE_TEAM_PORT |

### Restore
The Double-Team server `./double-team restore` can be configured with the following options:

| Flag | Description | Environment Variable |
| ---- | ----------- | -------------------- |
| --log.level | The log level to use (options: debug, info, warn, error). | DOUBLE_TEAM_LOG_LEVEL |
| --stats | The statistics service to send metrics to. | DOUBLE_TEAM_STATS |
| --kafka.brokers | The kafka seed brokers connect to. Format: 'ip:port' (multiple allowed). | DOUBLE_TEAM_KAFKA_BROKERS |
| --kafka.version | Version of Kafka for producing messages: '2.3.0'. | DOUBLE_TEAM_KAFKA_VERSION |
| --kafka.retry | The number of times to retry sending to Kafka. | DOUBLE_TEAM_KAFKA_RETRY |
| --s3.endpoint | The S3 endpoint to use. This is mainly for debugging. | DOUBLE_TEAM_S3_ENDPOINT |
| --s3.region | The S3 region the bucket exists in. | DOUBLE_TEAM_S3_REGION |
| --s3.bucket | The S3 bucket to read messages from. | DOUBLE_TEAM_S3_BUCKET |

## Server HTTP Endpoints

#### POST /

Accepts a JSON payload with the message topic and data.

##### Payload:
```json
{
	"topic": "test",
	"data": "test data"
}
```

#### GET /health

Gets the current health status of the server. Returns a 200 status code if the server is healthy, otherwise a 503 status code

## License

MIT-License. As is. No warranties whatsoever. Mileage may vary. Batteries not included.
