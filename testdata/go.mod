module testdata/main

go 1.17

replace github.com/bddjr/hlfhr => ../

require (
	github.com/bddjr/hlfhr v0.0.0
	golang.org/x/net v0.17.0
)

require golang.org/x/text v0.13.0 // indirect
