# gosmopolitan

`gosmopolitan` checks your Go codebase for code smells that may prove to be
hindrance to internationalization ("i18n") and/or localization ("l10n").

The name is a wordplay on "cosmopolitan".

## Checks

Currently `gosmopolitan` checks for the following anti-patterns:

*   Occurrences of string literals containing characters from a certain
    writing system.

    Existence of such strings often means the relevant logic is hard to
    internationalize, or at least, require special care when doing i18n/l10n.

*   Usages of `time.Local`.

    An internationalized app or library should almost never process time and
    date values in the timezone in which it is running; instead one should use
    the respective user preference, or the timezone as dictated by the domain
    logic.

## golangci-lint integration

`gosmopolitan` is not integrated into [`golangci-lint`][gcl-home] yet, but
you can nevertheless run it [as a custom plugin][gcl-plugin].

[gcl-home]: https://golangci-lint.run
[gcl-plugin]: https://golangci-lint.run/contributing/new-linters/#how-to-add-a-private-linter-to-golangci-lint

First make yourself a plugin `.so` file like this:

```go
// compile this with something like `go build -buildmode=plugin`

package main

import (
	"github.com/xen0n/gosmopolitan"
	"golang.org/x/tools/go/analysis"
)

type analyzerPlugin struct{}

func (analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	// You can customize the options via the gosmopolitan.NewAnalyzer.
	return []*analysis.Analyzer{
		gosmopolitan.DefaultAnalyzer,
	}
}

var AnalyzerPlugin analyzerPlugin
```

You just need to make sure the `golang.org/x/tools` version used to build the
plugin is consistent with that of your `golangci-lint` binary. (Of course the
`golangci-lint` binary should be built with plugin support enabled too;
notably, [the Homebrew `golangci-lint` is built without plugin support][hb-issue],
so be ware of this.)

[hb-issue]: https://github.com/golangci/golangci-lint/issues/1182

Then reference it in your `.golangci.yml`, and enable it in the `linters`
section:

```yaml
linters:
  # ...
  enable:
    # ...
    - gosmopolitan
    # ...

linters-settings:
  custom:
    gosmopolitan:
      path: 'path/to/your/plugin.so'
      description: 'Report i18n/l10n anti-patterns in your Go codebase'
      original-url: 'https://github.com/xen0n/gosmopolitan'
  # ...
```

Then you can `golangci-lint run` and `//nolint:gosmopolitan` as you would
with any other supported linter.

## License

`gosmopolitan` is licensed under the GPL license, version 3 or later.
