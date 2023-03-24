package mock

import (
	"errors"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

// Loader is the generic interface of the package and interface loader.
type Loader interface {
	// Search searches the go-packages with the given partial source type,
	// either using the provided file path or the provided package path
	// information. The file path is normalized assuming that any non-existing
	// final part is a base file name that should be removed for package
	// loading.
	Search(source *Type) *PkgResponse
	// Load loads the go-packages for the given package path. The details
	// depend on the backing package loader and its configuration.
	Load(path string) *PkgResponse
	// IFaces looks up the go-packages for the given source and extracts
	// the interfaces matching the naming pattern and the file.
	IFaces(source *Type) ([]*IFace, error)
}

// PkgResponse provides a cachable package loader response.
type PkgResponse struct {
	// List of resolved packages.
	pkgs []*packages.Package
	// Package loading errors.
	err error
}

// NewPkgResponse calculates a new package response using given path, package
// list and error information. The method also derives the package path and
// package names using this information.
func NewPkgResponse(
	_ string, pkgs []*packages.Package, err error,
) *PkgResponse {
	for _, pkg := range pkgs {
		// TODO: simplify package path logic for caching etc.
		// if pkg.PkgPath == ReadFromFile {
		// 	pkg.PkgPath = path
		// }

		if pkg.Name == "" && !filepath.IsAbs(pkg.PkgPath) {
			index := strings.LastIndexAny(pkg.PkgPath, "/-")
			name := pkg.PkgPath[index+1:]
			if token.IsIdentifier(name) {
				pkg.Name = name
			}
		}
	}

	return &PkgResponse{
		pkgs: pkgs, err: err,
	}
}

// Path resolves the package path if available.
func (r *PkgResponse) Path() string {
	if len(r.pkgs) > 0 {
		return r.pkgs[0].PkgPath
	}
	return ""
}

// Get expands the package loader response result.
func (r *PkgResponse) Get() ([]*packages.Package, error) {
	return r.pkgs, r.err
}

// CachedLoader allows to efficient load, parse, and analyze packages as well
// as interfaces. The loaded packages are chached by request path for repeated
// access and analyzing. The loader is safe to be used concurrently.
type CachedLoader struct {
	// Configuration for package loading.
	Config *packages.Config
	// Cache for packages mapped to load paths.
	cache map[string]*PkgResponse
	// Mutext to support concurrent usage.
	mutex sync.Mutex
}

// NewLoader creates a new caching package loader, that allows efficient access
// to types from files and packages. By default it provides also access to the
// test packages and searches these for interfaces. To disable you need to set
// `loader.(*CachingLoader).Config.Tests` to `false`.
func NewLoader(dir string) Loader {
	return &CachedLoader{
		Config: &packages.Config{
			Mode: packages.NeedName |
				packages.NeedTypes |
				packages.NeedSyntax,
			Tests: true,
			Dir:   dir,
		},
		cache: map[string]*PkgResponse{},
	}
}

// Search searches the go-packages with the given partial source type, either
// using the provided file path or the provided package path information. The
// file path is normalized assuming that any non-existing final part is a base
// file name that should be removed for package loading.
func (loader *CachedLoader) Search(source *Type) *PkgResponse {
	if source.File != "" {
		return loader.byFile(source.File)
	} else if source.Path != "" {
		return loader.byPath(source.Path)
	}
	return loader.byPath(".")
}

// Load loads the go-packages for the given path and caching them for repeated
// requests. The details provided by the package information depend on the
// configuration. By default only types and names are loaded.
func (loader *CachedLoader) Load(path string) *PkgResponse {
	return loader.byPath(path)
}

// IFaces looks up the go-packages for the given source and extracts the
// interfaces matching the naming pattern and the file.
func (loader *CachedLoader) IFaces(source *Type) ([]*IFace, error) {
	if _, err := os.Stat(loader.normalize(source.File)); err != nil {
		return nil, NewErrNotFound(source, source.Name)
	}
	resp := loader.byPath(source.Path)
	if resp.err != nil {
		return nil, resp.err
	} else if token.IsIdentifier(source.Name) {
		return loader.ifacesNamed(resp.pkgs, source)
	}
	return loader.ifacesAny(resp.pkgs, source)
}

// byFile looks up the go-packages by a given file path. The file path is
// normalized and removes a trailing or none-existing file.
func (loader *CachedLoader) byFile(path string) *PkgResponse {
	return loader.byPath(loader.trimfile(loader.normalize(path)))
}

// byPath looks up the go-packages identified by the package path or the given
// normalized (absolute) file path. The package is either resolved from cache
// or looked up via the package loader. The package is always added to cache
// before returning the result.
func (loader *CachedLoader) byPath(path string) *PkgResponse {
	loader.mutex.Lock()
	if resp, ok := loader.cache[path]; ok {
		loader.mutex.Unlock()
		return resp
	}
	loader.mutex.Unlock()

	resp := loader.load(path)
	rpath := resp.Path()

	loader.mutex.Lock()
	loader.cache[path] = resp
	if path != rpath && rpath != "" {
		loader.cache[rpath] = resp
	}
	loader.mutex.Unlock()

	return resp
}

// load loads the go-package related to the given path. To succeed the path
// needs to either define a valid package (repository) path, or an absolute
// file path targeting a directory. If the path is relative or targeting a
// file, the package is resolved but will not provide the package repository
// path nor the standardized package name by default. The method compensates
// this partially be calculating default package names on a best effort basis.
func (loader *CachedLoader) load(path string) *PkgResponse {
	pkgs, err := packages.Load(loader.Config, path)
	if err != nil {
		err = NewErrLoading(path, err)
	} else {
		err = NewErrPackageParsing(path, pkgs)
	}
	return NewPkgResponse(path, pkgs, err)
}

// ifacesNamed looks up the named interfaces provided by the defined source.
// The name must be a legal identifier.
func (loader *CachedLoader) ifacesNamed(
	pkgs []*packages.Package, source *Type,
) ([]*IFace, error) {
	ifaces := []*IFace{}
	for _, pkg := range pkgs {
		name, iface, err := loader.iface(pkg, source, source.Name)
		if err == nil {
			ifaces = append(ifaces,
				NewIFace(NewType(name, pkg.Fset), iface))
		} else if errors.Is(err, ErrNoIFace) {
			return nil, err
		}
	}

	if len(ifaces) == 0 {
		return nil, NewErrNotFound(source, source.Name)
	}
	return ifaces, nil
}

// ifacesAny looks up all interfaces provided by the defined source. If a name
// pattern and a file is porvided, the interface must match the name pattern
// and reside in the matching file.
func (loader *CachedLoader) ifacesAny(
	pkgs []*packages.Package, source *Type,
) ([]*IFace, error) {
	matcher, err := source.Matcher()
	if err != nil {
		return nil, NewErrMatcherInvalid(source, err)
	}

	ifaces := []*IFace{}
	for _, pkg := range pkgs {
		for _, name := range pkg.Types.Scope().Names() {
			name, iface, err := loader.iface(pkg, source, name)
			if err != nil {
				continue
			}
			nsource := NewType(name, pkg.Fset)
			if matcher.Matches(nsource) {
				ifaces = append(ifaces,
					NewIFace(nsource, iface))
			}
		}
	}
	return ifaces, nil
}

// iface looks up a single interface from the given package that matches the
// provided interface name.
func (*CachedLoader) iface(
	pkg *packages.Package, source *Type, name string,
) (*types.TypeName, *types.Interface, error) {
	if object := pkg.Types.Scope().Lookup(name); object == nil {
		return nil, nil, NewErrNotFound(source, name)
	} else if iface, ok := object.Type().Underlying().(*types.Interface); !ok {
		return nil, nil, NewErrNoIFace(source, name)
	} else if named, ok := object.(*types.TypeName); !ok {
		return nil, nil, NewErrNoNameType(source, name)
	} else {
		return named, iface, nil
	}
}

// normalize normalizes the file path by appending the working directory of
// loader to relative paths and calculate the absolute path, since only
// absolute paths can resolve to valid packages.
func (loader *CachedLoader) normalize(path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(loader.Config.Dir, path)
	}
	path, _ = filepath.Abs(path)
	return path
}

// trimfile removes the last part of the file path, if it is a file or if the
// part does not exist. This is following the assumption that a valid target
// package needs to exist to generate code or read a package.
func (*CachedLoader) trimfile(path string) string {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return filepath.Dir(path)
	}
	return path
}
