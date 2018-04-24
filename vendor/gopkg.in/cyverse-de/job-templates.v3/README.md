job-templates
=============

A Go package that is used in condor-launcher. It's been separated out to make dependency management a bit easier.

# Build

First, run dep to make sure all of the dependencies are pulled down:

```dep ensure```

The vendor/ directory is listed in the .gitignore since this is a Go package and not an executable.

Next: ```go build```

# Test

The required test files are located in the test/ directory. The unit tests expect them to be present.

As normal for Go projects: ```go test```

# Versioning

We're using gopkg.in to version the package. If you make a small change, tag it with a minor version bump. If it's a backwards incompatible change or a major change to the package, do a major version bump.
