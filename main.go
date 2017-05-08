// Package httper is a cli tool to implement http interface of a type.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mh-cbon/astutil"
	httper "github.com/mh-cbon/httper/lib"
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

	if flag.NArg() < 2 {
		panic("wrong usage")
	}
	args := flag.Args()

	pkgToLoad := getPkgToLoad()
	dest := os.Stdout

	o := args[0]
	restargs := args[1:]

	prog := astutil.GetProgram(pkgToLoad).Package(pkgToLoad)

	foundMethods := astutil.FindMethods(prog)

	if o != "-" {
		f, err := os.Create(o)
		if err != nil {
			panic(err)
		}
		dest = f
		defer func() {
			f.Close()
			cmd := exec.Command("go", "fmt", args[0])
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}()
	}

	fmt.Fprintf(dest, "package %v\n\n", outPkg)
	fmt.Fprintln(dest, `// file generated by`)
	fmt.Fprintf(dest, "// github.com/mh-cbon/%v\n", name)
	fmt.Fprintln(dest, `// do not edit`)
	fmt.Fprintln(dest, "")
	fmt.Fprintf(dest, "import (\n")
	fmt.Fprintf(dest, "	%q\n", "io")
	fmt.Fprintf(dest, "	%q\n", "net/http")
	fmt.Fprintf(dest, "	%q\n", "strconv")
	fmt.Fprintf(dest, "	%v %q\n", "httper", "github.com/mh-cbon/httper/lib")
	fmt.Fprintf(dest, ")\n")
	fmt.Fprintf(dest, "\n\n")
	// cheat.
	fmt.Fprintf(dest, `var xxStrconvAtoi = strconv.Atoi
var xxIoCopy = io.Copy
var xxHTTPOk = http.StatusOK
`)

	for _, todo := range restargs {
		y := strings.Split(todo, ":")
		if len(y) != 2 {
			panic("wrong name " + todo)
		}

		res := processType(mode, y[1], y[0], foundMethods)
		io.Copy(dest, &res)
	}
}

func showVer() {
	fmt.Printf("%v %v\n", name, version)
}

func showHelp() {
	showVer()
	fmt.Println()
	fmt.Println("Usage")
	fmt.Println()
	fmt.Printf("	%v [-p name] [-mode name] [out] [...types]\n\n", name)
	fmt.Printf("	out:   Output destination of the results, use '-' for stdout.\n")
	fmt.Printf("	types: A list of types such as src:dst.\n")
	fmt.Printf("	-p:    The name of the package output.\n")
	fmt.Printf("	-mode: The mode of generation to apply: std|gorilla (defaults to std).\n")
	fmt.Println()
}

func processType(mode, destName, srcName string, foundMethods map[string][]*ast.FuncDecl) bytes.Buffer {

	var b bytes.Buffer
	dest := &b

	// Declare the new type
	fmt.Fprintf(dest, `
// %v is an httper of %v.
type %v struct{
	embed %v
	cookier httper.CookieProvider
	dataer httper.DataerProvider
	sessioner httper.SessionProvider
}
		`, destName, srcName, destName, srcName)

	dstStar := astutil.GetPointedType(destName)
	srcConcrete := astutil.GetUnpointedType(srcName)
	hasHandleError := methodsContains(srcConcrete, "HandleError", foundMethods)
	hasHandleSuccess := methodsContains(srcConcrete, "HandleSuccess", foundMethods)

	// defiens the data provider factory
	factory := fmt.Sprintf("%T", getDataProviderFactory(mode))
	factory = astutil.GetUnpointedType(factory)
	sessionFactory := fmt.Sprintf("%T", getSessionProviderFactory(mode))
	sessionFactory = astutil.GetUnpointedType(sessionFactory)

	// Make the constructor
	fmt.Fprintf(dest, `// New%v constructs an httper of %v
func New%v(embed %v) *%v {
	ret := &%v{
		embed: embed,
		cookier: &httper.CookieHelperProvider{},
		dataer: &%v{},
		sessioner: &%v{},
	}
  return ret
}
`,
		destName, srcName, destName, srcName, destName, destName, factory, sessionFactory)

	// Add an error handler method
	if hasHandleError {
		fmt.Fprintf(dest, `// HandleError calls for embed.HandleError method.
func (t %v) HandleError(err error, w http.ResponseWriter, r *http.Request)bool{
	if err == nil {
		return false
	}
		return t.embed.HandleError(err, w, r)
}
`,
			dstStar)

	} else {
		fmt.Fprintf(dest, `// HandleError returns http 500 and prints the error.
func (t %v) HandleError(err error, w http.ResponseWriter, r *http.Request)bool{
	if err == nil {
		return false
	}
	w.WriteHeader(http.StatusInternalServerError)
	io.WriteString(w, err.Error())
	return true
}
`,
			dstStar)
	}

	// Add a success handler method
	if hasHandleSuccess {
		fmt.Fprintf(dest, `// HandleSuccess calls for embed.HandleSuccess method.
func (t %v) HandleSuccess(w http.ResponseWriter, r io.Reader) error {
	return t.embed.HandleSuccess(w, r)
}
`,
			dstStar)

	} else {
		fmt.Fprintf(dest, `// HandleSuccess prints http 200 and prints r.
func (t %v) HandleSuccess(w http.ResponseWriter, r io.Reader) error {
	w.WriteHeader(http.StatusOK)
	_, err := io.Copy(w, r)
	return err
}
`,
			dstStar)
	}
	fmt.Fprintln(dest)

	for _, m := range foundMethods[srcConcrete] {
		methodName := astutil.MethodName(m)
		// params := astutil.MethodParams(m)
		paramNames := astutil.MethodParamNames(m)
		paramTypes := astutil.MethodParamTypes(m)

		// ensure it is desired to facade this method.
		if astutil.IsExported(methodName) == false {
			continue
		}
		if methodName == "HandleError" {
			continue
		}
		if methodName == "HandleSuccess" {
			continue
		}

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
		  t.HandleSuccess(w, res)
`, methodName, paramNames, handleErr("err"))

		fmt.Fprintf(dest, `// %v invoke %v.%v using the request body as a json payload.
func (t %v) %v(w http.ResponseWriter, r *http.Request) {
  %v
  %v
}`,
			methodName, srcName, methodName, dstStar, methodName, methodInvokation, body)
		fmt.Fprintln(dest)
	}

	return b
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

func methodsContains(typeName, search string, methods map[string][]*ast.FuncDecl) bool {
	if funList, ok := methods[typeName]; ok {
		for _, fun := range funList {
			if astutil.MethodName(fun) == search {
				return true
			}
		}
	}
	return false
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
	methodInvokation := fmt.Sprintf(`if t.HandleError(%v,w,r) {
return
}
`, errVarName)
	return methodInvokation
}
func paramsHas(params string, what string) bool {
	return strings.Index(params, what) > -1 //todo: can do better.
}
func paramsLen(params string) int {
	return len(strings.Split(params, ","))
}
func paramType(params string) string {
	x := strings.Split(params, ",")
	return x[len(x)-1]
}

func getPkgToLoad() string {
	gopath := filepath.Join(os.Getenv("GOPATH"), "src")
	pkgToLoad, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pkgToLoad[len(gopath)+1:]
}
