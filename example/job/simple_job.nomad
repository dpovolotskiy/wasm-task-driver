job "simple_example" {
  datacenters = ["dc1"]
  type        = "batch"

  group "simple_example" {
    task "reverse-string" {
      driver = "wasm-task-driver"

      config {
        engine = "wasmtime"
        modulePath = "/home/dpovolotskii/git/opensource/wasm-task-driver/example/wasm-modules/buf_invertor.wasm"
        ioBuffer {
          enabled = true
          inputValue = "test_string"
        }
      }
    }
  }
}
