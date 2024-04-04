package generators

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/ettle/strcase"

	"github.com/diagram-code-generator/template/internal/utils"
)

// ErrUnsupportedFileType is an error indicating that the file type is not supported.
var ErrUnsupportedFileType = errors.New("unsupported file type")

// FormaterByExtMap is a map associating file extensions with formatting functions.
type FormaterByExtMap map[string]func(outputFile string) error

// TemplateGenerator struct holds the configuration and methods for generating templates.
type TemplateGenerator struct {
	funcs         template.FuncMap
	formaterByExt FormaterByExtMap
}

// Option is a functional option to configure TemplateGenerator.
type Option func(*TemplateGenerator)

// NewTemplateGenerator creates a new TemplateGenerator instance with the provided options.
func NewTemplateGenerator(opts ...Option) *TemplateGenerator {
	tg := &TemplateGenerator{
		funcs: template.FuncMap{
			"ToCamel":  strcase.ToCamel,
			"ToKebab":  strcase.ToKebab,
			"ToLower":  strings.ToLower,
			"ToPascal": strcase.ToPascal,
			"ToSpace":  func(s string) string { return strings.ReplaceAll(strcase.ToKebab(s), "-", " ") },
			"ToSnake":  strcase.ToSnake,
			"ToUpper":  strings.ToUpper,
		},
		formaterByExt: FormaterByExtMap{
			".go": utils.GoFormat,
			".tf": utils.TerraformFormat,
		},
	}

	for _, opt := range opts {
		opt(tg)
	}

	return tg
}

// WithExtraFuncs adds extra template functions to the TemplateGenerator.
func WithExtraFuncs(funcs template.FuncMap) Option {
	return func(tg *TemplateGenerator) {
		for k, v := range funcs {
			tg.funcs[k] = v
		}
	}
}

// WithExtraFormaterByExt adds extra file format functions to the TemplateGenerator.
func WithExtraFormaterByExt(formaterByExt FormaterByExtMap) Option {
	return func(tg *TemplateGenerator) {
		for k, v := range formaterByExt {
			tg.formaterByExt[k] = v
		}
	}
}

// Build executes the provided templateContent using the data supplied and returns the resulting string.
func (tg *TemplateGenerator) Build(data any, templateName, templateContent string) (string, error) {
	tmpl, err := tg.buildAndParseTemplate(templateName, templateContent)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	var output bytes.Buffer

	err = tmpl.Execute(&output, data)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	return output.String(), nil
}

// GenerateFile generates a file using the provided template and data, and writes the output to the specified
// outputFile.
func (tg *TemplateGenerator) GenerateFile(templatesMap map[string]string, fileName, fileTmpl, outputFile string, data any) error {
	var (
		tmpl     string
		tmplName = fmt.Sprintf("%s-template", strings.ReplaceAll(fileName, ".", "-"))
	)

	if fileTmpl == "" {
		tmpl = templatesMap[fileName]
	} else {
		tmpl = fileTmpl
	}

	if err := tg.buildFile(data, tmplName, tmpl, outputFile); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := tg.formatFileBasedOnExt(fileName, outputFile); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// GenerateFiles generates multiple files using the provided templates and data, and writes the outputs to the specified
// output directory.
func (tg *TemplateGenerator) GenerateFiles(
	defaultTemplatesMap map[string]string, templatesMap map[string]string, data any, output string,
) error {
	mergedTemplates := map[string]string{}
	var errs []error

	for filename, tmpl := range defaultTemplatesMap {
		mergedTemplates[filename] = tmpl
	}

	for filename, tmpl := range templatesMap {
		mergedTemplates[filename] = tmpl
	}

	for filename, fileTmpl := range mergedTemplates {
		tmplName := fmt.Sprintf("%s-template", strings.ReplaceAll(filename, ".", "-"))

		outputFile := path.Join(output, filename)

		err := tg.buildFile(data, tmplName, fileTmpl, outputFile)
		if err != nil {
			errs = append(errs, err)
		}

		err = tg.formatFileBasedOnExt(filename, outputFile)
		if err != nil && !errors.Is(err, ErrUnsupportedFileType) {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

// buildAndParseTemplate builds and parses a template with the given name and content.
func (tg *TemplateGenerator) buildAndParseTemplate(name, content string) (*template.Template, error) {
	tmpl, err := template.New(name).Funcs(tg.funcs).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return tmpl, nil
}

// buildFile builds a file from the given data and template content, writing it to the specified output path.
func (tg *TemplateGenerator) buildFile(data any, templateName, templateContent, outputPath string) error {
	tmpl, err := tg.buildAndParseTemplate(templateName, templateContent)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer output.Close()

	err = tmpl.Execute(output, data)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// formatFileBasedOnExt formats a file based on its extension using the corresponding formatter.
func (tg *TemplateGenerator) formatFileBasedOnExt(fileName, outputFile string) (err error) {
	ext := path.Ext(fileName)

	if formater, ok := tg.formaterByExt[ext]; ok {
		err = formater(outputFile)
	} else {
		err = ErrUnsupportedFileType
	}

	return err
}
