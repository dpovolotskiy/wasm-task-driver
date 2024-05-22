use json;

const IO_BUFFER_SIZE: usize = 4096;

static mut IO_BUFFER: [u8; IO_BUFFER_SIZE] = [0; IO_BUFFER_SIZE];

#[export_name = "get_io_buffer_ptr"]
pub extern "C" fn get_io_buffer_ptr() -> *const u8 {
    let ptr: *const u8;
    unsafe {
        ptr = IO_BUFFER.as_ptr()
    }

    return ptr
}

#[export_name = "handle_buffer"]
pub extern "C" fn handle_buffer(length: usize) -> usize {
    if length > IO_BUFFER_SIZE {
        return 0;
    }

    let mystring : String;

    unsafe {
        mystring = String::from_utf8(IO_BUFFER[0..length].to_vec()).unwrap();
    }

    let line = match json::parse(&mystring) {
        Ok(j) => j["line"].to_string(),
        Err(e) => e.to_string(),
    };

    let mut data = json::JsonValue::new_object();
    data["output"] = line.to_uppercase().into();

    let newstring = data.dump();
    let newbytes = newstring.as_bytes();
    unsafe {
        let mut i = 0;
        while i < newbytes.len() {
            IO_BUFFER[i] = newbytes[i];
            i = i+1;
        }
    }

    return if newbytes.len() > IO_BUFFER_SIZE { 0 } else { newbytes.len() }
}
