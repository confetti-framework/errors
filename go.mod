module github.com/lanvard/errors

go 1.15

require (
	github.com/stretchr/testify v1.6.1
	github.com/lanvard/syslog v0.0.0-20201006215111-98d4d91dbaa8
)

replace (
	github.com/lanvard/syslog v0.0.0-20201006215111-98d4d91dbaa8 => ../syslog
)