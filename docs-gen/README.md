# KFP-Operator Website

this website is built using [Hugo](https://gohugo.io/) and uses the [docsy](https://www.docsy.dev/) theme

## Development

Run the website locally:

```bash
make serve
```

and access the website at http://localhost:1313/kfp-operator/

## Build

Build the website:

```bash
make
```

This will populate the [/docs](/docs) directory, which will be served on [GitHub Pages](https://sky-uk.github.io/kfp-operator).

## Versioning
To document different versions of the KFP Operator, we store a copy of the `content/docs/` folder for each release version, under `content/versions/<version>/`.

To create a new version:
1. Run `make archive-version VERSION=<new_version>` where `new_version` is the version, e.g. `v0.6.0`
2. Ensure the new version is working correctly by validating via `make serve`
3. Generate the minified version of the website by running `make build`
4. Track the new files, commit the changes and push to the repository
