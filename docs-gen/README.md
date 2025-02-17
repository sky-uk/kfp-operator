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
1. copy the contents of `content/docs/` to `content/archive/<new_version>/`, replacing `<new_version>` with the version you want to document.
2. Add this new version to `hugo.toml`:
    ```toml
    [[params.versions]]
    version = "<new_version>"
    url = "/kfp-operator/archive/<new_version>"
    ```
3. Rebuild the site with `make build`.
