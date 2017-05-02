---
License: MIT
LicenseFile: LICENSE
LicenseColor: yellow
---
# {{.Name}}

{{template "badge/travis" .}} {{template "badge/appveyor" .}} {{template "badge/goreport" .}} {{template "badge/godoc" .}} {{template "license/shields" .}}

{{pkgdoc}}

Choose your gun! | [Aux armes!](https://www.youtube.com/watch?v=hD-wD_AMRYc&t=7)

# {{toc 5}}

# Install

#### glide
{{template "glide/install" .}}


# Usage

#### $ {{exec "httper" "-help" | color "sh"}}

## Cli examples

```sh
# Create a httped version of Tomate to MyTomate
httper http_vegetables_gen.go *JSONTomates:HTTPTomates
```

# API example

Following example demonstates a program using it to generate an `httped` version of a type.

#### > {{cat "demo/main.go" | color "go"}}

Following code is the generated implementation of an `httped` typed slice of `Tomate`.

#### > {{cat "demo/http_vegetables_gen.go" | color "go"}}

# Recipes

#### Release the project

```sh
gump patch -d # check
gump patch # bump
```

# History

[CHANGELOG](CHANGELOG.md)
