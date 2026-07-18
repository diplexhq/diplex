package domain

import (
	"strings"
)

// SourceFile is the absolute path to a Go source file.

type (
	SourceFile   string
	Parameter    string
	FunctionName string
	InterfaceID  string
	ProviderID   string

	MethodMap map[FunctionName]MethodContract

	Interfaces map[InterfaceID]InterfaceInfo
	Providers  map[ProviderID]*Provider
)

// InterfaceInfo holds the methods for an interface or a concrete type implementation.
// RealType is true for struct method implementations, false for declared interfaces.
type InterfaceInfo struct {
	Methods  MethodMap
	RealType bool
}

func (i InterfaceID) LocalName() string {
	split := strings.Split(string(i), ".")
	return split[len(split)-1]
}

// SourceFiles is a channel of source files to process.
type SourceFiles chan SourceFile

// ParsedData holds the intermediate representation of dependencies extracted from parsed Go files.
type ParsedData struct {
	Providers  Providers
	Interfaces Interfaces
}

type ProviderCollection struct {
	CollectionType string
	Providers      []*Provider
}

type ResolvedFacade map[FunctionName]ResolvedFacadeMethod

type ResolvedFacadeMethod struct {
	Result   Parameter
	Provider *Provider
}

type ResolvedProvider struct {
	Provider          *Provider
	ArgumentProviders []ProviderCollection
}

// ResolvedData holds the final resolved dependency graph with provider collections per parameter and facades.
type ResolvedData struct {
	ResolvedFacades   map[InterfaceID]ResolvedFacade
	ResolvedProviders map[ProviderID]ResolvedProvider
}

// MethodContract describes a single method signature.
type MethodContract struct {
	Arguments []Parameter
	Results   []Parameter
}

// Provider describes a service constructor function.
type Provider struct {
	ID         ProviderID
	Pkg        string
	Name       string
	Arguments  []Parameter
	ArgNames   []string
	Result     Parameter
	ResultName string
	Generic    map[string][]string // Maps aligned generic alias (T, T1, ...) → constraint strings
	Error      bool
}
