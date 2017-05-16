// Package httper is a cli tool to implement http interface of a type.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mh-cbon/astutil"
	httper "github.com/mh-cbon/httper/lib"
	"github.com/mh-cbon/httper/utils"
)

var name = "httper"
var version = "0.0.0"

func main() {

	var help bool
	var h bool
	var ver bool
	var v bool
	var outPkg string
	var mode string
	flag.BoolVar(&help, "help", false, "Show help.")
	flag.BoolVar(&h, "h", false, "Show help.")
	flag.BoolVar(&ver, "version", false, "Show version.")
	flag.BoolVar(&v, "v", false, "Show version.")
	flag.StringVar(&outPkg, "p", os.Getenv("GOPACKAGE"), "Package name of the new code.")
	flag.StringVar(&mode, "mode", "std", "Generation mode.")

	flag.Parse()

	if ver || v {
		showVer()
		return
	}
	if help || h {
		showHelp()
		return
	}
	if mode != stdMode && mode != gorillaMode {
		showHelp()
		return
	}

	if flag.NArg() < 1 {
		panic("wrong usage")
	}
	args := flag.Args()

	out := ""
	if args[0] == "-" {
		args = args[1:]
		out = "-"
	}

	todos, err := utils.NewTransformsArgs(utils.GetPkgToLoad()).Parse(args)
	if err != nil {
		panic(err)
	}

	filesOut := utils.NewFilesOut("github.com/mh-cbon/" + name)

	for _, todo := range todos.Args {
		if todo.FromPkgPath == "" {
			log.Println("Skipped ", todo.FromTypeName)
			continue
		}

		fileOut := filesOut.Get(todo.ToPath)

		fileOut.PkgName = outPkg
		if fileOut.PkgName == "" {
			fileOut.PkgName = findOutPkg(todo)
		}

		if err := processType(mode, todo, fileOut); err != nil {
			log.Println(err)
		}
	}
	filesOut.Write(out)
}

func showVer() {
	fmt.Printf("%v %v\n", name, version)
}

func showHelp() {
	showVer()
	fmt.Println()
	fmt.Println("Usage")
	fmt.Println()
	fmt.Printf("	%v [-p name] [-mode name] [...types]\n\n", name)
	fmt.Printf("  types:  A list of types such as src:dst.\n")
	fmt.Printf("          A type is defined by its package path and its type name,\n")
	fmt.Printf("          [pkgpath/]name\n")
	fmt.Printf("          If the Package path is empty, it is set to the package name being generated.\n")
	// fmt.Printf("          If the Package path is a directory relative to the cwd, and the Package name is not provided\n")
	// fmt.Printf("          the package path is set to this relative directory,\n")
	// fmt.Printf("          the package name is set to the name of this directory.\n")
	fmt.Printf("          Name can be a valid type identifier such as TypeName, *TypeName, []TypeName \n")
	fmt.Printf("  -p:     The name of the package output.\n")
	fmt.Printf("  -mode:  The mode of generation to apply: std|gorilla (defaults to std).\n")
	fmt.Println()
}

func findOutPkg(todo utils.TransformArg) string {
	if todo.ToPkgPath != "" {
		prog := astutil.GetProgramFast(todo.ToPkgPath)
		if prog != nil {
			pkg := prog.Package(todo.ToPkgPath)
			return pkg.Pkg.Name()
		}
	}
	if todo.ToPkgPath == "" {
		prog := astutil.GetProgramFast(utils.GetPkgToLoad())
		if len(prog.Imported) < 1 {
			panic("impossible, add [-p name] option")
		}
		for _, p := range prog.Imported {
			return p.Pkg.Name()
		}
	}
	if strings.Index(todo.ToPkgPath, "/") > -1 {
		return filepath.Base(todo.ToPkgPath)
	}
	return todo.ToPkgPath
}

func processType(mode string, todo utils.TransformArg, fileOut *utils.FileOut) error {

	dest := &fileOut.Body
	srcName := todo.FromTypeName
	destName := todo.ToTypeName

	prog := astutil.GetProgramFast(todo.FromPkgPath)
	pkg := prog.Package(todo.FromPkgPath)
	foundMethods := astutil.FindMethods(pkg)

	srcConcrete := astutil.GetUnpointedType(srcName)
	// the json input must provide a key/value for each params.
	structType := astutil.FindStruct(pkg, srcConcrete)
	structComment := astutil.GetComment(prog, structType.Pos())
	// todo: might do better to send only annotations or do other improvemenets.
	structComment = makeCommentLines(structComment)

	fileOut.AddImport("io", "")
	fileOut.AddImport("net/http", "")
	fileOut.AddImport("strconv", "")
	fileOut.AddImport("github.com/mh-cbon/httper/lib", "httper")

	// cheat.
	fmt.Fprintf(dest, `var xxStrconvAtoi = strconv.Atoi
	var xxIoCopy = io.Copy
	var xxHTTPOk = http.StatusOK
	`)

	// Declare the new type
	fmt.Fprintf(dest, `
// %v is an httper of %v.
%v
type %v struct{
	embed %v
	cookier httper.CookieProvider
	dataer httper.DataerProvider
	sessioner httper.SessionProvider
	finalizer httper.Finalizer
}
		`, destName, srcName, structComment, destName, srcName)

	dstStar := astutil.GetPointedType(destName)

	// defiens the data provider factory
	factory := fmt.Sprintf("%T", getDataProviderFactory(mode))
	factory = astutil.GetUnpointedType(factory)
	sessionFactory := fmt.Sprintf("%T", getSessionProviderFactory(mode))
	sessionFactory = astutil.GetUnpointedType(sessionFactory)

	// Make the constructor
	fmt.Fprintf(dest, `// New%v constructs an httper of %v
func New%v(embed %v, finalizer httper.Finalizer) *%v {
	if finalizer == nil {
		finalizer = &httper.HTTPFinalizer{}
	}
	ret := &%v{
		embed: embed,
		cookier: &httper.CookieHelperProvider{},
		dataer: &%v{},
		sessioner: &%v{},
		finalizer: finalizer,
	}
  return ret
}
`, destName, srcName, destName, srcName, destName, destName, factory, sessionFactory)

	for _, m := range foundMethods[srcConcrete] {
		methodName := astutil.MethodName(m)
		paramNames := astutil.MethodParamNames(m)
		paramTypes := astutil.MethodParamTypes(m)

		// ensure it is desired to facade this method.
		if astutil.IsExported(methodName) == false {
			continue
		}
		if methodName == "UnmarshalJSON" || methodName == "MarshalJSON" {
			continue
		}

		comment := astutil.GetComment(prog, m.Pos())
		comment = makeCommentLines(comment)

		methodInvokation := ""

		lParamNames := strings.Split(paramNames, ",")
		lParamTypes := strings.Split(paramTypes, ",")
		for i, p := range lParamNames {
			p = strings.TrimSpace(p)
			paramType := strings.TrimSpace(lParamTypes[i])

			if p == reqBodyVarName {
				methodInvokation += fmt.Sprintf("%v :=	r.Body\n", reqBodyVarName)

			} else if paramType == "httper.Cookier" {
				methodInvokation += fmt.Sprintf("var %v %v\n", p, paramType)
				methodInvokation += fmt.Sprintf("%v = t.cookier.Make(w, r)\n", p)

			} else if (paramType == "http.ResponseWriter" && p == "w") || paramType == "*http.Request" && p == "r" {
				//skip

			} else if paramType == "httper.Sessionner" {
				methodInvokation += fmt.Sprintf("var %v %v\n", p, paramType)
				methodInvokation += fmt.Sprintf("%v = t.sessioner.Make(w, r)\n", p)

			} else if isConvetionnedParam(mode, p) {
				prefix := getParamConvention(mode, p)
				name := strings.ToLower(p[len(prefix):])

				methodInvokation += fmt.Sprintf("var %v %v\n", p, paramType)

				//handle prefixed data
				expr := fmt.Sprintf("t.dataer.Make(w,r).Get(%q, %q)", prefix, name)
				methodInvokation += convertedStr(p, expr, paramType)

			} else {
				methodInvokation += fmt.Sprintf("var %v %v\n", p, paramType)
			}
		}

		// proceed to the method invokcation on embed
		body := fmt.Sprintf(`
		  res, err := t.embed.%v(%v)
		  %v
			t.finalizer.HandleSuccess(w, res)
`, methodName, paramNames, handleErr("err"))

		fmt.Fprintf(dest, `// %v invoke %v.%v using the request body as a json payload.
			%v
func (t %v) %v(w http.ResponseWriter, r *http.Request) {
  %v
  %v
}`, methodName, srcName, methodName, comment, dstStar, methodName, methodInvokation, body)
		fmt.Fprintln(dest)
	}

	return nil
}

func makeCommentLines(s string) string {
	s = strings.TrimSpace(s)
	comment := ""
	for _, k := range strings.Split(s, "\n") {
		comment += "// " + k + "\n"
	}
	comment = strings.TrimSpace(comment)
	if comment == "" {
		comment = "//"
	}
	return comment
}

var gorillaMode = "gorilla"
var stdMode = "std"
var reqBodyVarName = "reqBody"

func isUsingConvetionnedParams(mode, params string) bool {
	lParams := strings.Split(params, ",")
	for _, param := range lParams {
		k := strings.Split(param, " ")
		if len(k) > 1 {
			varType := strings.TrimSpace(k[1])
			if varType == "http.ResponseWriter" {
				return true

			} else if varType == "*http.Request" {
				return true

			} else if varType == "httper.Cookier" {
				return true

			} else if varType == "httper.Sessionner" {
				return true
			}
		}
		varName := strings.TrimSpace(k[0])
		if isConvetionnedParam(mode, varName) {
			return true
		}
	}
	return false
}

func isConvetionnedParam(mode, varName string) bool {
	if varName == reqBodyVarName {
		return true
	}
	return getVarPrefix(mode, varName) != ""
}

func getParamConvention(mode, varName string) string {
	if varName == reqBodyVarName {
		return reqBodyVarName
	}
	return getVarPrefix(mode, varName)
}

func getSessionProviderFactory(mode string) httper.SessionProvider {
	var factory httper.SessionProvider
	if mode == stdMode {
		factory = &httper.VoidSessionProvider{}
	} else if mode == gorillaMode {
		factory = &httper.GorillaSessionProvider{}
	}
	return factory
}

func getDataProviderFactory(mode string) httper.DataerProvider {
	var factory httper.DataerProvider
	if mode == stdMode {
		factory = &httper.StdHTTPDataProvider{}
	} else if mode == gorillaMode {
		factory = &httper.GorillaHTTPDataProvider{}
	}
	return factory
}

func getDataProvider(mode string) *httper.DataProviderFacade {
	return getDataProviderFactory(mode).MakeEmpty().(*httper.DataProviderFacade)
}

func getVarPrefix(mode, varName string) string {
	ret := ""
	provider := getDataProvider(mode)
	for _, p := range provider.Providers {
		prefix := p.GetName()
		if strings.HasPrefix(varName, strings.ToLower(prefix)) {
			f := string(varName[len(prefix):][0])
			if f == strings.ToUpper(f) {
				ret = prefix
				break
			}
		} else if strings.HasPrefix(varName, strings.ToUpper(prefix)) {
			f := string(varName[len(prefix):][0])
			if f == strings.ToLower(f) {
				ret = prefix
				break
			}
		}
	}
	return ret
}

func convertedStr(toVarName, expr string, toType string) string {
	if toType == "int" {
		expr = convStrToInt(expr, toVarName)
	} else {
		expr = fmt.Sprintf("%v := %v\n", toVarName, expr)
	}
	return expr
}
func convStrToInt(fromVarName, toVarName string) string {
	methodInvokation := fmt.Sprintf("temp%v, err := strconv.Atoi(%v)\n", toVarName, fromVarName)
	methodInvokation += handleErr("err")
	methodInvokation += fmt.Sprintf("%v = temp%v\n", toVarName, toVarName)
	return methodInvokation
}
func handleErr(errVarName string) string {
	methodInvokation := fmt.Sprintf(`if err != nil && t.finalizer.HandleError(%v,w,r) {
return
}
`, errVarName)
	return methodInvokation
}
func paramType(params string) string {
	x := strings.Split(params, ",")
	return x[len(x)-1]
}
