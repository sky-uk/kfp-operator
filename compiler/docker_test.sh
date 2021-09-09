docker build . -t compiler
TMP=$(mktemp -d)
docker run -v "$TMP:/compiler" compiler
diff -r compiler $TMP