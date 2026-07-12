package generator

import (
	_ "embed"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/diplexhq/diplex/internal/utils"
)

//go:embed tmpl/head.tmpl
var headTmpl string

//go:embed tmpl/import.tmpl
var importsTmpl string

//go:embed tmpl/facade.tmpl
var facadeTmpl string

//go:embed tmpl/provider.tmpl
var providerTmpl string

func (dig *DiGenerator) render(data *buildData, name, fname string) {
	out := utils.Must(os.Create(dig.cfg.OutputDir() + "/" + fname + ".go")) // nolint:gosec

	defer func() { _ = out.Close() }()

	dig.renderHead(out)
	dig.renderImports(out, data.tmplData)
	dig.renderFacade(out, name, data.tmplData)
	dig.renderProvider(out, name, data.tmplData)
}

func (dig *DiGenerator) renderHead(out io.Writer) {
	tmpl := utils.Must(template.New("head").Parse(headTmpl))
	utils.NoErr(tmpl.Execute(out, struct{ PackageName string }{PackageName: filepath.Base(dig.cfg.OutputDir())}))
}

func (dig *DiGenerator) renderImports(out io.Writer, tmplData tmplData) {
	tmpl := utils.Must(template.New("imports").Parse(importsTmpl))
	utils.NoErr(tmpl.Execute(out, struct {
		Imports []string
	}{
		Imports: tmplData.imports,
	}))
}

func (dig *DiGenerator) renderFacade(out io.Writer, name string, tmplData tmplData) {
	tmpl := utils.Must(template.New("facade").Parse(facadeTmpl))
	utils.NoErr(tmpl.Execute(out, struct {
		Name   string
		Facade []tmplDataFacadeMethod
	}{
		Name:   name,
		Facade: tmplData.facade,
	}))
}

func (dig *DiGenerator) renderProvider(out io.Writer, name string, tmplData tmplData) {
	tmpl := utils.Must(template.New("provider").Parse(providerTmpl))
	utils.NoErr(tmpl.Execute(out, struct {
		Name      string
		Providers []tmplDataProvider
	}{
		Name:      name,
		Providers: tmplData.providers,
	}))
}
