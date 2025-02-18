# This help file needs be imported before targets that have a double-hash label
# that are not under another category (double-hash-ampersand), including
# other imports that introduce such targets. This groups them under `Other`.
##@ Help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Other
# empty category required so that when this file is imported, any labeled
# targets that do not have a category label will be under this category instead
# of whichever category is processed last.
