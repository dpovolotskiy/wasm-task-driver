#[export_name = "sum"]
pub extern "C" fn sum(num1: usize, num2: usize) -> usize {
    return num1 + num2
}
