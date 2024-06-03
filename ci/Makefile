
.PHONY: set-autoscaler-pipeline
set-autoscaler-pipeline:
	@./autoscaler/set-pipeline.sh

.PHONY: unpause-pipeline
unpause-pipeline:
	@./scripts/unpause-pipeline.sh


.PHONY: delete-pipeline
delete-pipeline:
	@./scripts/delete-pipeline.sh

.PHONY: set-target
set-target:
	@fly --target app-autoscaler-release login --team-name app-autoscaler --concourse-url https://concourse.app-runtime-interfaces.ci.cloudfoundry.org

.phony: lint
lint:
	shellcheck --external-sources -x -s bash autoscaler/**/**/*.sh

.phony: fix-lint
fix-lint:
	shellcheck --external-sources -x -s bash -f diff autoscaler/**/**/*.sh


