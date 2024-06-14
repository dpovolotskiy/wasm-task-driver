job "full_example" {
  datacenters = ["dc1"]
  type        = "batch"

  group "full_example" {
    task "up-string" {
      driver = "wasmtime"

      config {
        modulePath = "/home/dpovolotskii/git/nomad-wasmtime-driver-plugin/example/wasm-modules/upper.wasm"
        ioBuffer {
          enabled = true
          size = 4096
          inputValue = "{ \"line\": \"test\" }"
          IOBufFuncName = "alloc"
          args = []
        }
        main {
          mainFuncName = "handle_buffer"
          args = []
        }
      }
    }
  }
}
