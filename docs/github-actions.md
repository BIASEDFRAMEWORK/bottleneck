# GitHub Actions

Bottleneck can publish SDLC maturity evidence to every pull request without making CI punitive by default.

Use the copyable workflow in [`examples/github-actions/bottleneck-assessment.yml`](../examples/github-actions/bottleneck-assessment.yml). The workflow:

- builds the local Bottleneck CLI from the repository checkout,
- runs `bottleneck discover`,
- runs `bottleneck ingest --auto`,
- appends `bottleneck assess --format=markdown` to `$GITHUB_STEP_SUMMARY`,
- writes `bottleneck assess --format=json` as an artifact.

The workflow does not require a GitHub token, network calls, a database, or external Bottleneck service. It reads local artifacts that your repository or CI job already produced, such as Cucumber, JUnit, LCOV, SARIF or CodeQL, telemetry JSON, design docs, and GitHub Actions workflow files.

## Minimal Commands

```sh
bottleneck discover
bottleneck ingest --auto
bottleneck assess --format=markdown
bottleneck assess --format=json
```

`assess` exits successfully by default so teams can review maturity without blocking the pull request. Keep separate release gates explicit with commands such as `bottleneck diagnose --gate=release` when a repository is ready for punitive checks.
