build: clean install-npm-dependencies
	hugo --minify
	@# Workaround for https://github.com/google/docsy/issues/350 and https://github.com/gohugoio/hugo/issues/7714
	$(eval URL_PATH = $(shell grep baseURL hugo.toml | sed "s/'//g" |  cut -d '/' -f 4))
	@ if [ ! -z "$(URL_PATH)" ]; \
	  then \
	    sed -i.bak "s/background-image:url(/background-image:url(\/$(URL_PATH)/g" ../docs/index.html && rm ../docs/index.html.bak; \
      fi

quickstart:
	$(MAKE) -C includes/quickstart docker-build

clean:
	find ../docs -mindepth 1 ! -name '.nojekyll' -delete

serve:
	hugo server -b http://localhost:1313/

install-npm-dependencies:
	npm install
