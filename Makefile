test:
	go  test -v -race                     \
			-failfast                     \
			-parallel=2                   \
			-count=10                     \
			-timeout=1m                   \
			-cover                        \
			-covermode=atomic  -coverprofile=coverage.out
