package version

// set via -ldflags "-X github.com/prjvvl/exe/internal/version.Version=vX.Y.Z"
// (goreleaser does this; a plain `go build` leaves it "dev")
var Version = "dev"
