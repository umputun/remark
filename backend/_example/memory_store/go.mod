module github.com/umputun/remark/memory_store

go 1.14

require (
	github.com/go-pkgz/jrpc v0.1.0
	github.com/go-pkgz/lgr v0.7.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/umputun/go-flags v1.5.1
	github.com/umputun/remark/backend v1.5.0
)

replace github.com/umputun/remark/backend => ../../
