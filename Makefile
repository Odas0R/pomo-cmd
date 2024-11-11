build:
	rm -f ./pomo && go build -o pomo ./cmd/pomo && install ./pomo /home/odas0r/.local/bin/pomo
test:
	clear && gotestsum --format standard-verbose ./cmd/pomo
watch:
	find . -name '*.go' | entr -cp richgo test ./cmd/pomo

.PHONY: build test watch
