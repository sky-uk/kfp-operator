# Plan: Fix KFP v2 Launcher GCS Directory Artifact Download Bug

## Problem Statement

The KFP v2 launcher's `DownloadBlob` function in `backend/src/v2/objectstore/object_store.go` fails when downloading directory-structured artifacts from GCS. This breaks any downstream component that consumes a directory artifact produced by an upstream component (e.g. Pusher consuming a Trainer's model output).

**Error**: `mkdir .../model: not a directory`

## Root Cause Analysis

The bug is in `DownloadBlob` → `downloadFile` in `objectstore/object_store.go` ([source](https://github.com/kubeflow/pipelines/blob/master/backend/src/v2/objectstore/object_store.go)).

### The Flow

1. TFX Trainer saves a model directly to GCS via `tf.saved_model.save(model, "gs://bucket/.../model")`
2. TensorFlow's GCS filesystem creates a **zero-byte placeholder blob** at the exact artifact URI path (e.g. `Trainer/.../model`) as a directory marker
3. The model files are written under that prefix: `model/saved_model.pb`, `model/variables/...`, `model/assets/`, etc.
4. When the Pusher runs, the launcher calls `DownloadBlob` with `blobDir = "Trainer/.../model"`

### The Bug in `DownloadBlob`

```go
func DownloadBlob(ctx context.Context, bucket *blob.Bucket, localDir, blobDir string) error {
    iter := bucket.List(&blob.ListOptions{Prefix: blobDir})
    for {
        obj, err := iter.Next(ctx)
        // ...
        if obj.IsDir {
            continue  // only catches "virtual" dirs from delimiter-based listing
        } else if isBlobKeyUnderPrefix(obj.Key, blobDir) {
            // Downloads EVERY blob, including the marker at blobDir itself
            localPath, _ := sanitizeDownloadPath(localDir, blobDir, obj.Key)
            downloadFile(ctx, bucket, obj.Key, localPath)
        }
    }
}
```

When `obj.Key == blobDir` (the marker blob), `sanitizeDownloadPath` computes `relativePath = "."`, so `localPath = localDir` itself. Then `downloadFile` creates `localDir` **as a regular file** (0 bytes). When the next blob (e.g. `model/saved_model.pb`) is processed, `downloadFile` calls `os.MkdirAll(filepath.Dir(localFilePath))` which needs `localDir` to be a directory — but it's already a file. **Crash.**

The same issue occurs for blobs ending with `/` (GCS directory markers like `model/assets/`), which get downloaded as files and block creation of subdirectories under them.

## Proposed Fix

### Change 1: Skip marker blobs in `DownloadBlob` (objectstore/object_store.go)

Add two skip conditions to the `DownloadBlob` loop:

```go
func DownloadBlob(ctx context.Context, bucket *blob.Bucket, localDir, blobDir string) error {
    iter := bucket.List(&blob.ListOptions{Prefix: blobDir})
    for {
        obj, err := iter.Next(ctx)
        if err != nil {
            if err == io.EOF {
                break
            }
            return fmt.Errorf("failed to list objects in remote storage %q: %w", blobDir, err)
        }
        if obj.IsDir {
            continue
        }

        // NEW: Skip zero-byte directory marker blobs.
        // GCS clients (including TensorFlow's gfile) create these as
        // placeholders when making "directories". They end with '/' or
        // match the prefix exactly. Downloading them as files prevents
        // subsequent MkdirAll calls from creating the actual directory.
        if obj.Size == 0 && (obj.Key == blobDir || strings.HasSuffix(obj.Key, "/")) {
            glog.V(4).Infof("DownloadBlob: skipping directory marker %q", obj.Key)
            continue
        }

        if isBlobKeyUnderPrefix(obj.Key, blobDir) {
            localPath, err := sanitizeDownloadPath(localDir, blobDir, obj.Key)
            if err != nil {
                return err
            }
            if err := downloadFile(ctx, bucket, obj.Key, localPath); err != nil {
                return err
            }
        } else {
            glog.V(4).Infof("DownloadBlob: skipping blob key %q not under expected prefix %q",
                obj.Key, blobDir)
        }
    }
    return nil
}
```

### Change 2: Add tests (objectstore/object_store_test.go)

Add test cases covering:

1. **Marker blob at prefix root**: A zero-byte blob at `blobDir` itself is skipped
2. **Marker blob with trailing slash**: A zero-byte blob like `model/assets/` is skipped
3. **Non-zero blobs are still downloaded**: Regular files even at the root prefix are downloaded
4. **Nested directory structures work**: `model/variables/variables.data-*` is downloaded correctly after markers are skipped

```go
func TestDownloadBlob_SkipsDirectoryMarkers(t *testing.T) {
    // Set up an in-memory bucket with:
    //   "model"              → 0 bytes (marker at prefix root)
    //   "model/saved_model.pb" → non-zero
    //   "model/assets/"      → 0 bytes (nested dir marker)
    //   "model/variables/variables.index" → non-zero
    //
    // Assert:
    //   - localDir/saved_model.pb exists as file
    //   - localDir/variables/variables.index exists as file
    //   - localDir is a directory, NOT a file
    //   - localDir/assets does NOT exist as a file
}
```

## Files to Modify

All changes are in the upstream KFP repository: [`kubeflow/pipelines`](https://github.com/kubeflow/pipelines)

| File | Change |
|------|--------|
| `backend/src/v2/objectstore/object_store.go` | Add marker skip logic to `DownloadBlob` (~5 lines) |
| `backend/src/v2/objectstore/object_store_test.go` | Add test cases for directory marker handling |

## Impact

- **Fixes**: All directory-structured artifact downloads (Model, Examples, etc.)
- **No regressions**: Only skips zero-byte blobs that match marker patterns; real files are unaffected
- **Eliminates need for**:
  - Fix 5 (patching `path_utils.serving_model_dir` to remove Format-Serving)
  - Fix 6 (GCS marker cleanup in trainer.py)
- **Scope**: Global fix — works for all components, all artifact types, all pipelines

## Delivery Path

1. **Fork** `kubeflow/pipelines` and create a feature branch
2. **Implement** Change 1 and Change 2
3. **Run** existing objectstore tests: `go test ./backend/src/v2/objectstore/...`
4. **Open PR** against `kubeflow/pipelines` with reproduction steps
5. **Interim**: Until merged and released, continue using Fix 5 + Fix 6 in the quickstart image as workarounds
6. **Post-merge**: Once a new KFP release includes the fix, remove Fix 5, Fix 6, and the `_cleanup_gcs_markers` function from `trainer.py`

## Confidence: 9/10

The fix is minimal, targeted, and safe:
- Zero-byte blobs ending with `/` are universally understood as directory markers in GCS
- Zero-byte blobs matching the listing prefix exactly are placeholder artifacts
- Neither of these should ever be materialised as local files
- The `obj.Size` field is populated by the `gocloud.dev/blob` library from GCS object metadata
- Existing `obj.IsDir` check already shows intent to skip directories; this extends it to handle real-world GCS markers
