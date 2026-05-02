package enum

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// generateConfig holds construction-time options for the generator.
type generateConfig struct {
	outputSuffix   string
	withJSON       bool
	withSQL        bool
	withExhaustive bool
}

// GenerateOption configures the code generator at call time.
type GenerateOption func(*generateConfig)

// WithOutputSuffix overrides the generated file name suffix.
// Default: "_enum.gen.go".
func WithOutputSuffix(suffix string) GenerateOption {
	return func(c *generateConfig) {
		if suffix != "" {
			c.outputSuffix = suffix
		}
	}
}

// WithJSON enables or disables generation of MarshalJSON/UnmarshalJSON methods.
// Default: true.
func WithJSON(enabled bool) GenerateOption {
	return func(c *generateConfig) {
		c.withJSON = enabled
	}
}

// WithSQL enables or disables generation of driver.Valuer and sql.Scanner methods.
// Default: true.
func WithSQL(enabled bool) GenerateOption {
	return func(c *generateConfig) {
		c.withSQL = enabled
	}
}

// WithExhaustive enables or disables generation of the Exhaustive() method.
// Default: true.
func WithExhaustive(enabled bool) GenerateOption {
	return func(c *generateConfig) {
		c.withExhaustive = enabled
	}
}

func applyGenerateOptions(opts []GenerateOption) *generateConfig {
	c := &generateConfig{
		outputSuffix:   "_enum.gen.go",
		withJSON:       true,
		withSQL:        true,
		withExhaustive: true,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// buildImports assembles the import paths needed by the generated file.
// fmt is included only when at least one generated method requires it.
func buildImports(c *generateConfig, info *EnumInfo) []string {
	var imports []string

	// fmt is needed by Parse<Type> (HasString) and Scan (WithSQL).
	needsFmt := info.HasString || c.withSQL
	if needsFmt {
		imports = append(imports, `"fmt"`)
	}
	if c.withJSON {
		imports = append(imports, `"encoding/json"`)
	}
	if c.withSQL {
		imports = append(imports, `"database/sql/driver"`)
	}
	return imports
}

// EnumInfo holds the data extracted from the source files used to generate code.
type EnumInfo struct {
	PkgName        string
	TypeName       string
	Kind           string // "string" or "int"
	Constants      []string
	Values         []string // raw constant values
	HasString      bool
	HasInt         bool
	WithJSON       bool
	WithSQL        bool
	WithExhaustive bool
	Imports        []string
}

// Generate is the main entry point invoked by the command.
func Generate(dir string, typeName string, opts ...GenerateOption) error {
	c := applyGenerateOptions(opts)

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), c.outputSuffix)
	}, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parser error: %w", err)
	}

	var info EnumInfo
	var pkgName string
	var allDecls []ast.Decl

	for _, pkg := range pkgs {
		if pkgName == "" {
			pkgName = pkg.Name
		}
		for _, f := range pkg.Files {
			allDecls = append(allDecls, f.Decls...)
		}
	}

	if pkgName == "" {
		return fmt.Errorf("no Go files found in %s", dir)
	}

	info.PkgName = pkgName
	info.TypeName = typeName
	info.WithJSON = c.withJSON
	info.WithSQL = c.withSQL
	info.WithExhaustive = c.withExhaustive

	if info.Kind == "int" {
		for _, v := range info.Values {
			if _, err := fmt.Sscanf(v, "%d", new(int)); err == nil {
				info.HasInt = true
			} else {
				info.HasString = true
			}
		}
	}

	for _, decl := range allDecls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valueSpec.Names) == 0 {
				continue
			}

			if valueSpec.Type != nil {
				ident, ok := valueSpec.Type.(*ast.Ident)
				if !ok || ident.Name != typeName {
					continue
				}
			} else {
				continue
			}

			for i, name := range valueSpec.Names {
				info.Constants = append(info.Constants, name.Name)
				if len(valueSpec.Values) > i {
					if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
						info.Values = append(info.Values, strings.Trim(lit.Value, `"`))
					} else {
						info.Values = append(info.Values, "")
					}
				} else {
					info.Values = append(info.Values, "")
				}
			}
		}
	}

	if len(info.Constants) == 0 {
		return fmt.Errorf("no constants of type %s found", typeName)
	}

	if len(info.Values) > 0 && info.Values[0] != "" {
		if _, err := fmt.Sscanf(info.Values[0], "%d", new(int)); err == nil {
			info.Kind = "int"
			info.HasInt = true
		} else {
			info.Kind = "string"
			info.HasString = true
		}
	} else {
		info.Kind = "string"
		info.HasString = true
	}

	if info.Kind == "int" {
		for _, v := range info.Values {
			if _, err := fmt.Sscanf(v, "%d", new(int)); err == nil {
				info.HasInt = true
			} else {
				info.HasString = true
			}
		}
	}

	info.Imports = buildImports(c, &info)
	return generateCode(dir, c.outputSuffix, info)
}

const enumTemplate = `// Code generated by Typed Enum Generator; DO NOT EDIT.

package {{.PkgName}}

{{if .Imports}}
import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{end}}

// ===== Generated methods for enum {{.TypeName}} =====

// IsValid reports whether the value is one of the declared constants.
func (t {{.TypeName}}) IsValid() bool {
	switch t {
{{- $constants := .Constants}}
	case {{range $i, $c := $constants}}{{if $i}}, {{end}}{{$c}}{{end}}:
		return true
	default:
		return false
	}
}

// Values returns every constant of the enum.
func ({{.TypeName}}) Values() []{{.TypeName}} {
	return []{{.TypeName}}{
{{- range .Constants}}
		{{.}},
{{- end}}
	}
}

{{if .HasString}}
// String returns the underlying string value.
func (t {{.TypeName}}) String() string {
	return string(t)
}

// Parse{{.TypeName}} converts a string into the enum value.
func Parse{{.TypeName}}(s string) ({{.TypeName}}, error) {
	switch s {
{{- range $i, $c := .Constants}}
	case "{{index $.Values $i}}":
		return {{$c}}, nil
{{- end}}
	default:
		return {{(index .Constants 0)}}, fmt.Errorf("invalid {{.TypeName}}: %s", s)
	}
}
{{end}}

{{if .WithJSON}}
// MarshalJSON implements the json.Marshaler interface.
func (t {{.TypeName}}) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *{{.TypeName}}) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := Parse{{.TypeName}}(s)
	if err != nil {
		return err
	}
	*t = val
	return nil
}
{{end}}

{{if .WithSQL}}
// Value implements the driver.Valuer interface for database/sql.
func (t {{.TypeName}}) Value() (driver.Value, error) {
	return string(t), nil
}

// Scan implements the sql.Scanner interface.
func (t *{{.TypeName}}) Scan(value interface{}) error {
	if value == nil {
		*t = {{index .Constants 0}}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	val, err := Parse{{.TypeName}}(s)
	if err != nil {
		return err
	}
	*t = val
	return nil
}
{{end}}

// {{.TypeName}}FromIndex returns the enum value at the given index.
func {{.TypeName}}FromIndex(idx int) {{.TypeName}} {
	switch idx {
{{- range $i, $c := .Constants}}
	case {{$i}}:
		return {{$c}}
{{- end}}
	default:
		return {{(index .Constants 0)}}
	}
}

// Index returns the 0-based position of the constant.
func (t {{.TypeName}}) Index() int {
	switch t {
{{- range $i, $c := .Constants}}
	case {{$c}}:
		return {{$i}}
{{- end}}
	default:
		return -1
	}
}

{{if .WithExhaustive}}
// Exhaustive forces every constant to be handled at the call site.
// Panics if a new constant is added without updating this method.
func (t {{.TypeName}}) Exhaustive() string {
	switch t {
{{- range $i, $c := .Constants}}
	case {{$c}}:
		return "{{index $.Values $i}}"
{{- end}}
	default:
		panic("unreachable: enum {{.TypeName}} has unhandled constant")
	}
}
{{end}}
`

func generateCode(dir, suffix string, info EnumInfo) error {
	outputName := strings.ToLower(info.TypeName) + suffix
	outputPath := filepath.Join(dir, outputName)

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer f.Close()

	tmpl, err := template.New("enum").Parse(enumTemplate)
	if err != nil {
		return fmt.Errorf("template error: %w", err)
	}

	return tmpl.Execute(f, info)
}
