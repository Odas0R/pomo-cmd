build:
	rm -f ./pomo && go build -o pomo ./cmd/pomo
test:
	clear && gotestsum --format standard-verbose ./cmd/pomo
watch:
	find . -name '*.go' | entr -cp richgo test ./cmd/pomo

.PHONY: build test watch
