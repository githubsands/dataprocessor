module github.com/data-processing

go 1.18

require (
	data-processing/batch v0.0.0-00010101000000-000000000000
	github.com/davecgh/go-spew v1.1.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/kr/pretty v0.1.0 // indirect
	gonum.org/v1/gonum v0.11.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace data-processing/batch => ./batch
