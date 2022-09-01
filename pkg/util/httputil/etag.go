package httputil

// We used to have a ETag middleware.
// But the implementation of ETag essentially cache the whole response in memory
// in order to calculate the hash.
// This will consume too much memory.
// With httputil.FilesystemCache and httputil.FileServer,
// we already have good enough measures to make serving static files fast.
