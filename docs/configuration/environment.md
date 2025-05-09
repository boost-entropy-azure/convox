---
title: "Environment Variables"
draft: false
slug: Environment Variables
url: /configuration/environment
---
# Environment Variables

Convox encourages the use of environment variables for managing application secrets. Using environment
variables allows you to keep secrets out of your codebase and to have different configuration values
for different deployments (i.e. staging and production).

## Definition

Environment variables that can be used by the application are defined in [`convox.yml`](/configuration/convox-yml):

### Application Level

Environment variables defined at the top level affect every service in the application.
```html
    environment:
      - ENCRYPTION_KEY
    services:
      web:
        environment:
          - ALLOWED_IPS
          - COOKIE_SECRET
      worker:
        environment:
          - QUEUE
```
This application would require four environment variables to be set: `ALLOWED_IPS`, `COOKIE_SECRET`, `ENCRYPTION_KEY`, and `QUEUE`.

The `ENCRYPTION_KEY` variable will be available to both services.

### Service Level

Environment variables can be defined for each [Service](/reference/primitives/app/service).
```html
    services:
      web:
        environment:
          - ALLOWED_IPS
          - COOKIE_SECRET
      worker:
        environment:
          - QUEUE
```
This application would require three environment variables to be set: `ALLOWED_IPS`, `COOKIE_SECRET`, and `QUEUE`.

> Environment variables defined at the service level will only be available to that service. In the example above,
> the `web` service would not have a value set for the `QUEUE` environment variable as that variable is only defined
> on the `worker` service.


### Default Values

You can set a default value for any environment variable in the manifest:
```html
    environment:
      - QUEUE=main
```
### Interpolation

You can also use environment variables to add dynamic configuration to your `convox.yml`:
```html
    services:
      web:
        health: ${HEALTH_CHECK_PATH}
```
## Configuration

You can set values for your environment variables using `convox env set`:
```html
    $ convox env set ALLOWED_IPS=1.2.3.4 COOKIE_SECRET=foo QUEUE=main
    Setting ALLOWED_IPS, COOKIE_SECRET, QUEUE... OK
    Release: RABCDEFGHI
```
Setting environment variables will cause a new [Release](/reference/primitives/app/release) to be created. In order to deploy
your changes you will need to promote this release.
```html
    $ convox releases promote RABCDEFGHIJ
    Promoting RABCDEFGHIJ... OK
```

> Environment variables can be set using the CLI or through the Console and their values will be available to the
> [Service](/reference/primitives/app/service).

### Release-Based Environment Management

You can manage environment variables for a specific release by using the `--release` flag with any `convox env` command. This is useful in workflows where multiple builds may occur between promotions, allowing you to target environment changes to a specific release rather than always applying them to the latest release.

View environment variables for a specific release:
```html
    $ convox env -a myapp --release RXYZ123
    ALLOWED_IPS=1.2.3.4
    COOKIE_SECRET=foo
    QUEUE=main
```

Set environment variables for a specific release:
```html
    $ convox env set FEATURE_FLAG=true -a myapp --release RXYZ123
    Setting FEATURE_FLAG... OK
    Release: RXYZ123
```

Unset environment variables for a specific release:
```html
    $ convox env unset LEGACY_MODE -a myapp --release RXYZ123
    Unsetting LEGACY_MODE... OK
    Release: RXYZ123
```

Edit environment variables interactively for a specific release:
```html
    $ convox env edit -a myapp --release RXYZ123
    Setting ... OK
    Release: RXYZ123
```

This release-specific environment management ensures that environment changes are only applied to the intended release, preventing unintentional modifications to other releases in your workflow.

## System Variables

The following environment variables are automatically set by Convox.

| Name                | Description                                                                                   |
| ------------------- | --------------------------------------------------------------------------------------------- |
| **APP**               | Name of the [App](/reference/primitives/app)                                                |
| **BUILD**             | ID of the currently-promoted [Build](/reference/primitives/app/build)                    |
| **BUILD_DESCRIPTION** | Description of the currently-promoted [Build](/reference/primitives/app/build)           |
| **PORT**              | The value of the **port:** attribute for this [Service](/reference/primitives/app/service) |
| **RACK**              | The name of the [Rack](/reference/primitives/rack)                                       |
| **RELEASE**           | ID of the currently-promoted [Release](/reference/primitives/app/release)                |
| **SERVICE**           | Name of the [Service](/reference/primitives/app/service)                                 |
