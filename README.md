[![Build Status](https://travis-ci.org/BlueMedora/bluemedora-firehose-nozzle.svg?branch=master)](https://travis-ci.org/BlueMedora/bluemedora-firehose-nozzle) 
# bluemedora-firehose-nozzle

The **bluemedora-firehose-nozzle** is a Cloud Foundry component which collects metrics for the Loggregator Firehose and exposes them via a RESTful API.

## BOSH Release

If you wish to use BOSH to deploy the nozzle, the BOSH release for can be found [here](https://github.com/BlueMedora/bluemedora-firehose-nozzle-release).

## Configure Cloud Foundry UAA for Firehose Nozzle

The Blue Medora nozzle requires a UAA user who is authorized to access the loggregator firehose, has `doppler.firehose` premissions. You can add a user by editing your Cloud Foundry BOSH manifest to include the details about this user under the **properties.uaa.clients** section. Example configuration would look like:

```
properties:
  uaa:
    clients:
      bluemedora-firehose-nozzle:
        access-token-validity: 1209600
        authorized-grant-types: authorization_code,client_credentials,refresh_token
        override: true
        secret: <password>
        scope: openid,oauth.approvals,doppler.firehose
        authorities: oauth.login,doppler.firehose
```

For more on setting up a nozzle user with BOSH refer to Cloud Foundry [documentation](https://docs.cloudfoundry.org/loggregator/nozzle-tutorial.html).

For information on managing UAA users within Cloud Foundry refer to [this guide](https://docs.cloudfoundry.org/adminguide/uaa-user-management.html).

## Configuring Nozzle

The Blue Medora Nozzle uses a configuration file, located at `config/bluemedora-firehose-nozzle.json`, to successfully connect to the firehose and expose a RESTful API. Here is an example configuration of the file:

```
{
    "UAAURL": "https://uaa.pcf.envrioment.com",
    "UAAUsername": "apps_metrics_processing",
    "UAAPassword": "password",
    "TrafficControllerURL": "wss://doppler.pcf.envrionment.com:443",
    "SubscriptionID": "bluemedora-nozzle-id",
    "DisableAccessControl": false,
    "InsecureSSLSkipVerify": true,
    "IdleTimeoutSeconds": 30,
    "MetricCacheDurationSeconds": 60,
    "WebServerPort": 8081,
    "WebServerUseSSL": true
}
```

|Config Field | Description |
|:-----------|:-----------|
| UAAURL | The UAA login URL of the Cloud Foundry deployment. |
| UAAUsername | The UAA username that has access to read from Loggregator Firehose. |
| UAAPassword | Password for the `UAAUsername`. |
| TrafficControllerURL | The URL for the Traffic Controller. To find this follow the instructions in the [documentation](https://docs.cloudfoundry.org/loggregator/architecture.html#firehose). |
| SubscriptionID | The subscription ID of the nozzle. To find out more about subscription IDs and nozzle scaling see the [documentation](https://docs.cloudfoundry.org/loggregator/log-ops-guide.html#scaling-nozzles).|
| DisableAccessControl | If `true`, disables authentication with UAA. Used in lattice deployments. |
| InsecureSSLSkipVerify | If `true`, allows insecure connections to the UAA and Traffic Controller endpoints. |
| IdleTimeoutSeconds |  The amount of time, in seconds, the connection to the Firehose can be idle before disconnecting. |
| MetricCacheDurationSeconds | The amount of time, in seconds, the RESTful API web server will cache metric data. The higher this duration the less likely the data will be correct for a certain metric as it could hold stale data. |
| WebServerPort | Port to connect to the RESTful API. |
| WebServerUseSSL | If `true` the RESTful API web server will use HTTPS, else it uses HTTP  |

### Environment Variables

The nozzle can also be configured by setting a set of environment variables. The variables and what config field they map to are listed below.

|Environment Variable | Config Field |
|:-----------|:-----------|
| BM_UAA_URL | UAAURL |
| BM_UAA_USERNAME | UAAUsername |
| BM_UAA_PASSWORD | UAAPassword |
| BM_TRAFFIC_CONTROLLER_URL | TrafficControllerURL |
| BM_SUBSCRIPTION_ID | SubscriptionID |
| BM_DISABLE_ACCESS_CONTROL | DisableAccessControl |
| BM_INSECURE_SSL_SKIP_VERIFY | InsecureSSLSkipVerify |
| BM_IDLE_TIMEOUT_SECONDS | IdleTimeoutSeconds |
| BM_METRIC_CACHE_DURATION_SECONDS | MetricCacheDurationSeconds |
| PORT | WebServerPort |
| BM_WEBSERVER_USE_SSL | WebServerUseSSL |
| BM_STDOUT_LOGGING | Does not correspond to a config field, but signals if logging should save to files or straight to stdout |


## SSL Certificates

The Blue Medora Nozzle uses SSL for it's REST web server if the `WebServerUseSSL` flag is set to true. In order to generate these certificates simply run the command below and answer the questions.

```
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout certs/key.pem -out certs/cert.pem
```

## Running

To run the Blue Medora nozzle simple execute:

```
go run main.go
```
## Webserver

The webserver is how metrics can be pulled out of the nozzle. It provides a RESTful API that requires an authentication token. 

### Token Request 

A token can be requested from the `/token` endpoint. A token times out after 60 seconds. In order to request a token a `GET` with the two header pairs
`username` and `password` with values that correspond to the UAA user in the `bluemedora-firehose-nozzle.json` config.

If a successful login occurs the response will contain a header pair of `token` and the value will be your token.

### Metric Endpoints

Once a valid token is acquired a `GET` request with the header pair `token` and value of your token can be sent to one of the following endpoints:

* `/metron_agents`
* `/syslog_drains`
* `/tps_watchers`
* `/tps_listeners`
* `/stagers`
* `/ssh_proxies`
* `/senders`
* `/route_emitters`
* `/reps`
* `/receptors`
* `/nsync_listeners`
* `/nsync_bulkers`
* `/garden_linuxs`
* `/file_servers`
* `/fetchers`
* `/convergers`
* `/cc_uploaders`
* `/bbs`
* `/auctioneers`
* `/etcds`
* `/doppler_servers`
* `/cloud_controllers`
* `/traffic_controllers`
* `/gorouters`

A JSON response will be sent in the following form:

```
[
   {
      "Deployment":"deployment_name",
      "Job":"job_name",
      "Index":"0",
      "IP":"X.X.X.X",
      "ValueMetrics":{
         "MetricName":integer_value,
         "MetricName":integer_value
      },
      "CounterMetrics":{
         "MetricName":integer_value,
         "MetricName":integer_value
      }
   }
]
```

**NOTE**: Counter metrics are reported as totals over time. The consumer must take the delta between two totals to get the current value as time changes.