job "simple_example" {
  datacenters = ["dc1"]
  type        = "batch"

  group "simple_example" {
    task "reverse-string" {
      driver = "wasmtime"

      config {
        modulePath = "/home/dpovolotskii/git/nomad-wasmtime-driver-plugin/example/wasm-modules/buf_invertor.wasm"
        ioBuffer {
          enabled = true
          inputValue = "test_string"
        }
      }
    }
  }
}
