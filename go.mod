module github.com/data-processing

go 1.18

require (
	data-processing/sensor/batch v0.0.0-00010101000000-000000000000
	data-processing/sensor/streamer v0.0.0-00010101000000-000000000000
	github.com/davecgh/go-spew v1.1.1
	gonum.org/v1/gonum v0.11.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	golang.org/x/exp v0.0.0-20191002040644-a1355ae1e2c3 // indirect
	golang.org/x/net v0.0.0-20220412020605-290c469a71a5 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220414192740-2d67ff6cf2b4 // indirect
	google.golang.org/grpc v1.45.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace data-processing/sensor/batch => ./sensor/batch

replace data-processing/sensor/streamer => ./sensor/streamer
