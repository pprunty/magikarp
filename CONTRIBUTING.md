# Contributing to magikarp

We love your input! We want to make contributing to magikarp as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Run the test suite (`make test`)
6. Run the linter (`make lint`)
7. Commit your changes (`git commit -m 'Add some amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## Pull Request Process

1. Update the README.md with details of changes to the interface, if applicable
2. Update any documentation that might be affected by your changes
3. The PR will be merged once you have the sign-off of at least one maintainer

## Development Setup

1. Install dependencies:
```bash
make install
```

2. Build the project:
```bash
make build
```

3. Available commands:
- `make install` - Install dependencies
- `make build` - Build the binary
- `make run` - Build and run the binary
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make lint` - Run linter
- `make tools` - Install development tools

## Adding New Tools

1. Use the CLI to generate a tool skeleton:
```bash
mgk create-tool --plugin yourplugin --name yourtool --description "Your tool description"
```

2. Implement the tool logic in the generated files
3. Add tests for your tool
4. Register the tool in your plugin's `Initialize()` method

## Adding New Plugins

1. Use the CLI to generate a plugin skeleton:
```bash
mgk create-plugin --name yourplugin --description "Your plugin description"
```

2. Implement the Plugin interface in the generated files
3. Add your tools to the plugin
4. Register the plugin in `main.go`

## Bug Reports

We use GitHub issues to track public bugs. Report a bug by [opening a new issue](../../issues/new).

## License

By contributing, you agree that your contributions will be licensed under its MIT License.