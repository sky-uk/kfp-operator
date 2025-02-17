build: clean install-npm-dependencies
	hugo --minify

quickstart:
	$(MAKE) -C includes/quickstart docker-build

clean:
	find ../docs -mindepth 1 ! -name '.nojekyll' -delete

serve:
	hugo server -b http://localhost:1313/kfp-operator/

install-npm-dependencies:
	npm install

# Archive a version of the documentation. Requires VERSION to be set, e.g. make archive-version VERSION=v0.6.0
archive-version:
	@if [ -z ${VERSION} ]; then \
		echo "You must specify the version of the docs you want to archive by setting VERSION=<new_version>, e.g. VERSION=v0.6.0"; \
		exit 1; \
	fi
	cp -r content/docs/ content/archive/${VERSION}/
	cp -r static/images/ static/images/archive/${VERSION}/
	cp -r includes/ includes/archive/${VERSION}/

	echo '\n[[params.versions]]\nversion = "${VERSION}"\nurl = "/kfp-operator/archive/${VERSION}"' >> hugo.toml

	find content/archive/${VERSION}/ -type f -exec sed -i '' 's|/includes|/includes/archive/${VERSION}|g' {} +
	find content/archive/${VERSION}/ -type f -exec sed -i '' 's|/images|/images/archive/${VERSION}|g' {} +

	sed -i '' 's|"Documentation"|"Documentation (${VERSION})"|g' content/archive/${VERSION}/_index.md
