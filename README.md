# gosmopolitan

`gosmopolitan` checks your Go codebase for code smells that may prove to be
hindrance to internationalization ("i18n") and/or localization ("l10n").

The name is a wordplay on "cosmopolitan".

## Checks

Currently `gosmopolitan` checks for the following patterns:

*   Occurrences of string literals containing characters from a certain
    writing system.

    Existence of such strings often means the relevant logic is hard to
    internationalize, or at least, require special care when doing i18n/l10n.

## License

`gosmopolitan` is licensed under the GPL license, version 3 or later.
