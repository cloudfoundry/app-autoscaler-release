
.PHONY: set-pipeline
set-pipeline:
	./set-pipeline.sh

.PHONY: lint
lint:
	shellcheck -x -s bash scripts/*.sh
	shellcheck --external-sources -x -s bash **/*.sh

.PHONY: lint-fix
lint-fix:
	shellcheck --external-sources -x -s bash -f diff **/*.sh
