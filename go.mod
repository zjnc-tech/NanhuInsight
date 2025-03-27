module github.com/zjnc-tech/xenoview

// renovate: datasource=golang-version depName=go
go 1.22.0

replace (
    github.com/zjnc-tech/hive => ./pkg/hive
    github.com/zjnc-tech/agent => ./pkg/agent
    github.com/zjnc-tech/odre => ./pkg/odre
    github.com/zjnc-tech/sdk => ./pkg/sdk
)