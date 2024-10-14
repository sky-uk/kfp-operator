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
