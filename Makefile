lint:
	 goimports -format-only -w -l -local `go list -m` ./  && golangci-lint run ./...
cl:
	sh ./scripts/tag.sh