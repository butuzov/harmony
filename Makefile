
GO := $(shell command -v go1.18beta1 || echo "go")

test:
	$(GO) test -v -race -failfast -parallel=2  \
	-count=10 -timeout=1m \
	-cover  -covermode=atomic -coverprofile=coverage.out
