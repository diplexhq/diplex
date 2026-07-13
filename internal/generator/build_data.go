package generator

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/diplexhq/diplex/internal/domain"
	"github.com/diplexhq/diplex/internal/utils"
)

func lookupStd(pkg string) bool {
	root := pkg
	if before, _, ok := strings.Cut(pkg, "/"); ok {
		root = before
	}

	_, ok := stdPkgs[root]

	return ok
}

var stdPkgs = map[string]struct{}{
	"archive":   {},
	"bufio":     {},
	"builtin":   {},
	"bytes":     {},
	"compress":  {},
	"container": {},
	"context":   {},
	"crypto":    {},
	"database":  {},
	"debug":     {},
	"embed":     {},
	"encoding":  {},
	"errors":    {},
	"expvar":    {},
	"flag":      {},
	"fmt":       {},
	"go":        {},
	"hash":      {},
	"html":      {},
	"image":     {},
	"index":     {},
	"io":        {},
	"log":       {},
	"math":      {},
	"mime":      {},
	"net":       {},
	"os":        {},
	"path":      {},
	"reflect":   {},
	"regexp":    {},
	"runtime":   {},
	"sort":      {},
	"strconv":   {},
	"strings":   {},
	"sync":      {},
	"syscall":   {},
	"testlog":   {},
	"testing":   {},
	"text":      {},
	"time":      {},
	"unicode":   {},
	"unsafe":    {},
}

type tmplDataProvider struct {
	Field  string
	Call   string
	Args   []tmplDataProviderArg
	Result string
	Error  bool
}

type tmplDataProviderArg struct {
	Type           string
	CollectionType string
	Providers      []string
}

type tmplDataFacadeMethod struct {
	Name      string
	Result    string
	Type      string
	Providers []string
}

type tmplData struct {
	imports   []string
	providers []tmplDataProvider
	facade    []tmplDataFacadeMethod
	FacadePkg string
}

type buildData struct {
	facadeID     domain.InterfaceID
	providerUse  map[string]struct{}
	pkgAlias     map[string]string
	providerList []*domain.Provider
	resolvedData domain.ResolvedData
	tmplData     tmplData
}

func (dig *DiGenerator) newBuildData(facadeID domain.InterfaceID, resolvedData domain.ResolvedData) *buildData {
	return &buildData{
		facadeID:     facadeID,
		providerUse:  make(map[string]struct{}),
		pkgAlias:     make(map[string]string),
		resolvedData: resolvedData,
	}
}

func (dig *DiGenerator) buildData(facadeID domain.InterfaceID, resolvedData domain.ResolvedData) *buildData {
	data := dig.newBuildData(facadeID, resolvedData)

	dig.addPkg(data, "sync")

	for _, providers := range resolvedData.Providers {
		for _, provider := range providers.Providers {
			dig.addProvider(data, provider)
		}
	}

	dig.buildDataPkg(data)
	dig.buildDataProvider(data)
	dig.buildDataFacade(data)

	return data
}

func (dig *DiGenerator) addProvider(d *buildData, provider *domain.Provider) {
	if _, ok := d.providerUse[provider.Id()]; ok {
		return
	}

	d.providerUse[provider.Id()] = struct{}{}
	d.providerList = append(d.providerList, provider)
	d.pkgAlias[provider.Pkg] = provider.Pkg
	dig.addPkgsFromParam(d, provider.Result)
	dig.addPkgsFromParam(d, domain.Parameter(provider.Name))
}

func (dig *DiGenerator) addPkg(d *buildData, pkg string) {
	d.pkgAlias[pkg] = pkg
}

func (dig *DiGenerator) addPkgsFromParam(d *buildData, param domain.Parameter) {
	for i := 0; i < len(param); {
		start := i
		dot := i

		for i < len(param) && utils.IsIdentChar(param[i]) {
			if param[i] == '.' {
				dot = i
			}

			i++
		}

		if i > start {
			if dot > start {
				dig.addPkg(d, string(param)[start:dot])
			}
		} else {
			i++
		}
	}
}

func (dig *DiGenerator) buildDataPkg(data *buildData) {
	var (
		standard []string
		others   []string
	)

	local := make([]string, 0, len(data.pkgAlias))

	for pkg := range data.pkgAlias {
		switch {
		case strings.HasPrefix(pkg, dig.cfg.Module()):
			local = append(local, pkg)
		case lookupStd(pkg):
			standard = append(standard, pkg)
		default:
			others = append(others, pkg)
		}
	}

	// add facade package (strip "PackageName.InterfaceName" → "PackageName")
	facadePkg := string(data.facadeID)
	if dot := strings.LastIndex(facadePkg, "."); dot > 0 {
		facadePkg = facadePkg[:dot]
	}
	local = append(local, facadePkg)

	used := make(map[string]struct{})

	data.tmplData.imports = make([]string, 0, len(standard)+len(local)+len(others)+3)
	for _, pkgs := range [][]string{standard, others, local} {
		if len(pkgs) == 0 {
			continue
		}

		if len(data.tmplData.imports) > 0 {
			data.tmplData.imports = append(data.tmplData.imports, "")
		}

		sort.Strings(pkgs)

		for _, pkg := range pkgs {
			alias := dig.idUniq(dig.simplePkg(pkg), used)

			data.pkgAlias[pkg] = alias
			if alias == filepath.Base(pkg) {
				data.tmplData.imports = append(data.tmplData.imports, `"`+pkg+`"`)
			} else {
				data.tmplData.imports = append(data.tmplData.imports, alias+` "`+pkg+`"`)
			}
		}
	}

	data.tmplData.FacadePkg = data.pkgAlias[facadePkg] + "." + data.facadeID.LocalName()
}

func (dig *DiGenerator) buildDataProvider(data *buildData) {
	sort.Slice(data.providerList, func(i, j int) bool {
		if data.providerList[i].Pkg == data.providerList[j].Pkg {
			return data.providerList[i].Name < data.providerList[j].Name
		}

		return data.providerList[i].Pkg < data.providerList[j].Pkg
	})

	data.tmplData.providers = make([]tmplDataProvider, 0, len(data.providerList))
	for _, provider := range data.providerList {
		pkgAlias := data.pkgAlias[provider.Pkg]
		name := dig.replacePkgAlias(data.pkgAlias, provider.Name)
		data.tmplData.providers = append(data.tmplData.providers, tmplDataProvider{
			Field:  utils.SanitizeIdent(pkgAlias + name),
			Call:   pkgAlias + "." + name,
			Args:   dig.buildDataProviderArgs(data, provider.Arguments),
			Error:  provider.Error,
			Result: dig.replacePkgAlias(data.pkgAlias, string(provider.Result)),
		})
	}
}

func (dig *DiGenerator) buildDataProviderArgs(data *buildData, arguments []domain.Parameter) []tmplDataProviderArg {
	if len(arguments) == 0 {
		return nil
	}

	res := make([]tmplDataProviderArg, 0, len(arguments))
	for _, arg := range arguments {
		providers := data.resolvedData.Providers[arg]
		res = append(res, tmplDataProviderArg{
			Type:           dig.replacePkgAlias(data.pkgAlias, string(arg)),
			CollectionType: providers.CollectionType,
			Providers:      dig.buildDataTmplProvider(data, providers),
		})
	}

	return res
}

func (dig *DiGenerator) replacePkgAlias(pkgAlias map[string]string, str string) string {
	var buf strings.Builder
	buf.Grow(len(str))

	for i := 0; i < len(str); {
		start, dot := i, i
		for i < len(str) && utils.IsIdentChar(str[i]) {
			if str[i] == '.' {
				dot = i
			}

			i++
		}

		if i > start {
			if dot > start {
				if alias, ok := pkgAlias[str[start:dot]]; ok {
					buf.WriteString(alias)
					buf.WriteString(str[dot:i])

					continue
				}
			}

			buf.WriteString(str[start:i])
		} else {
			buf.WriteByte(str[i])
			i++
		}
	}

	return buf.String()
}

func (dig *DiGenerator) buildDataFacade(data *buildData) {
	methods := data.resolvedData.Facades[data.facadeID].Methods

	data.tmplData.facade = make([]tmplDataFacadeMethod, 0, len(methods))
	for name, method := range methods {
		resultArg := method.Results[0]
		providers := data.resolvedData.Providers[resultArg]
		data.tmplData.facade = append(data.tmplData.facade, tmplDataFacadeMethod{
			Name:      string(name),
			Type:      providers.CollectionType,
			Providers: dig.buildDataTmplProvider(data, providers),
			Result:    dig.replacePkgAlias(data.pkgAlias, string(resultArg)),
		})
	}

	sort.Slice(data.tmplData.facade, func(i, j int) bool {
		return data.tmplData.facade[i].Name < data.tmplData.facade[j].Name
	})
}

func (dig *DiGenerator) buildDataTmplProvider(data *buildData, providers domain.ProviderCollection) []string {
	res := make([]string, 0, len(providers.Providers))
	for _, provider := range providers.Providers {
		field := data.pkgAlias[provider.Pkg] + dig.replacePkgAlias(data.pkgAlias, provider.Name)
		res = append(res, utils.SanitizeIdent(field))
	}

	return res
}
