
GO := $(shell command -v go1.18beta1 || echo "go")

test:
	$(GO) test -v -race -failfast -parallel=2  \
	-count=10 -timeout=20s \
	-cover  -covermode=atomic \
	-coverpkg=github.com/butuzov/harmony \
	-coverprofile=coverage.out \
	./...

cover: test
	go tool cover -html=coverage.out

gomarkdo:
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@v0.3.0

gen-doc:
	gomarkdoc --output doc.md.out  \
		--template-file package=.github/godoc-tpl/package.gotxt \
		--template-file func=.github/godoc-tpl/func.gotxt \
		--template-file file=.github/godoc-tpl/file.gotxt \
		.

doc: gen-doc
	@ cp README.md README.md.orig
	@ gawk '/^<!-- End --->/{f=0}f{next}/^<!-- Start --->/{f=1}1' README.md.orig >README.md
	@ sed -i '1,2d;$d' doc.md.out
	@ sed -i 's/\\`/`/g' doc.md.out
	@ sed -i 's/\\,/,/g' doc.md.out
	@ sed -i 's/\\-/\n-/g' doc.md.out
	@ sed -i 's/\\././g' doc.md.out
	@ sed -i '/<!-- Start --->/r doc.md.out' README.md
	@ unlink README.md.orig
