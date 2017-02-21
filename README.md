# orca-executions

app to get longest running Orca executions

## Install

Get it from the releases page.

## Usage

```
Usage of orca-executions
  -debug
    	Set if you want debug level logging
  -fields string
    	Extra fields to return in the data table for each record, comma-delimited
  -filters string
    	Extra filters in comma-delimited Key=Value format
  -quiet
    	Set if you do not want logging enabled
  -redisAddr string
    	The address for orca's redis instance (default "localhost:6379")
  -status string
    	the execution status to filter on (default "RUNNING")
  -type string
    	orchestration or pipeline (default "orchestration")
```

* `redisAddr` the address to your redis host (defaults to `localhost:6379`)
* `type` will default to `orchestration`
* `status` will default to `RUNNING`
* `filters` is a comma-delimited `Key=Value` list of filters.
* `quiet` will disable logging
* `debug` will enable debug-level logging

`ExecutionStatus` is one of the enum values here: https://github.com/spinnaker/orca/blob/master/orca-core/src/main/groovy/com/netflix/spinnaker/orca/ExecutionStatus.groovy

### Filters

Filters are applied to each matching hash, inspecting individual hash keys and
their values. By default, filters will perform an equality comparison, but there
are other custom filters for doing more advanced inspection:

#### ContainsStage

`orca-executions -type pipeline -filters ContainsStage=canary`

Will return all running pipelines that contains a stage type of `canary`.

# examples

`orca-executions -status TERMINAL -filters parallel=false,canceled=true`

Will return all executions with a status of TERMINAL, that are not parallel and
have been canceled.

# ContainsStage

`orca-executions -type pipeline -filters ContainsStage=canary`

Will return all running pipelines that contains a stage type of `canary`.
