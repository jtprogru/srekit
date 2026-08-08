## MODIFIED Requirements

### Requirement: Recognized configuration keys

The configuration file SHALL recognize at least `author`, `full_name`, `email`, `templates_dir`, and `changelog_lang`. `full_name` SHALL act as a fallback for `author`. `changelog_lang` SHALL accept the same values as the `--lang` flag and SHALL be overridden by it.

#### Scenario: full_name serves as the author
- **GIVEN** a config file setting `full_name` but not `author`
- **WHEN** a generator resolves the author
- **THEN** the `full_name` value SHALL be used

#### Scenario: Changelog language comes from configuration
- **GIVEN** a config file setting `changelog_lang: ru`
- **WHEN** a user generates a changelog without `--lang`
- **THEN** the Russian variant SHALL be rendered

#### Scenario: An unrecognized language in configuration fails loudly
- **GIVEN** a config file setting `changelog_lang: de`
- **WHEN** a user generates a changelog without `--lang`
- **THEN** the command SHALL exit non-zero with an error naming the accepted values rather than silently falling back to English
