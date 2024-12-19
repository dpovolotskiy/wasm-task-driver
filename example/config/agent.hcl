log_level = "TRACE"
bind_addr = "0.0.0.0"

plugin "wasm-task-driver" {
  config {
    engines = [
      {
        name = "wasmtime"
        enabled = true
        cache {
          enabled = true
          type = "lru"
          size = 10
          expiration {
            enabled = true
            entryTTL = 10
          }
          preCache {
            enabled = true
            modulesDir = "/home/dpovolotskii/git/opensource/wasm-task-driver/example/wasm-modules"
          }
        }
      }
    ]
  }
}
