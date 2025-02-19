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
	# Check if version param is set
	@if [ -z ${VERSION} ]; then \
		echo "You must specify the version of the docs you want to archive by setting VERSION=<new_version>, e.g. VERSION=v0.6.0"; \
		exit 1; \
	fi

	# Check if version is in the correct format
	@if ! echo "$(VERSION)" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "Invalid version: $(VERSION). Expected format: vMAJOR.MINOR.PATCH"; \
		exit 1; \
	fi

	# Check if version already exists
	@if [ -d "content/versions/${VERSION}" ]; then \
		echo "Version already exists for ${VERSION}"; \
		exit 1; \
	fi

	# Copy existing docs into a versioned directory
	cp -r content/docs/ content/versions/${VERSION}/
	cp -r static/master/ static/versions/${VERSION}/
	cp -r includes/master/ includes/versions/${VERSION}/

	# Add an entry to our version selector for the new version
	echo '\n[[params.versions]]\nversion = "${VERSION}"\nurl = "/kfp-operator/versions/${VERSION}"' >> hugo.toml

	# find all links to the master version and update them to the new version
	find content/versions/${VERSION}/ -type f -exec sed -i '' 's|/master/|/versions/${VERSION}/|g' {} +

	# Update the index page to reflect its version
	sed -i '' 's|"Documentation (master)"|"Documentation (${VERSION})"|g' content/versions/${VERSION}/_index.md

	# Update the quickstart Makefile to include with correct path
	sed -i '' 's|../../../../|../../../../../|g' includes/versions/${VERSION}/quickstart/Makefile

	# Update the links to point to latest version
	sed -i '' -E 's|href="versions/v[0-9]+\.[0-9]+\.[0-9]+"|href="versions/${VERSION}"|g' content/_index.md
	sed -i '' -E 's|latest_stable_version = "/v[0-9]+\.[0-9]+\.[0-9]+"|latest_stable_version = "${VERSION}"|g' hugo.toml
	sed -i '' -E 's|url = "/versions/v[0-9]+\.[0-9]+\.[0-9]+"|url = "/versions/${VERSION}"|g' hugo.toml
