# vulcan-reports-generator

Micro service responsible for vulcan reports generation, upload and optionally notification.

## Report generation

Report generation requests are read from a queue, where the expected payload complies with:

```javascript
{
    "type": "livereport",
    "team_info": {
        "id": "4d823e6f-7c5b-4174-85ae-6c0add4d65a7",
        "name": "TestTeam",
        "recipients": ["tom@vulcan.example.com"]
    },
    "data": {
        ...
    },
    "auto_send": true
}
```

- **type** indicates the type of report to generate and is used internally to load and execute the appropiate generator and repository for each type.
- **team_info** contains information related to the vulcan team associated with the report and the recipients for the report notification (if auto_send param is set to true).
- **data** contains data only relevant to the report generator for the specified type, so this data JSON object is not fixed, is opaque to the queue events processor and passed to the correspondent generator type.
- **auto_send** indicates if generated report's notification should be sent to the specified recipients.

## API

Reports generation micro service also exposes an API with the following methods:

**Get Report Notification**

```bash
Req:
GET /api/v1/reports/{report_type}/{report_id}/notification

Resp:
{
    "subject": "Vulcan Digest - Adevinta",
    "body": "Vulcan Weekly Digest....",
    "format": "HTML"
}
```

**Send Report Notification**:

```bash
Req:
POST /api/v1/reports/{report_type}/{report_id}/send
{
    "recipients":["tom@vulcan.example.com"]
}

Resp:
HTTP 200 Ok
```

**Healthcheck**

```bash
Req:
GET /healthcheck

Resp:
HTTP 200 Ok
```

## Docker execute

These are the variables you have to use:

|Variable|Description|Sample|
|---|---|---|
|LOG_LEVEL|panic, fatal, error, warn, info, debug or trace (default info)|debug|
|PORT||8080|
|PG_HOST||localhost|
|PG_PORT||5438|
|PG_USER||vulcan_reportgen|
|PG_PASSWORD||vulcan_reportgen|
|PG_SSLMODE|one of: disable, allow, prefer, require, verify-ca, verify-full|disable|
|PG_NAME||vulcan_reportgen|
|SQS_QUEUE_ARN|SQS to push report generation requestsfrom vulcan-api|arn:aws:sqs:xxx:123456789012:yyy|
|SQS_NUM_PROCESSORS|Number of processors|2|
|SES_REGION|AWS region for SES service|xxx|
|SES_FROM|From address to use for AWS SES|vulcan@vulcan.example.com|
|SES_CC|Comma separated list of CC email adresses strings. E.g.: "vulcan@vulcan.example.com","reports@vulcan.example.com"||
|LIVEREPORT_EMAIL_SUBJECT||[Test] Live Report|

```bash
docker build . -t vrg:local

# Use the default config.toml customized with env variables and mapping host AWS dir:
docker run --env-file local.env -p 8080:8080 -v $HOME/.aws:/root/.aws vrg:local

# Use custom config.toml:
docker run -v `pwd`/custom.toml:/app/config.toml -p 8080:8080 vrg:local
```
