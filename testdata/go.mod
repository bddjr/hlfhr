module testdata/main

go 1.21

replace github.com/bddjr/hlfhr => ../

require (
	github.com/bddjr/hlfhr v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.46.0
)

require golang.org/x/text v0.30.0 // indirect
