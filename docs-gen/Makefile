build: clean
	hugo --minify
	@# Workaround for https://github.com/google/docsy/issues/350 and https://github.com/gohugoio/hugo/issues/7714
	$(eval URL_PATH = $(shell grep baseURL config.toml | sed "s/'//g" |  cut -d '/' -f 4))
	@ if [ ! -z "$(URL_PATH)" ]; \
	  then \
	    sed -i '' "s/background-image:url(/background-image:url(\/$(URL_PATH)/g" ../docs/index.html; \
      fi

clean:
	rm -rf ../docs/

serve:
	hugo server -b http://localhost:1313/
