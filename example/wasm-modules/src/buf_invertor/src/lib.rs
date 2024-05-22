const IO_BUFFER_SIZE: usize = 4096;

static mut IO_BUFFER: [u8; IO_BUFFER_SIZE] = [0; IO_BUFFER_SIZE];

#[export_name = "get_io_buffer_ptr"]
pub extern "C" fn get_io_buffer_ptr() -> *const u8 {
    let ptr: *const u8;
    unsafe {
        ptr = IO_BUFFER.as_ptr()
    }

    return ptr;
}

#[export_name = "handle_buffer"]
pub extern "C" fn handle_buffer(length: usize) -> usize {
    if length > IO_BUFFER_SIZE {
        return 0;
    }

    let mut start: usize = 0;
    let mut end: usize = length - 1;
    while start < end {
        unsafe {
            (IO_BUFFER[start], IO_BUFFER[end]) = (IO_BUFFER[end], IO_BUFFER[start]);
        }

        start += 1;
        end -= 1;
    }

    return length
}
