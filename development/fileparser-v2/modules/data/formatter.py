import struct

def write_to_buffer(buf, data, little_endian=True):
    """Helper function to write data into bytes with specified byte order."""
    buf.write(data if little_endian else data[::-1])  # Reverse bytes for big-endian

def format_bytes(data):
    """Format byte data into a string representation."""
    return ' '.join(f'{b}' for b in data) + ' '  # Convert bytes to space-separated string

# 1. Signed Int
def parse_signed_int(buf, value, size):
    """Parse signed integers of various sizes and write to the buffer."""
    if value == "":
        return parse_dummy_data(buf, size)

    value = int(value)

    if size == 1:
        data = struct.pack('<b', value)  # int8
    elif size == 2:
        data = struct.pack('<h', value)  # int16
    elif size == 4:
        data = struct.pack('<i', value)  # int32
    elif size == 8:
        data = struct.pack('<q', value)  # int64
    else:
        raise ValueError("unsupported signed integer size")

    write_to_buffer(buf, data, True)
    return format_bytes(data)


# 2. Unsigned Int
def parse_unsigned_int(buf, value, size):
    """Parse unsigned integers of various sizes and write to the buffer."""
    if value == "":
        return parse_dummy_data(buf, size)
    

    if type(value) != "int":
        value = int(value)

    if size == 1:
        data = struct.pack('<B', value)  # int8
    elif size == 2:
        data = struct.pack('<H', value)  # int16
    elif size == 4:
        data = struct.pack('<I', value)  # int32
    elif size == 8:
        data = struct.pack('<Q', value)  # int64
    else:
        raise ValueError("unsupported unsigned integer size")
    
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 3. Float
def parse_float(buf, value, decimals):
    """Parse float values as scaled integers and write to the buffer."""
    scaled_value = int(float(value) * (10 ** decimals))  # Scale based on decimals
    data = struct.pack('<q', scaled_value)  # int64
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 4. FixedString
def parse_fixed_string(buf, value, size):
    """Parse fixed-length strings with null padding."""
    if type(value) != "str":
        value = str(value)

    if value == "":
        return parse_dummy_data(buf, size)
    if len(value) > size:
        raise ValueError("string length exceeds fixed size")
    data = value.encode('utf-8') + b'\x00' * (size - len(value))  # Null padding
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 5. FixedString(0-terminable)
def parse_fixed_string_terminable(buf, value, size):
    """Parse fixed-length strings that are optionally 0-terminated."""
    if value is None:
        return parse_dummy_data(buf, size)
    if len(value) > size:
        raise ValueError("string length exceeds fixed size")
    data = value.encode('utf-8') + b'\x00' + b'\x00' * (size - len(value) - 1)  # Null terminate
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 6. VariableString
def parse_variable_string(buf, value, max_size):
    """Parse variable-length strings with length information."""
    if value is None:
        return parse_dummy_data(buf, max_size)
    str_len = len(value)
    if str_len > max_size:
        raise ValueError("string length exceeds maximum size")
    data = struct.pack('<B', str_len)  # Write length
    write_to_buffer(buf, data, True)
    buf.write(value.encode('utf-8'))
    buf.write(b'\x00' * (max_size - str_len))  # Padding to multiple of 8 bytes
    return format_bytes(data)

# 7. Counter
def parse_counter(buf, value, size):
    """Parse counters, represented as unsigned integers."""
    return parse_unsigned_int(buf, value, size)

# 8. LocalMktDate
def parse_local_mkt_date(buf, date):
    """Parse LocalMktDate in YYYYMMDD format."""
    data = struct.pack('<I', date)  # uint32
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 9. PriceType
def parse_price_type(buf, value):
    """Parse price as an 8-byte signed integer."""
    if value == 0:
        return parse_dummy_data(buf, 8)
    data = struct.pack('<q', int(value))  # int64
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 10. Amount
def parse_amount(buf, value):
    """Parse amount as an 8-byte signed integer with 2 decimal places."""
    value = float(value or 0.0)  # Default to zero if null
    scaled_value = int(value * 100)  # Scale for 2 decimals
    data = struct.pack('<q', scaled_value)  # int64
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 11. Qty
def parse_qty(buf, value):
    """Parse quantity as a signed 4-byte integer."""
    if value == 0:
        return parse_dummy_data(buf, 4)
    data = struct.pack('<i', int(value or 0))  # Default to zero if null
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 12. SeqNum
def parse_seq_num(buf, value):
    """Parse sequence number in 8-byte unsigned binary format."""
    data = struct.pack('<Q', int(value or 0))  # Default to zero if null
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 13. UTCTimestamp
def parse_utc_timestamp(buf, timestamp):
    """Parse UTC timestamp in nanoseconds since UNIX epoch (8-byte unsigned)."""
    data = struct.pack('<Q', int(timestamp or 0))  # Default to zero if null
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 14. char
def parse_char(buf, value):
    """Parse a string of length 1."""
    if not value or len(value) != 1:
        return parse_dummy_data(buf, 1)
    data = value.encode('utf-8')  # Write single character
    write_to_buffer(buf, data, True)
    return format_bytes(data)

# 15. Data
def parse_data(buf, data):
    """Parse byte array where each byte has a value from 0 to 255."""
    buf.write(data)
    return format_bytes(data)

# 16. chartext
def parse_char_text(buf, value, size):
    """Parse text fields with optional null termination."""
    return parse_fixed_string_terminable(buf, value, size)

def parse_dummy_data(buf, size, fill_char='X'):
    """Generate dummy data of specified size."""
    val = fill_char * size
    return parse_fixed_string(buf,val,size)

def write_to_redis(redis_client, key, value):
    """Write bytecode to Redis under the specified key."""
    redis_client.set(key, value)

