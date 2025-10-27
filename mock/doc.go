// Package mock contains the basic collection of functions and types for
// controlling mocks and mock request/response setup. It is part of the public
// interface and starting to get stable, however, we are still experimenting
// to optimize the interface and the user experience.
package mock

//revive:disable:line-length-limit // go:generate line length

//go:generate mockgen -package=mock_test -destination=mock_iface_test.go -source=mocks_test.go  IFace,XFace

//revive:enable:line-length-limit
