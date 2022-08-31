help: #### Details how to build, install, package, and/or deploy.
	@awk 'BEGIN {FS = ":.*## "; printf "\nTargets:\n"} /^[a-zA-Z_-]+:.*?#### / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@awk 'BEGIN {FS = ":.* ## "; printf "\n  \033[1;32mBuild\033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*? ## / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@awk 'BEGIN {FS = ":.* ### "; printf "\n  \033[1;32mRelease targets\033[36m\033[0m\n\033[0;37m  Targets for producing a TCE release package.\033[36m\033[0m\n\n"} /^[a-zA-Z_-]+:.*? ### / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# Build options [start]

build: ## create proctor binary at ./out/proctor. Uses host's OS and Arch.
	go build -o ./out/proctor ./proctor/main.go
	@printf $(green_start)"Built and saved proctor to ./out/proctor."$(green_end)

# Build targets [end]

# Makefile constants [start]

green_start := "\033[1;32m"
green_end = "\033[36m\033[0m\n"

# Makefile constants [end]
