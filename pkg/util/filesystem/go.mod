module github.com/DataDog/datadog-agent/pkg/util/filesystem

go 1.20

replace (
	github.com/DataDog/datadog-agent/pkg/util/log => ../log/
	github.com/DataDog/datadog-agent/pkg/util/scrubber => ../scrubber/

)

require (
	github.com/DataDog/datadog-agent/pkg/util/log v0.0.0-00010101000000-000000000000
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95
	github.com/shirou/gopsutil/v3 v3.23.9
	github.com/stretchr/testify v1.8.4
	golang.org/x/sys v0.12.0
)

require (
	github.com/DataDog/datadog-agent/pkg/util/scrubber v0.49.0-rc.2 // indirect
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
