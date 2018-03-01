# Double-Team

A HTTP Kafka producer that handles outages.

## Usage

Double-Team can be used in two different modes: `server` and `restore`

### Server

Server mode accepts HTTP post requests and publishes them to Kafka.

## Configuration

### Server
The Double-Team server `./double-team server` can be configured with the following options:

| Flag | Description | Environment Variable |
| ---- | ----------- | -------------------- |
| --log.level | The log level to use (options: debug, info, warn, error). | DOUBLE_TEAM_LOG_LEVEL |
| --stats | The statistics service to send metrics to. | DOUBLE_TEAM_STATS |
| --kafka.brokers | The kafka seed brokers connect to. Format: 'ip:port' (multiple allowed). | DOUBLE_TEAM_KAFKA_BROKERS |
| --kafka.retry | The number of times to retry sending to Kafka. | DOUBLE_TEAM_KAFKA_RETRY |
| --s3.endpoint | The S3 endpoint to use. This is mainly for debugging. | DOUBLE_TEAM_S3_ENDPOINT |
| --s3.region | The S3 region the bucket exists in. | DOUBLE_TEAM_S3_REGION |
| --s3.bucket | The S3 bucket to write messages 2 | DOUBLE_TEAM_S3_BUCKET |
| --port | The address to bind to for the http server. | DOUBLE_TEAM_PORT |

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
