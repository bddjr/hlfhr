module testdata/main

go 1.21

replace github.com/bddjr/hlfhr => ../

require (
	github.com/bddjr/hlfhr v0.0.0
	golang.org/x/net v0.35.0
)

require golang.org/x/text v0.22.0 // indirect
