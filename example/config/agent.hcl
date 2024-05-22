log_level = "TRACE"
bind_addr = "0.0.0.0"

plugin "wasmtime-driver" {
  config {
    cache {
      enabled = true
      type = "lru"
      size = 10
      expiration {
        enabled = true
        entryTTL = 10
      }
      preCache {
        enabled = false
        modulesDir = "/home/dpovolotskii/git/nomad-wasmtime-driver-plugin/example/wasm-modules"
      }
    }
  }
}
