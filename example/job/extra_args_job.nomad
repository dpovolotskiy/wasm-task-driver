job "extra_args_example" {
  datacenters = ["dc1"]
  type        = "batch"

  group "extra_args_example" {
    task "sum-numbers" {
      driver = "wasmtime"

      config {
        modulePath = "/home/dpovolotskii/git/nomad-wasmtime-driver-plugin/example/wasm-modules/sum.wasm"
        main {
          mainFuncName = "sum"
          args = [123456, 678910]
        }
      }
    }
  }
}
