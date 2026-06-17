"""
MoveKit Frida Tracer - 拦截所有 Win32 API 调用
用法：先启动 MoveKit.exe，然后运行 python frida_trace.py
"""
import frida
import sys
import json
import time
from datetime import datetime

# Hook 这些关键 Win32 API
HOOK_APIS = """
// === Registry ===
var advapi32 = Module.findBaseAddress('advapi32.dll');
var kernel32 = Module.findBaseAddress('kernel32.dll');
var shell32 = Module.findBaseAddress('shell32.dll');

// RegOpenKeyExW
Interceptor.attach(Module.findExportByName('advapi32.dll', 'RegOpenKeyExW'), {
    onEnter: function(args) {
        this.key = args[1].readUtf16String();
        this.subkey = args[2].readUtf16String();
    },
    onLeave: function(retval) {
        send({type: 'REG', api: 'RegOpenKeyExW', key: this.key, subkey: this.subkey, result: retval.toInt32()});
    }
});

// RegSetValueExW
Interceptor.attach(Module.findExportByName('advapi32.dll', 'RegSetValueExW'), {
    onEnter: function(args) {
        this.valueName = args[1].readUtf16String();
        this.dataType = args[4].toInt32();
        var data = args[5];
        var size = args[6].toInt32();
        var value = '';
        if (this.dataType === 1 || this.dataType === 2) { // REG_SZ / REG_EXPAND_SZ
            value = data.readUtf16String();
        } else if (this.dataType === 4) { // REG_DWORD
            value = data.readInt().toString();
        }
        send({type: 'REG', api: 'RegSetValueExW', name: this.valueName, dataType: this.dataType, value: value});
    }
});

// RegQueryValueExW
Interceptor.attach(Module.findExportByName('advapi32.dll', 'RegQueryValueExW'), {
    onEnter: function(args) {
        this.valueName = args[1].readUtf16String();
    },
    onLeave: function(retval) {
        send({type: 'REG', api: 'RegQueryValueExW', name: this.valueName, result: retval.toInt32()});
    }
});

// === File I/O ===
// CreateFileW
Interceptor.attach(Module.findExportByName('kernel32.dll', 'CreateFileW'), {
    onEnter: function(args) {
        this.path = args[0].readUtf16String();
        this.access = args[1].toInt32();
        this.share = args[3].toInt32();
        this.disp = args[4].toInt32();
        this.flags = args[5].toInt32();
    },
    onLeave: function(retval) {
        send({type: 'FILE', api: 'CreateFileW', path: this.path, access: this.access, flags: this.flags, handle: retval.toInt32()});
    }
});

// CopyFileExW
Interceptor.attach(Module.findExportByName('kernel32.dll', 'CopyFileExW'), {
    onEnter: function(args) {
        this.src = args[0].readUtf16String();
        this.dst = args[1].readUtf16String();
    },
    onLeave: function(retval) {
        send({type: 'FILE', api: 'CopyFileExW', src: this.src, dst: this.dst, success: retval.toInt32() !== 0});
    }
});

// MoveFileExW
Interceptor.attach(Module.findExportByName('kernel32.dll', 'MoveFileExW'), {
    onEnter: function(args) {
        this.src = args[0].readUtf16String();
        this.dst = args[1].readUtf16String();
        this.flags = args[2].toInt32();
    },
    onLeave: function(retval) {
        send({type: 'FILE', api: 'MoveFileExW', src: this.src, dst: this.dst, flags: this.flags, success: retval.toInt32() !== 0});
    }
});

// DeleteFileW
Interceptor.attach(Module.findExportByName('kernel32.dll', 'DeleteFileW'), {
    onEnter: function(args) {
        this.path = args[0].readUtf16String();
    },
    onLeave: function(retval) {
        send({type: 'FILE', api: 'DeleteFileW', path: this.path, success: retval.toInt32() !== 0});
    }
});

// CreateDirectoryW
Interceptor.attach(Module.findExportByName('kernel32.dll', 'CreateDirectoryW'), {
    onEnter: function(args) {
        this.path = args[0].readUtf16String();
    },
    onLeave: function(retval) {
        send({type: 'FILE', api: 'CreateDirectoryW', path: this.path, success: retval.toInt32() !== 0});
    }
});

// === Volume ===
// GetVolumeInformationW
Interceptor.attach(Module.findExportByName('kernel32.dll', 'GetVolumeInformationW'), {
    onEnter: function(args) {
        this.root = args[0].readUtf16String();
    },
    onLeave: function(retval) {
        send({type: 'VOLUME', api: 'GetVolumeInformationW', root: this.root, success: retval.toInt32() !== 0});
    }
});

// GetDriveTypeW
Interceptor.attach(Module.findExportByName('kernel32.dll', 'GetDriveTypeW'), {
    onEnter: function(args) {
        this.root = args[0].readUtf16String();
    },
    onLeave: function(retval) {
        send({type: 'VOLUME', api: 'GetDriveTypeW', root: this.root, driveType: retval.toInt32()});
    }
});

// GetDiskFreeSpaceExW
Interceptor.attach(Module.findExportByName('kernel32.dll', 'GetDiskFreeSpaceExW'), {
    onEnter: function(args) {
        this.root = args[0].readUtf16String();
    },
    onLeave: function(retval) {
        send({type: 'VOLUME', api: 'GetDiskFreeSpaceExW', root: this.root, success: retval.toInt32() !== 0});
    }
});

// === Shell ===
// SHChangeNotify
Interceptor.attach(Module.findExportByName('shell32.dll', 'SHChangeNotify'), {
    onEnter: function(args) {
        send({type: 'SHELL', api: 'SHChangeNotify', event: args[0].toInt32(), flags: args[1].toInt32()});
    }
});

// === Process ===
// DeviceIoControl (for junction operations)
Interceptor.attach(Module.findExportByName('kernel32.dll', 'DeviceIoControl'), {
    onEnter: function(args) {
        this.ctrlCode = args[1].toInt32();
        if (this.ctrlCode === 0x000900A4 || this.ctrlCode === 0x000900A8 || this.ctrlCode === 0x000900AC) {
            // FSCTL_SET/GET/DELETE_REPARSE_POINT
            send({type: 'JUNCTION', api: 'DeviceIoControl', ctrlCode: '0x' + this.ctrlCode.toString(16)});
        }
    }
});

// GetFileAttributesW
Interceptor.attach(Module.findExportByName('kernel32.dll', 'GetFileAttributesW'), {
    onEnter: function(args) {
        this.path = args[0].readUtf16String();
    },
    onLeave: function(retval) {
        send({type: 'FILE', api: 'GetFileAttributesW', path: this.path, attrs: retval.toInt32()});
    }
});

send({type: 'INFO', msg: 'Frida hooks installed for MoveKit'});
"""

def on_message(message, data):
    if message['type'] == 'send':
        payload = message['payload']
        payload['ts'] = datetime.now().strftime('%H:%M:%S.%f')[:-3]
        print(json.dumps(payload, ensure_ascii=False))
    elif message['type'] == 'error':
        print(json.dumps({'type': 'ERROR', 'msg': message['description']}))

def main():
    # Find MoveKit process
    pid = None
    try:
        device = frida.get_local_device()
        for proc in device.enumerate_processes():
            if 'MoveKit' in proc.name:
                pid = proc.pid
                print(f"Found MoveKit.exe PID={pid}", file=sys.stderr)
                break
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)

    if pid is None:
        print("MoveKit.exe not found. Start it first, then run this script.", file=sys.stderr)
        sys.exit(1)

    # Attach and inject hooks
    session = frida.attach(pid)
    script = session.create_script(HOOK_APIS)
    script.on('message', on_message)
    script.load()

    print(f"Tracing MoveKit.exe (PID={pid}). Press Ctrl+C to stop.", file=sys.stderr)
    print(f"Output: JSON lines to stdout", file=sys.stderr)

    try:
        while True:
            time.sleep(0.1)
    except KeyboardInterrupt:
        print("\nStopping trace...", file=sys.stderr)
        session.detach()

if __name__ == '__main__':
    main()
