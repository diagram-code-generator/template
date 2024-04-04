package generators

import (
	_ "embed"
	"os"
	"path"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	type args struct {
		data            any
		templateName    string
		templateContent string
	}

	tests := []struct {
		name           string
		args           args
		want           string
		expectedErrMsg string
	}{
		{
			name: "valid case",
			args: args{
				data:            map[string]any{"Name": "John", "Age": 30},
				templateName:    "example",
				templateContent: "Name: {{.Name}}, Age: {{.Age}}",
			},
			want: "Name: John, Age: 30",
		},
		{
			name: "ToCamel func",
			args: args{
				data:            map[string]any{"Name": "my-name"},
				templateName:    "example",
				templateContent: "Name: {{ToCamel .Name}}",
			},
			want: "Name: myName",
		},
		{
			name: "ToKebab func",
			args: args{
				data:            map[string]any{"Name": "myName"},
				templateName:    "example",
				templateContent: "Name: {{ToKebab .Name}}",
			},
			want: "Name: my-name",
		},
		{
			name: "ToLower func",
			args: args{
				data:            map[string]any{"Name": "MY-NAME"},
				templateName:    "example",
				templateContent: "Name: {{ToLower .Name}}",
			},
			want: "Name: my-name",
		},
		{
			name: "ToPascal func",
			args: args{
				data:            map[string]any{"Name": "my-name"},
				templateName:    "example",
				templateContent: "Name: {{ToPascal .Name}}",
			},
			want: "Name: MyName",
		},
		{
			name: "ToSpace func",
			args: args{
				data:            map[string]any{"Name": "my-name"},
				templateName:    "example",
				templateContent: "Name: {{ToSpace .Name}}",
			},
			want: "Name: my name",
		},
		{
			name: "ToSnake func",
			args: args{
				data:            map[string]any{"Name": "my-name"},
				templateName:    "example",
				templateContent: "Name: {{ToSnake .Name}}",
			},
			want: "Name: my_name",
		},
		{
			name: "ToUpper func",
			args: args{
				data:            map[string]any{"Name": "my-name"},
				templateName:    "example",
				templateContent: "Name: {{ToUpper .Name}}",
			},
			want: "Name: MY-NAME",
		},
		{
			name: "missing field",
			args: args{
				data:            map[string]any{"Name": "John", "Age": 30},
				templateName:    "invalid",
				templateContent: "{{ .MissingField }}",
			},
			want: "<no value>",
		},
		{
			name: "invalid template",
			args: args{
				data:            map[string]any{"Name": "John", "Age": 30},
				templateName:    "invalid",
				templateContent: "{{ InvalidFunction .Name }}",
			},
			want:           "",
			expectedErrMsg: "template: invalid:1: function \"InvalidFunction\" not defined",
		},
	}

	for i := range tests {
		tc := tests[i]

		t.Run(tc.name, func(t *testing.T) {
			got, err := NewTemplateGenerator().Build(tc.args.data, tc.args.templateName, tc.args.templateContent)

			if tc.expectedErrMsg == "" {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				require.Equal(t, tc.expectedErrMsg, err.Error())
			}
		})
	}
}

func TestGenerateFile(t *testing.T) {
	type args struct {
		templatesMap map[string]string
		fileName     string
		fileTmpl     string
		outputFile   string
		data         any
	}

	testOutput := "./testoutput"
	_ = os.MkdirAll(testOutput, os.ModePerm)

	tests := []struct {
		name             string
		args             args
		extraValidations func(testing.TB, string, error)
		targetErr        error
	}{
		{
			name: "successful go file generation and formatting",
			args: args{
				templatesMap: map[string]string{"test.go": "type  My{{.Name}}Struct    struct   {}"},
				fileName:     "test.go",
				outputFile:   path.Join(testOutput, "output.go"),
				data:         struct{ Name string }{Name: "World"},
			},
			extraValidations: func(tb testing.TB, outputFile string, err error) {
				if err != nil {
					return
				}

				data, err := os.ReadFile(outputFile)
				require.NoError(tb, err)
				require.Equal(tb, "type MyWorldStruct struct{}", string(data))
			},
		},
		{
			name: "successful tf file generation and formatting",
			args: args{
				templatesMap: map[string]string{"test.tf": `resource    "aws_s3_bucket"    "{{.Name}}_bucket"  {}`},
				fileName:     "test.tf",
				outputFile:   path.Join(testOutput, "output.tf"),
				data:         struct{ Name string }{Name: "world"},
			},
			extraValidations: func(tb testing.TB, outputFile string, err error) {
				if err != nil {
					return
				}

				data, err := os.ReadFile(outputFile)
				require.NoError(tb, err)
				require.Equal(tb, `resource "aws_s3_bucket" "world_bucket" {}`, string(data))
			},
		},
		{
			name: "should file generation fails when file ext is unsuported",
			args: args{
				templatesMap: map[string]string{"test.txt": "Hello, {{.Name}}!"},
				fileName:     "test.txt",
				outputFile:   path.Join(testOutput, "output.txt"),
				data:         struct{ Name string }{Name: "World"},
			},
			extraValidations: func(tb testing.TB, outputFile string, err error) {
				if err != nil {
					return
				}

				data, err := os.ReadFile(outputFile)
				require.NoError(tb, err)
				require.Equal(tb, "Hello, World!", string(data))
			},
			targetErr: ErrUnsupportedFileType,
		},
	}

	defer func() {
		_ = os.RemoveAll(testOutput)
	}()

	for i := range tests {
		tc := tests[i]

		t.Run(tc.name, func(t *testing.T) {
			err := NewTemplateGenerator().GenerateFile(
				tc.args.templatesMap, tc.args.fileName, tc.args.fileTmpl, tc.args.outputFile, tc.args.data)

			require.ErrorIs(t, err, tc.targetErr)

			if tc.extraValidations != nil {
				tc.extraValidations(t, tc.args.outputFile, err)
			}
		})
	}
}

func TestGenerateFiles(t *testing.T) {
	type args struct {
		defaultTemplatesMap map[string]string
		templatesMap        map[string]string
		data                any
		output              string
	}

	testOutput := "./testoutput"
	_ = os.MkdirAll(testOutput, os.ModePerm)

	tests := []struct {
		name      string
		args      args
		targetErr error
	}{
		{
			name: "generate single file",
			args: args{
				defaultTemplatesMap: map[string]string{
					"template.txt": "Hello, {{.Name}}!",
				},
				templatesMap: nil,
				data:         struct{ Name string }{"John"},
				output:       testOutput,
			},
			targetErr: nil,
		},
	}

	defer func() {
		_ = os.RemoveAll(testOutput)
	}()

	for i := range tests {
		tc := tests[i]

		t.Run(tc.name, func(t *testing.T) {
			err := NewTemplateGenerator().GenerateFiles(
				tc.args.defaultTemplatesMap, tc.args.templatesMap, tc.args.data, tc.args.output)

			require.ErrorIs(t, err, tc.targetErr)
		})
	}
}

func TestWithFuncs(t *testing.T) {
	extraFuncs := template.FuncMap{
		"CustomFunc1": func() string { return "CustomFunc1" },
		"CustomFunc2": func() string { return "CustomFunc2" },
	}

	tg := NewTemplateGenerator(WithExtraFuncs(extraFuncs))

	for funcName, expectedFunc := range extraFuncs {
		actualFunc, ok := tg.funcs[funcName]

		require.True(t, ok)
		require.Equal(t, expectedFunc.(func() string)(), actualFunc.(func() string)())
	}
}

func TestWithExtraFormaterByExt(t *testing.T) {
	extraFormatters := FormaterByExtMap{
		".py":   func(outputFile string) error { return nil },
		".java": func(outputFile string) error { return nil },
	}

	tg := NewTemplateGenerator(WithExtraFormaterByExt(extraFormatters))

	for ext, expectedFormatter := range extraFormatters {
		actualFormatter, ok := tg.formaterByExt[ext]

		require.True(t, ok)
		require.Equal(t, expectedFormatter(""), actualFormatter(""))
	}
}
