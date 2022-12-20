help: help-text #### Details how to build, install, package, and/or deploy.

###################
## Build targets ##
###################

test: ## Runs unit tests against all packages.
	go test -v ./...
	@printf $(green_start)"Completed running all unit tests (read output, this does not mean they passed)."$(green_end)

test-all: ## Runs unit and integration tests against all packages.
	go test -v -tags=integration ./...
	@printf $(green_start)"Completed running all integration tests (read output, this does not mean they passed)."$(green_end)

build: ## Creates a proctor binary at ./out/proctor. Uses host's OS and Arch.
	go build -o ./out/proctor ./proctor/main.go
	@printf $(green_start)"Built and saved proctor to ./out/proctor."$(green_end)

install: ## Creates a proctor binary and installs it to $GOBIN.
	go install ./proctor
	@printf $(green_start)"Installed proctor to "$(install_path)"proctor"$(green_end)

##################
## Help targets ##
##################

help-text:
	@awk 'BEGIN {FS = ":.*## "; printf "\nTargets:\n"} /^[a-zA-Z_-]+:.*?#### / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@awk 'BEGIN {FS = ":.* ## "; printf "\n  \033[1;32mDevelopment\033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*? ## / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@awk 'BEGIN {FS = ":.* ### "; printf "\n  \033[1;32mRelease\033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*? ### / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

###############
## Constants ##
###############

green_start := "\033[1;32m"
green_end = "\033[36m\033[0m\n"

# install_path reflects where a `go install` may land a binary
install_path = "$${HOME}/go/bin/"
ifdef GOPATH
install_path = "$${GOPATH}/bin/"
endif
ifdef GOBIN
install_path = "$${GOBIN}/"
endif
