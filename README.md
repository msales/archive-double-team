# Double-Team

A HTTP Kafka producer that handles outages.

## Usage

Double-Team can be used in two different modes: `server` and `restore`

### Server

Server mode accepts HTTP post requests and publishes them to Kafka.

## Configuration

### Server
The Double-Team server `./double-team server` can be configured with the following options:

| Flag | Options | Multiple Allowed | Description | Environment Variable |
| ---- | ------- | ---------------- | ----------- | -------------------- |
| --log.level | debug, info, warn, error | No | The log level to use. | DOUBLE_TEAM_LOG_LEVEL |
| --stats | | No | The statistics service to send metrics to. | DOUBLE_TEAM_STATS |
| --kafka.brokers | | Yes | The kafka seed brokers connect to. Format: 'ip:port'. | DOUBLE_TEAM_KAFKA_BROKERS |
| --kafka.retry | | No | The number of times to retry sending to Kafka. | DOUBLE_TEAM_KAFKA_RETRY |
| --s3.endpoint | | No | The S3 endpoint to use. This is mainly for debugging. | DOUBLE_TEAM_S3_ENDPOINT |
| --s3.region | | No | The S3 region the bucket exists in. | DOUBLE_TEAM_S3_REGION |
| --s3.bucket | | No | The S3 bucket to write messages 2 | DOUBLE_TEAM_S3_BUCKET |
| --port | | No | The address to bind to for the http server. | KAGE_PORT |

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

Gets the current health status of the server. Returns a 200 status code if Kage is healthy, otherwise a 503 status code
