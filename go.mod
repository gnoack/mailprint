module github.com/gnoack/mailprint

go 1.24.0

toolchain go1.24.5

require (
	github.com/gnoack/picon v0.0.0-20240407101117-35066a944c38
	github.com/landlock-lsm/go-landlock v0.0.0-20250303204525-1544bccde3a3
	github.com/signintech/gopdf v0.33.0
)

require (
	github.com/phpdave11/gofpdi v1.0.15 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/sys v0.36.0 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.71 // indirect
)

exclude (
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.72
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.73
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.74
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.75
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.76
)
