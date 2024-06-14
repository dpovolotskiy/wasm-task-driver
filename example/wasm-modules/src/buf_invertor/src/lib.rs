use std::mem;
use std::os::raw::c_void;

static mut IO_BUFFER_SIZE: usize = 0;

#[export_name = "alloc"]
pub extern "C" fn alloc(length: usize) -> *mut c_void {
    let mut io_buffer = Vec::with_capacity(length);
    let ptr = io_buffer.as_mut_ptr();

    unsafe {
        IO_BUFFER_SIZE = length
    }

    mem::forget(io_buffer);

    ptr
}

#[export_name = "handle_buffer"]
pub unsafe extern "C" fn handle_buffer(ptr: *mut u8, length: usize) -> usize {
    if length > IO_BUFFER_SIZE {
        return 0;
    }

    let io_buffer = std::slice::from_raw_parts_mut(ptr, length);

    let mut start: usize = 0;
    let mut end: usize = length - 1;
    while start < end {
        (io_buffer[start], io_buffer[end]) = (io_buffer[end], io_buffer[start]);

        start += 1;
        end -= 1;
    }

    return length
}
