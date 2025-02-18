include ../help.mk

build: clean install-npm-dependencies ## Build the site
	hugo --minify

quickstart: ## Build quickstart docker image
	$(MAKE) -C includes/master/quickstart docker-build

clean: ## Clean the site
	find ../docs -mindepth 1 ! -name '.nojekyll' -delete

serve: ## Serve the site locally
	hugo server -b http://localhost:1313/kfp-operator/

install-npm-dependencies: ## Install npm dependencies
	npm install

archive-version: ## Archive a version of the documentation. Requires VERSION to be set, e.g. make archive-version VERSION=v0.6.0
	@if [ -z ${VERSION} ]; then \
		echo "You must specify the version of the docs you want to archive by setting VERSION=<new_version>, e.g. VERSION=v0.6.0"; \
		exit 1; \
	fi
	cp -r content/docs/ content/versions/${VERSION}/
	cp -r static/master/ static/versions/${VERSION}/
	cp -r includes/master/ includes/versions/${VERSION}/

	echo '\n[[params.versions]]\nversion = "${VERSION}"\nurl = "/kfp-operator/versions/${VERSION}"' >> hugo.toml

	find content/versions/${VERSION}/ -type f -exec sed -i '' 's|/master/|/versions/${VERSION}/|g' {} +

	sed -i '' 's|"Documentation (master)"|"Documentation (${VERSION})"|g' content/versions/${VERSION}/_index.md
