

set-pipeline:
	./set-pipeline.sh

lint:
	shellcheck -x -s bash scripts/*.sh

lint-fix:
	# uses shellcheck to fix any issues it can
	shellcheck -x -s bash scripts/*.sh --fix
