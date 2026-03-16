# Contributing Guide

## Raising Issues

If you want to report a bug, please create a [bug report](https://github.com/sky-uk/kfp-operator/issues/new?template=bug_report.md).

If you want to propose a new feature, please raise a [feature request](https://github.com/sky-uk/kfp-operator/issues/new?template=feature_request.md).

## Contributing

If you want to submit a change, please fork this repository and submit a pull request to merge your changes.

Please refer to the [development guide](DEVELOPMENT.md) to get started.

## PR Titles (Commit Message Format to master)

PRs are squashed and merged into the `master` branch, and the PR title will form
the commit message on master by default. The required format is:

```
<type>(<scope>): <short summary>
  │       │             │
  │       │             └─⫸ Summary in present tense. Not capitalized. No period at the end.
  │       │
  │       └─⫸ Commit Scope
  │
  └─⫸ Commit Type
```

### Commit Type

Typically this will be:

| Type         | Description                                                                                              |
|--------------|----------------------------------------------------------------------------------------------------------|
| **build**    | Changes that affect the build system or external dependencies (example scopes: go, helm, docker, kfpsdk) |
| **chore**    | Routine tasks, dependency updates, and tooling changes that don't affect production code                 |
| **ci**       | Changes to our CI configuration files and scripts (examples: Github Actions, SauceLabs)                  |
| **docs**     | Documentation only changes                                                                               |
| **feat**     | A new feature                                                                                            |
| **fix**      | A bug fix                                                                                                |
| **perf**     | A code change that improves performance                                                                  |
| **refactor** | A code change that neither fixes a bug nor adds a feature                                                |
| **test**     | Adding missing tests or correcting existing tests                                                        |

Anything that suits the commit better and is stylistically similar is also acceptable.

### Commit Scope

The scope of the affected changes. Sometimes changes might span multiple modules
so discretion is required when choosing the scope. If no single module is appropriate
then it can be omitted. It could be a sign that the changes are affecting too much
scope and therefore should be broken down into separate PRs.
