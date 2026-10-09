# loko examples

Four small architectures written in the loko language. Each directory is a complete project: one
or more `*.loko.hcl` files plus a `docs/` folder of markdown prose referenced with
`docs = "./docs/…"`. All four pass `loko validate --strict` with no errors and no warnings.

The language is described in [docs/language.md](../docs/language.md).

| Example | Shows |
|---|---|
| [simple-project](./simple-project/) | The minimum: one system, a client, three containers, three components, one declared view |
| [3layer-app](./3layer-app/) | A three-tier web app as one system; components per tier, a read replica and backups, the `elk` layout, a tag-based view |
| [microservices](./microservices/) | Four services as separate systems, one file per service, Kafka events (`kind = "async"` and `"trigger"`), owners, a `prod` deployment with Terraform bindings |
| [serverless](./serverless/) | AWS Lambda, API Gateway, SQS, DynamoDB and EventBridge; Lambda triggers, `shape` on queues, functions and tables, a `prod` deployment with CloudFormation bindings |

## Running an example

Install loko (`brew install --cask madstone-tech/tap/loko` or
`go install github.com/madstone-tech/loko@latest`), then from the repository root:

```bash
# Check the model. --strict makes warnings fail the run too.
loko validate --strict -p examples/serverless

# Check the files are in canonical form (loko fmt rewrites them).
loko fmt --check -p examples/serverless

# Render diagrams (d2, svg), markdown and an HTML site into examples/serverless/dist.
loko build -p examples/serverless

# Serve the site on http://127.0.0.1:8080 and rebuild on every change.
loko serve -p examples/serverless

# Ask questions of the model, or export it.
loko query -p examples/serverless --help
loko export -p examples/serverless --format json
```

Without `-p`, loko uses the current directory, so `cd examples/serverless && loko build` works
too. Rendering is in-process; no `d2` binary is needed.

`dist/` is generated output and is not committed here.

## Starting your own

Copy `simple-project`, rename the `project` block, and replace the elements. Run
`loko validate` after each change: every error names a file, line and column.

## CI

[ci/](./ci/) has a GitHub Actions workflow and a Docker Compose file for validating and building
a loko project.
