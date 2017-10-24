.PHONY: extension
extension::
	./scripts/update
	zip -r app-autoscaler-extension.zip config manifests jobs packages scripts src templates extension-manifest.yml LICENSE NOTICE README.md