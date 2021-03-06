# DESCRIPTION

This command is used to show details about  an oncall duty team
from SOMA.

If the oncall duty name is specified as a valid UUID, that ID is
used as the oncallID of the oncall duty to display.

# SYNOPSIS

```
soma oncall show ${name}
```

# ARGUMENT TYPES

Name | Type |     Description   | Default | Optional
 --- |  --- | ----------------- | ------- | --------
name | string | Name of the oncall duty | | no

# PERMISSIONS

The request is authorized if the user either has at least one
sufficient or all required permissions.

Category | Section | Action | Required | Sufficient
 ------- | ------- | ------ | -------- | ----------
omnipotence | | | no | yes
system | global | | no | yes
global | oncall | show | yes | no

# EXAMPLES

```
soma oncall show "Emergency Phone"
```
