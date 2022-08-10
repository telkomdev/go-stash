.PHONY : test test-cover format clean

ALL_PACKAGES=$(shell go list ./... | grep -v "vendor")

format:
	find . -name "*.go" -not -path "./vendor/*" -not -path ".git/*" | xargs gofmt -s -d -w

test:
	$(foreach pkg, $(ALL_PACKAGES),\
	go test -race -v $(pkg);)

test-cover:
	$(foreach pkg, $(ALL_PACKAGES),\
	go test -race -covermode=atomic -coverprofile=coverage.txt $(pkg);)

clean:
	@echo "cleaning unused file"
	rm -rf coverage.txt
