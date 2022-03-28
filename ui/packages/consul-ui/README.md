# consul-ui

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Yarn Commands](#yarn-commands)
- [Running / Development](#running--development)
  - [Environment Variables](#environment-variables)
  - [Contributing/Engineering Documentation](#contributingengineering-documentation)
  - [Browser 'Debug Utility' Functions and 'Environment' Variables](#browser-debug-utility-functions-and-environment-variables)
  - [Code Generators](#code-generators)
  - [Running Tests](#running-tests)
  - [Linting](#linting)
  - [Building](#building)
    - [Running Tests in Parallel](#running-tests-in-parallel)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Prerequisites

You will need the following software installed:

* [Git](https://git-scm.com/)
* [Node.js](https://nodejs.org/) v14, with npm
    * Optionally, install and manage via [Node Version Manager (nvm)](https://github.com/nvm-sh/nvm)
* [yarn](https://yarnpkg.com/getting-started/install)
* [Ember CLI](https://ember-cli.com/)
* [Google Chrome](https://google.com/chrome/)

## Setup & First Time Run

* `git clone https://github.com/hashicorp/consul.git` this repository
* `cd ui/packages/consul-ui`
* `make start` (first time runs)
    * This ensures dependencies are installed fresh and then calls `ember serve` with the necessary flags to run locally
    * `yarn start` for future runs
* Visit [http://localhost:4200/ui/](http://localhost:4200/ui/) in your browser

### Running The Test Suite

* `cd ui/packages/consul-ui`
* `make test-oss-view` which will run the tests in Chrome

(see "Running Tests" under "Development" below and/or the [testing section of the engineering docs](./docs/testing.mdx) for
further detail)

## Available Yarn Commands

Most used tooling scripts below primarily use `make` which will `yarn install`
and in turn call node package scripts.

List of available project commands.  `yarn run <command-name>`

| Command | Description |
| ------- | ----------- |
| build:staging | Builds the UI in staging mode (ready for PR preview site). |
| build:ci | Builds the UI for CI. |
| build | Builds the UI for production. |
| lint | Runs all lint commands. |
| lint:hbs | Lints `hbs` template files. |
| lint:js | Lints `js` files. |
| format | Runs all auto-formatters. |
| format:js | Auto-formats `js` files using Prettier. |
| format:sass | Auto-formats `scss` files using Prettier. |
| start | Runs the development app on a local server using the mock API. |
| start:consul | Runs the development app local server using a real consul instance as the backend. |
| start:staging | Runs the staging app local server. |
| test | Runs the ember tests in a headless browser. |
| test:view | Runs the ember tests in a non-headless browser. |
| test:oss | Runs only the OSS ember tests in a headless browser. |
| test:oss:view | Runs only the OSS ember tests in a non-headless browser. |
| test:coverage:view | Runs only the test specified for coverage in a non-headless browser. |
| test:node | Runs tests that can't be run in ember using node. |
| doc:toc | Automatically generates a table of contents for this README file. |

## Development

### Running Locally w/ Mock API

The source code includes a small development mode that provides a minimal set of Consul API
mocks that allow you to run the UI without having to run Consul. The mock API is run automatically.

* `make start` or `yarn start` to start the ember app
* Visit your app at [http://localhost:4200/ui](http://localhost:4200/ui).

### Running Against The Consul Backend

* `consul server -dev` to start consul listening on http://localhost:8500
* `cd ui/packages/consul-ui && make start-consul` to start the ember app proxying to `consul` (this will
respect the `CONSUL_HTTP_ADDR` environment variable to locate the Consul
installation.
* Visit your app at [http://localhost:4200/ui](http://localhost:4200/ui).

Example:

```bash
consul server -dev
CONSUL_HTTP_ADDR=http://10.0.0.1:8500 make start-consul
```

### Environment Variables

See [./docs/index.mdx](./docs/index.mdx#environment-variables)

### Branching

Follow a `ui/**/**` branch naming pattern. This branch naming pattern allows front-end focused builds, such as FE tests, to run automatically in Pull Requests. It also adds the `theme/ui` label to Pull Requests.

Examples:
- `ui/feature/add...`
- `ui/bugfix/fix...`
- `ui/enhancement/update...`

### Contributing/Engineering Documentation

We have an in-app (only during development) component storybook and
documentation site which can be visited using the [Eng
Docs](http://localhost:4200/ui/docs) link in the top navigation of the UI.
Alternatively all of these docs are also readable via GitHub's UI, so folks can
use whatever works best for them.

### Browser 'Debug Utility' Functions and 'Environment' Variables

Run `make start` then visit http://localhost:4200/ui/docs/bookmarklets for a
list of debug/engineering utilities you can use to help development of the UI
under certain scenarios.

### Code Generators

Many classes used in the UI can be generated with ember generators, try `ember help generate` for more details

### Running Tests

Tests use the mock api (see ./mock-api for details), the mock-api runs
automatically during testing, you don't need to run anything separately from
the below commands in order for the tests to use the mock-api.

* `make test` or `yarn run test`
* `make test-view` or `yarn run test:view` to view the tests running in Chrome

For more guidance on running tests, see the [testing section of the engineering docs](./docs/testing.mdx).

OSS only tests can also be run using:

* `make test-oss` or `yarn run test:oss`
* `make test-oss-view` or `yarn run test:oss:view` to view the tests running in Chrome

### Linting

`make lint` currently runs linting on the majority of js files and hbs files (using `ember-template-lint`).

See `.eslintrc.js` and `.eslintignore` for specific configuration.

### Building

* `make build` builds the UI for production usage (env=production)
* `make build-ci` builds the UI for CI/test usage (env=test)

Static files are built into ./dist

#### Running Tests in Parallel

You probably don't need to understand this if you are simply running some
tests locally.

Alternatively, `ember-exam` can be used to split the tests across multiple browser instances for faster results. Most options are the same as `ember test`. To see a full list of options, run `ember exam --help`.

**Note:** The `EMBER_EXAM_PARALLEL` environment variable must be set to override the default `parallel` value of `1` browser instance in [testem.js](./testem.js).

To quickly run the tests across 4 parallel browser instances:
```sh
make test-parallel
```

To run manually:
```sh
$ EMBER_EXAM_PARALLEL=true ./node_modules/.bin/ember exam --split <num> --parallel
```

More ways to split tests can be found in the [ember-exam README.md](https://github.com/trentmwillis/ember-exam/blob/master/README.md).
