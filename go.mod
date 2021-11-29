module gitlab.com/contextualcode/pbrew

go 1.17

replace gitlab.com/contextualcode/pbrew/core => ./core

require (
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	gitlab.com/contextualcode/platform_cc/v2 v2.2.10
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/aws/aws-sdk-go v1.42.13 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)
