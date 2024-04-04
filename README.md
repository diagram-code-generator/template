<div align="center">

# template

[![GitHub tag](https://img.shields.io/github/release/diagram-code-generator/template?include_prereleases=&sort=semver&color=2ea44f&style=for-the-badge)](https://github.com/diagram-code-generator/template/releases/)
[![Go Report Card](https://goreportcard.com/badge/github.com/diagram-code-generator/template?style=for-the-badge)](https://goreportcard.com/report/github.com/diagram-code-generator/template)
[![Code coverage](https://img.shields.io/badge/Coverage-87.5%25-2ea44f?style=for-the-badge)](#)

[![Made with Golang](https://img.shields.io/badge/Golang-1.21.6-blue?logo=go&logoColor=white&style=for-the-badge)](https://go.dev "Go to Golang homepage")
[![Using Diagrams](https://img.shields.io/badge/diagrams.net-orange?logo=&logoColor=white&style=for-the-badge)](https://app.diagrams.net/ "Go to Diagrams homepage")

[![BuyMeACoffee](https://img.shields.io/badge/Buy%20Me%20a%20Coffee-ffdd00?style=for-the-badge&logo=buy-me-a-coffee&logoColor=black)](https://www.buymeacoffee.com/joselitofilho)

</div>

# Overview

The `TemplateGenerator` provides functionality to generate files using Go's [text/template][text/template] and apply 
formatting based on file extensions. It encapsulates the generation process and allows customization through options.

## Features

- Parse text templates.
- Execute templates with provided data.
- Generate files based on templates.
- Apply formatting based on file extensions. Supported formaters:
	- [x] .go -> Go
	- [x] .tf -> Terraform

## How to Use

```bash
$ go get github.com/diagram-code-generator/template@latest
```

## Example Usage

```Go
ackage main

import (
	"github.com/diagram-code-generator/template/pkg/generators"
)

func main() {
	// Create a new TemplateGenerator instance
	tg := generators.NewTemplateGenerator()

	// Define data to be used in the template
	data := struct {
		Name string
	}{
		Name: "Joselito",
	}

	// Define a text template
	templateContent := "Hello, {{.Name}}!"

	// Build and execute the template
	output, err := tg.Build(data, "example_template", templateContent)
	if err != nil {
		panic(err)
	}

	// Print the output
	fmt.Println(output)
}

```

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, feel free to create an 
[issue][issues] or submit a pull request. Your contribution is much appreciated. See [Contributing](CONTRIBUTING.md).

[![open - Contributing](https://img.shields.io/badge/open-contributing-blue?style=for-the-badge)](CONTRIBUTING.md "Go to contributing")

## License

This project is licensed under the [MIT License](LICENSE).

[issues]: https://github.com/diagram-code-generator/template/issues
[text/template]: https://pkg.go.dev/text/template