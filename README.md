# vulcan-reports-generator
Micro service responsible for vulcan reports generation, upload and optionally notification.

### Report generation
Report generation requests are readed from a queue, where the expected payload complies with:
```
// example for vulcan scan report
{
    "type": "scan", 
    "team_info": {
        "id": "4d823e6f-7c5b-4174-85ae-6c0add4d65a7",
        "name": "TestTeam",
        "recipients": ["tom@vulcan.example.com"]
    }, 
    "data": {
        "scan_id": "2a636dc9-4f50-401a-96a7-169c952aeaa8",
        "program_name": "Periodic Scan"
    },
    "auto_send": true
}
```
- **type** indicates the type of report to generate and is used internally to load and execute the appropiate generator and repository for each type.
- **team_info** contains information related to the vulcan team associated with the report and the recipients for the report notification (if auto_send param is set to true).
- **data** contains data only relevant to the report generator for the specified type, so this data JSON object is not fixed, is opaque to the queue events processor and passed to the correspondent generator type.
- **auto_send** indicates if generated report's notification should be sent to the specified recipients.

### API
Reports generation micro service also exposes an API with the following methods:

**Important NOTE:**
Take into account that for the scan reports, in the following API specs the _report\_id_ should be the report _scan\_id_. This is due to the fact that the API (which is the main component communicating with this microservice), is not aware of auto generated report_id, instead it only knows about _scan\_id_.

**Get Report:**
```
Req:
GET /api/v1/reports/{report_type}/{report_id}

Resp:
// response is dependent on each report type. example for scan report:
{
    "id":"b4b5b6cc-e359-46c8-aebf-2204ca261554",
    "report_url":"https://insights.vulcan.example.com/e0c5af2796904e9a50251e71e1e0b71ef2d02bccaa427f574f3cefb2c33af643/2020-04-08/2a636dc9-4f50-401a-96a7-169c952aecca-full-report.html",
    "report_json_url":"https://insights.vulcan.example.com/e0c5af2796904e9a50251e71e1e0b71ef2d02bccaa427f574f3cefb2c33af643/2020-04-08/2a636dc9-4f50-401a-96a7-169c952aecca-full-report.json","scan_id":"2a636dc9-4f50-401a-96a7-169c952aecca",
    "program_name":"Periodic Scan",
    "status":"FINISHED",
    "risk":2,
    "email_subject":"[ACTION SUGGESTED] [Test] Security Overview - TestTeam - Periodic Scan",
    "email_body":"HTML Email Body data...",
    "delivered_to":"tom@vulcan.example.com",
    "created_at":"2020-05-04T07:28:32.234763Z",
    "updated_at":"2020-05-04T07:28:32.234764Z"
}
```
**Get Report Notification**
```
Req:
GET /api/v1/reports/{report_type}/{report_id}/notification

Resp:
{
    "subject":"[ACTION SUGGESTED] [Test] Security Overview - TestTeam - Periodic Scan",
    "body":"HTML Email Body data...",
    "format":"HTML"
}
```
**Send Report Notification**:
```
Req:
POST /api/v1/reports/{report_type}/{report_id}/send
{
    "recipients":["tom@vulcan.example.com"]
}

Resp:
HTTP 200 Ok
```
**Healthcheck**
```
Req:
GET /healthcheck

Resp:
HTTP 200 Ok
```

### Docker execute
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
|SES_REGION|AWS region for SES service|xxx|
|SES_FROM|From address to use for AWS SES|vulcan@vulcan.example.com|
|SES_CC|Comma separated list of CC email adresses strings. E.g.: "vulcan@vulcan.example.com","reports@vulcan.example.com"||
|SES_SUBJECT|Subject to use for notifications of scan reports|[Test] Security Overview|
|SCAN_S3_PUBLIC_BUCKET|Bucket to use for public resources of scan reports|public-vulcan-insights-abcdefghij|
|SCAN_S3_PRIVATE_BUCKET|Bucket to use for private resources of scan reports|vulcan-insights-abcdefghij|
|SCAN_GA_ID|Google Analytics ID to use for scan reports||
|PERSISTENCE_ENDPOINT|vulcan-scan-engine API endpoint|https://scanengine.vulcan.example.com/|
|RESULTS_ENDPOINT|vulcan-results API endpoint|https://results.vulcan.example.com/|
|SCAN_PROXY_ENDPOINT|Proxy endpoint to access reports|https://insights.vulcan.example.com/|
|SCAN_COMPANY_NAME|Company name to use on scan reports||
|SCAN_SUPPORT_EMAIL|Support email to use on scan reports||
|SCAN_CONTACT_EMAIL|Contact email to use on scan reports||
|SCAN_CONTACT_CHANNEL|Contact Slack channel to use on scan reports||
|SCAN_CONTACT_JIRA|Jira project link to use on scan reports||
|SCAN_DOCS_API_LINK|Vulcan API docs link to use on scan reports||
|SCAN_DOCS_ROADMAP_LINK|Vulcan roadmap link to use on scan reports||
|VULCAN_UI||https://www.vulcan.example.com/|
|SCAN_VIEW_REPORT||www.vulcan.example.com/api/v1/report?team_id=%s&scan_id=%s|
|SCAN_REDIRECT_URL|||
|LIVEREPORT_EMAIL_SUBJECT||[Test] Live Report|

```bash
docker build . -t vrg:local

# Use the default config.toml customized with env variables and mapping host AWS dir:
docker run --env-file local.env -p 8080:8080 -v $HOME/.aws:/root/.aws vrg:local

# Use custom config.toml:
docker run -v `pwd`/custom.toml:/app/config.toml -p 8080:8080 vrg:local
```
