# orca-executions

app to get longest running Orca executions

## Install

Get it from the releases page.

## Usage

```
$ orca-executions [-type orchestration|pipeline] [-status ExecutionStatus] [-filters FILTERS]
```

* `type` will default to `orchestration`
* `status` will default to `RUNNING`
* `filters` is a comma-delimited `Key=Value` list of filters.

`ExecutionStatus` is one of the enum values here: https://github.com/spinnaker/orca/blob/master/orca-core/src/main/groovy/com/netflix/spinnaker/orca/ExecutionStatus.groovy

### Filters

Filters are applied to each matching hash, inspecting individual hash keys and
their values. By default, filters will perform an equality comparison, but there
are other custom filters for doing more advanced inspection:

# standard behavior

`orca-executions -status TERMINAL -filters parallel=false,canceled=true`

Will return all executions with a status of TERMINAL, that are not parallel and
have been canceled.

# ContainsStage

`orca-executions -type pipeline -filters ContainsStage=canary`

Will return all running pipelines that contains a stage type of `canary`.
