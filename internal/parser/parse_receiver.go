package parser

import (
	"go/ast"
	"strconv"

	"github.com/diplexhq/diplex/internal/domain"
	astStringer "github.com/diplexhq/diplex/internal/parser/ast_stringer"
	"github.com/diplexhq/diplex/internal/utils"
)

// parseReceiver parse method on receiver
//
// Supports concrete generic instantiations: InMemoryStore[Order], Cache[string, SessionData].
func (fp *Parser) parseReceiver(funcDecl *ast.FuncDecl, imports map[string]string, pkg string, state *parseState) {
	if funcDecl.Recv == nil || !utils.IsExported(funcDecl.Name.Name) {
		return
	}

	recv := fp.unStar(funcDecl.Recv.List[0].Type)

	baseName, ok := fp.baseIdent(recv)
	if !ok || !utils.IsExported(baseName) {
		return
	}

	// Build the struct key for the implementation.
	// For non-generics: pkg.BaseName (e.g. diplex/testdata/order.OrderServiceImpl)
	// For generics: full instantiated type (e.g. diplex/testdata/generic.InMemoryStore[diplex/testdata/generic.Order])
	suffix, genericAlias := fp.parseReceiverGenerics(recv)
	interfaceID := domain.InterfaceID(pkg + "." + baseName + suffix)

	arguments, _ := astStringer.FieldsToStrings(funcDecl.Type.Params, imports, pkg, genericAlias)
	results, _ := astStringer.FieldsToStrings(funcDecl.Type.Results, imports, pkg, genericAlias)

	method := domain.MethodContract{
		Arguments: arguments,
		Results:   results,
	}
	methodName := domain.FunctionName(funcDecl.Name.Name)

	state.mu.Lock()

	info, ok := state.interfaces[interfaceID]
	if !ok {
		info = domain.InterfaceInfo{
			Methods:  make(domain.MethodMap),
			RealType: true, // struct method implementation
		}
	}

	info.Methods[methodName] = method
	state.interfaces[interfaceID] = info
	state.mu.Unlock()
	fp.log.Debug("receiver parsed", "interface", interfaceID, "method", methodName)
}

func (fp *Parser) parseReceiverGenerics(recv ast.Expr) (string, map[string]string) {
	switch recv := recv.(type) {
	case *ast.IndexExpr:
		name := recv.Index.(*ast.Ident).Name

		return "[T]", map[string]string{
			name: "T",
		}
	case *ast.IndexListExpr:
		aliasMap := make(map[string]string)

		for i, index := range recv.Indices {
			name := index.(*ast.Ident).Name
			aliasMap[name] = "T" + strconv.FormatInt(int64(i), 10)
		}

		return utils.TList(len(aliasMap)), aliasMap
	}

	return "", nil
}
