# App Autoscaler release process

1. Published releases are picked up by [bosh.io](https://bosh.io/releases/github.com/cloudfoundry-incubator/app-autoscaler-release?all=1) automatically, and it is not possible to remove published releases, so make sure that the main branch is in good shape before releasing.
2. In our release pipeline check the latest [draft] build and check the generated changelog:
3. For each PR listed (except those created by dependabot) check:
    * Are the consequences for operators or end-users of the App Autoscaler clear from the PR title?
      If not, feel free to reword the tile to make it clearer
    * Does the PR have an impact on the operators or end-users of the App Autoscaler?
      Changes to our CI should probably be labelled with `exclude-from-changelog` as they typically have no impact on operators or end-users.
    * Is it labeled correctly?
      Not labeled PRs end up in the `Other` section, but should probably be assigned a valid [label].
4. If you have changed PR titles or labels, re-run the [draft] job and restart at step 2.
5. Once the [draft] looks fine, trigger a [release] build.
6. Feel free to edit the GitHub release notes, if you feel operators or end-users of the App Autoscaler should get more information for this release.
   E.g., for breaking changes it might make sense to explain the change at the top of the release.

[label]: https://github.com/cloudfoundry/app-autoscaler-release/blob/d563f615957d86fcf6500c0535e1e8fde8fce53f/src/changelog/display/output.go#L23-L37
[draft]: https://concourse.app-runtime-interfaces.ci.cloudfoundry.org/teams/app-autoscaler/pipelines/app-autoscaler-release/jobs/draft/builds/latest
[release]: https://concourse.app-runtime-interfaces.ci.cloudfoundry.org/teams/app-autoscaler/pipelines/app-autoscaler-release/jobs/release/builds/latest
