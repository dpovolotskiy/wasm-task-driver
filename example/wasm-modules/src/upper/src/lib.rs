use json;
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

    let mystring = String::from_utf8(io_buffer.to_vec()).unwrap();

    let line = match json::parse(&mystring) {
        Ok(j) => j["line"].to_string(),
        Err(e) => e.to_string(),
    };

    let mut data = json::JsonValue::new_object();
    data["output"] = line.to_uppercase().into();

    let newstring = data.dump();
    let newbytes = newstring.as_bytes();

    if newbytes.len() > IO_BUFFER_SIZE {
        return 0;
    }

    let mut i = 0;
    while i < newbytes.len() {
        io_buffer[i] = newbytes[i];
        i = i+1;
    }

    return newbytes.len()
}
