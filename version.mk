ifndef VERSION

VERSION := 0.2.1-34-gcf5a440#$(shell (git describe --tags --match 'v[0-9]*\.[0-9]*\.[0-9]*' --dirty 2>/dev/null || echo v0.0.0) | sed 's/^v//')

version:
	@echo ${VERSION}

endif # VERSION
