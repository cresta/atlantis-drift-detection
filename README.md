# atlantis-drift-detection
Detect terraform drift in atlantis

# What it does

The general workflow of this repository is:
1. Check out a mono repo of terraform code
2. Find an atlantis.yaml file inside the repository
3. Use atlantis to run /plan on each project in the atlantis.yaml file
4. For each project with drift
    1. Trigger a GitHub workflow that can resolve the drift
    2. Comment the existence of the drift in slack
5. For each project directory in the atlantis.yaml
   1. Run workspace list
   2. If any workspace isn't tracked by atlantis, notify slack

There is an optional flag to cache drift results inside DynamoDB, so we don't check the same directory twice in a short period of time.

# Example for "Trigger a github workflow that can resolve the drift"

Here is the example we use to resolve a drift.  This workflow touches a "trigger.txt" file that we configure all
of our projects to listen to.

```yaml
name: Force a terraform PR
on:
  workflow_dispatch:
    inputs:
      directory:
        description: >-
          Which directory to force a terraform workflow upon
        required: true
        type: string
jobs:
  plantrigger:
    runs-on: [self-hosted]
    name: Force terraform plan PR
    steps:
      # We use an application to make PRs, rather than hard code a user token
      - name: Generate token
        id: generate_token
        uses: peter-murray/workflow-application-token-action@v1
        with:
          application_id: $ {{ secrets.OUR_APP_ID_FOR_CREATING_PRS }}
          application_private_key: ${{ secrets.OUR_PEM_ID_FOR_CREATING_PRS }}
      - name: Checkout
        uses: actions/checkout@v2
      - run: ./scripts/update_trigger.sh ${DIR}
        env:
          DIR: ${{ github.event.inputs.directory }}
      - name: Create PR to terraform repo
        uses: peter-evans/create-pull-request@v3
        id: make-pr
        with:
          token: ${{ steps.generate_token.outputs.token }}
          branch: replan-${{ github.event.inputs.directory }}
          delete-branch: true
          title: "Force replan of directory ${{ github.event.inputs.directory }}"
          labels: forced-workflow
          committer: Forced Replan <noreply@cresta.ai>
          body: "A forced replan of this directory was triggered via github actions"
          commit-message: "Regenerated plan for ${{ github.event.inputs.directory }}"
```

Our script update_trigger.sh
```bash
#!/bin/bash
set -exou pipefail
# Modify the trigger.txt file of a directory which we expect to trigger an atlantis workflow
if [ ! -d "$1" ]; then
  echo "Unable to find directory $1"
  exit 1
fi
date > "$1/trigger.txt"
```


# Use as a github action

```yaml
name: Drift detection
on:
  workflow_dispatch:
jobs:
  drift:
    name: detects drift
    runs-on: [self-hosted]
    steps:
      - name: detect drift
        uses: cresta/atlantis-drift-detection@v0.0.7
        env:
          ATLANTIS_HOST: atlantis.atlantis.svc.cluster.local
          ATLANTIS_TOKEN: ${{ secrets.ATLANTIS_TOKEN }}
          REPO: cresta/terraform-monorepo
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
          DYNAMODB_TABLE: atlantis-drift-detection
          WORKFLOW_OWNER: cresta
          WORKFLOW_REPO: terraform-monorepo
          WORKFLOW_ID: force_terraform_workflow.yaml
          WORKFLOW_REF: master
          GITHUB_APP_ID: 123456
          GITHUB_INSTALLATION_ID: 123456
          GITHUB_PEM_KEY: ${{ secrets.PR_CREATOR_PEM }}
          CACHE_VALID_DURATION: 168h
```

# Configuration

| Environment Variable     | Description                                                                      | Required | Default                    | Example                                                             |
|--------------------------|----------------------------------------------------------------------------------|----------|----------------------------|---------------------------------------------------------------------|
| `REPO`                   | The github repo to check                                                         | Yes      |                            | `cresta/terraform-monorepo`                                         |
| `ATLANTIS_HOST`          | The Hostname of the Atlantis server                                              | Yes      |                            | `atlantis.example.com`                                              |
| `ATLANTIS_TOKEN`         | The Atlantis API token                                                           | Yes      |                            | `1234567890`                                                        |
| `WORKFLOW_OWNER`         | The github owner of the workflow to trigger on drift                             | No       |                            | `cresta`                                                            |
| `WORKFLOW_REPO`          | The github repo of the workflow to trigger on drift                              | No       |                            | `atlantis-drift-detection`                                          |
| `WORKFLOW_ID`            | The ID of the workflow to trigger on drift                                       | No       |                            | `drift.yaml`                                                        |
| `WORKFLOW_REF`           | The git ref to trigger the workflow on                                           | No       |                            | `master`                                                            |
| `DIRECTORY_WHITELIST`    | A comma separated list of directories to check                                   | No       |                            | `terraform,modules`                                                 |
| `SLACK_WEBHOOK_URL`      | The Slack webhook URL to post updates to                                         | No       |                            | `https://hooks.slack.com/services/1234567890/1234567890/1234567890` |
| `SKIP_WORKSPACE_CHECK`   | Skip checking if the workspace have drifted                                      | No       | `false`                    | `true`                                                              |
| `PARALLEL_RUNS`          | The number of parallel runs to use                                               | No       | `1`                        | `10`                                                                |
| `DYNAMODB_TABLE`         | The name of the DynamoDB table to use for caching results                        | No       | `atlantis-drift-detection` | `atlantis-drift-detection`                                          |
| `CACHE_VALID_DURATION`   | The duration that previous results are still valid                               | No       | `24h`                      | `180h`                                                              |
| `GITHUB_APP_ID`          | An application ID to use for github API calls                                    | No       |                            | `123123`                                                            |
| `GITHUB_INSTALLATION_ID` | An application install ID to use for github API calls                            | No       |                            | `123123`                                                            |
| `GITHUB_PEM_KEY`         | A GitHub PEM key of an application, used to authenticate the app for API calls   | No       |                            | `1231DEADBEAF....`                                                  |

# Local development

Create a file named `.env` inside the root directory and populate it with the correct variables.
Check out the [example file](example.env) or [configuration](#configuration) for details.
This file  won't be checked in because it's inside the [.gitignore](.gitignore).
