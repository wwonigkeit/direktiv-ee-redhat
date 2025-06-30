# 0.8.2

## DB Schema Change

The database schema was changed to remove a UNIQUE constraint on access token descriptions. Additionally, it was changed to reference namespaces by their name, not by their UUID. Since all existing policies and tokens are likely to break in the transition, we recommend dropping these tables entirely and repopulating them after the upgrade.

```sql
DROP TABLE IF EXISTS "access_tokens";
DROP TABLE IF EXISTS "access_groups";
DROP TABLE IF EXISTS "policies";
```

## Rego Changes

Minor tweaks have been made to the input to rego policy evaluation, and as a result the default policy file has been adjusted. Consider overwriting your policy file with the following updated default:

```
package direktiv.authz

import future.keywords.if
import future.keywords.in

verb := "READ" if {
	input.request.method == "GET"
} else := "WRITE"

allowed if {
	some perm in input.user.permissions
    perm == sprintf("%s:%s", [verb, input.request.topic])
} else if {
	some perm in input.user.permissions
    perm == sprintf("%s:%s", [verb, input.recommendedTopicTranslation])
} else := false
```

## Permissions Changes

V2 file-system APIs have been implemented on a new endpoint, this means we've added new default permissions "READ:files-tree" and "WRITE:files-tree". This update has not yet removed the older related permissions ("READ:tree" and "WRITE:tree"), so both are currently still important. The new default policy file includes logic to make a best-effort attempt at keeping existing tokens behaving correctly.
