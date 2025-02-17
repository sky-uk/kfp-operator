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
To document different versions of the KFP Operator, we store an archive of the `content/docs/` folder for each release version, under `content/archive/<version>/`.

To create a new archive version:
1. run `make archive-version VERSION=<new_version>` with the new version i.e. `v0.6.0`
8. Rebuild the site with `make build`.
