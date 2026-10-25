# Contract: Language additions

Two permanent v1.x attributes. Both are optional; a project that uses neither compiles and exports
exactly as in 1.0.2.

## `project.ignore_warnings`

```hcl
project "shop" {
  description     = "Online shop"
  ignore_warnings = ["missing_docs", "orphan_element"]
}
```

| Rule | Result |
|---|---|
| Each entry is a warning code loko emits | that warning is left out of every command's output and of `--strict` |
| An entry is an error code | `invalid_attribute_value` error: errors cannot be ignored |
| An entry is unknown (for example retired in an upgrade) | `unknown_warning_code` warning, naming the entry |
| Not a list of strings | `invalid_attribute_value` error |

`unknown_warning_code` itself can be ignored.

## `instance.planned`

```hcl
deployment "prod" {
  node "aws" {
    instance "api" {
      of      = container.api
      planned = true
    }
  }
}
```

| Rule | Result |
|---|---|
| `planned = true`, no binding | no `unbound_instance` warning |
| `planned = true`, with a binding | `planned_instance_bound` warning: remove `planned` |
| `planned = false` or absent | as in 1.0.2 |
| Not a bool | `invalid_attribute_value` error |

The export carries `"planned": true` on the instance only when it is set. Diagrams do not show it
in this feature.

## New diagnostic codes

| Code | Severity | Meaning | Fix |
|---|---|---|---|
| `planned_instance_bound` | warning | a `planned` instance has a binding | remove `planned`; the instance exists now |
| `unknown_warning_code` | warning | `ignore_warnings` names a code loko does not emit | remove or correct the entry |
