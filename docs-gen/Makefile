include ../help.mk

build: clean install-npm-dependencies ## Build the site
	hugo --minify

quickstart: ## Build quickstart docker image
	$(MAKE) -C includes/master/quickstart docker-build

docker-push-quickstart-standalone:
	$(MAKE) -C includes/master/quickstart docker-push

docker-push-quickstart-base:
	$(MAKE) -C includes/master/quickstart-with-dependencies/quickstart-base docker-push

docker-push-quickstart-dependant:
	$(MAKE) -C includes/master/quickstart-with-dependencies/quickstart-dependant docker-push

docker-push-kfpsdk-quickstart:
	$(MAKE) -C includes/master/kfpsdk-quickstart docker-push

docker-push-quickstart: ## Push quickstart docker images
	$(MAKE) -j docker-push-quickstart-standalone docker-push-quickstart-base docker-push-quickstart-dependant docker-push-kfpsdk-quickstart

clean: ## Clean the site
	find ../docs -mindepth 1 ! -name '.nojekyll' -delete

serve: ## Serve the site locally
	hugo server -b http://localhost:1313/kfp-operator/

install-npm-dependencies: ## Install npm dependencies
ifeq ($(CI),true)
	npm ci
else
	npm install
endif

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

	# Copy existing docs and static assets into versioned directories
	cp -r content/docs/ content/versions/${VERSION}/
	cp -r static/master/ static/versions/${VERSION}/

	# Inline all readfile shortcodes in versioned content, replacing them with fenced code blocks
	@find content/versions/${VERSION}/ -name '*.md' | while read md_file; do \
		grep -q 'readfile' "$$md_file" || continue; \
		tmp=$$(mktemp); \
		while IFS= read -r line; do \
			filepath=$$(echo "$$line" | sed -n 's|.*readfile file="/\([^"]*\)".*|\1|p'); \
			lang=$$(echo "$$line" | sed -n 's|.*lang="\([^"]*\)".*|\1|p'); \
			if [ -n "$$filepath" ] && [ -f "$$filepath" ]; then \
				echo '```'"$$lang"; \
				cat "$$filepath"; \
				echo '```'; \
			else \
				echo "$$line"; \
			fi; \
		done < "$$md_file" > "$$tmp"; \
		mv "$$tmp" "$$md_file"; \
	done

	# Update static asset paths (images, etc.) from /master/ to /versions/VERSION/
	find content/versions/${VERSION}/ -type f -name '*.md' -exec sed -i '' 's|>}}/master/|>}}/versions/${VERSION}/|g' {} +

	# Update GitHub branch parameter references to point to the release tag
	find content/versions/${VERSION}/ -type f -name '*.md' -exec sed -i '' 's|{{< param "github_branch" >}}|${VERSION}|g' {} +

	# Update hardcoded GitHub URLs to point to the release tag
	find content/versions/${VERSION}/ -type f -name '*.md' -exec sed -i '' 's|github.com/sky-uk/kfp-operator/tree/master|github.com/sky-uk/kfp-operator/tree/${VERSION}|g' {} +
	find content/versions/${VERSION}/ -type f -name '*.md' -exec sed -i '' 's|github.com/sky-uk/kfp-operator/blob/master|github.com/sky-uk/kfp-operator/blob/${VERSION}|g' {} +
	find content/versions/${VERSION}/ -type f -name '*.md' -exec sed -i '' 's|raw.githubusercontent.com/sky-uk/kfp-operator/master|raw.githubusercontent.com/sky-uk/kfp-operator/${VERSION}|g' {} +

	# Add an entry to our version selector for the new version
	echo '\n[[params.versions]]\nversion = "${VERSION}"\nurl = "/kfp-operator/versions/${VERSION}"' >> hugo.toml

	# Update the index page to reflect its version
	sed -i '' 's|"Documentation (master)"|"Documentation (${VERSION})"|g' content/versions/${VERSION}/_index.md

	# Update the links to point to latest version
	sed -i '' -E 's|href="versions/v[0-9]+\.[0-9]+\.[0-9]+"|href="versions/${VERSION}"|g' content/_index.md
	sed -i '' -E 's|latest_stable_version = "/v[0-9]+\.[0-9]+\.[0-9]+"|latest_stable_version = "${VERSION}"|g' hugo.toml
	sed -i '' -E 's|url = "/versions/v[0-9]+\.[0-9]+\.[0-9]+"|url = "/versions/${VERSION}"|g' hugo.toml

	# Add current date to the versioned docs to maintain order of section navigation
	grep -q '^date:' content/versions/${VERSION}/_index.md || \
	perl -i'' -pe 's/^(title:.*)/$$1\ndate: $(shell date +%F)/' content/versions/${VERSION}/_index.md
