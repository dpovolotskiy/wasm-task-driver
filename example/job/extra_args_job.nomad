job "extra_args_example" {
  datacenters = ["dc1"]
  type        = "batch"

  group "extra_args_example" {
    task "sum-numbers" {
      driver = "wasm-task-driver"

      config {
        modulePath = "/home/dpovolotskii/git/wasm-task-driver/example/wasm-modules/sum.wasm"
        main {
          mainFuncName = "sum"
          args = [123456, 678910]
        }
      }
    }
  }
}
