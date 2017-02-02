# orca-executions

app to get longest running Orca executions

## Install

Get it from the releases page.

## Usage

```
$ orca-executions [-type orchestration|pipeline] [-status ExecutionStatus]
```

* `type` will default to `orchestration`
* `status` will default to `RUNNING`

`ExecutionStatus` is one of the enum values here: https://github.com/spinnaker/orca/blob/master/orca-core/src/main/groovy/com/netflix/spinnaker/orca/ExecutionStatus.groovy
