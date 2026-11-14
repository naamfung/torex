#!/bin/sh
echo "Building winkit.exe with Shell..."

# Check if we have the required tools
if ! command -v gcc &> /dev/null; then
    echo "Error: GCC not found in PATH"
    echo "Please install MinGW-w64 or add it to your PATH"
    exit 1
fi

if ! command -v windres &> /dev/null; then
    echo "Error: windres not found in PATH"
    echo "Please install MinGW-w64 or add it to your PATH"
    exit 1
fi

if ! command -v ar &> /dev/null; then
    echo "Error: ar not found in PATH"
    echo "Please install MinGW-w64 or add it to your PATH"
    exit 1
fi

# Build libmemory.a
echo "Building libmemory.a..."
gcc -c -DSET -Wa,--no-pad-sections -Os -fno-builtin -fno-asynchronous-unwind-tables -fno-ident -o set.o libmemory.c
gcc -c -DCOPY -Wa,--no-pad-sections -Os -fno-builtin -fno-asynchronous-unwind-tables -fno-ident -o copy.o libmemory.c
gcc -c -DFILL -Wa,--no-pad-sections -Os -fno-builtin -fno-asynchronous-unwind-tables -fno-ident -o fill.o libmemory.c
gcc -c -DCMP -Wa,--no-pad-sections -Os -fno-builtin -fno-asynchronous-unwind-tables -fno-ident -o cmp.o libmemory.c
gcc -c -DLEN -Wa,--no-pad-sections -Os -fno-builtin -fno-asynchronous-unwind-tables -fno-ident -o len.o libmemory.c
ar r libmemory.a set.o copy.o fill.o cmp.o len.o
rm -f set.o copy.o fill.o cmp.o len.o

# Build libchkstk.a
echo "Building libchkstk.a..."
gcc -c -DCHKSTK_MS -Wa,--no-pad-sections -o chkstk_ms.o libchkstk.S
gcc -c -DCHKSTK -Wa,--no-pad-sections -o chkstk.o libchkstk.S
ar r libchkstk.a chkstk_ms.o chkstk.o
rm -f chkstk_ms.o chkstk.o

# Create resource file
echo "Creating resource file..."
echo 'id ICON "logo.ico"' > logo.rc

# Compile resources
echo "Compiling resources..."
windres -O coff -F pe-x86-64 logo.rc -o logo.o

# Compile winkit.exe
echo "Compiling winkit.exe..."
gcc -DVERSION="2.4.0" -nostdlib -fno-asynchronous-unwind-tables -fno-builtin -Wl,--gc-sections -s -o winkit.exe winkit.c logo.o -L. -lkernel32 -luser32 -lshell32 -lmemory -ladvapi32 -ldbghelp

if [ $? -ne 0 ]; then
    echo "Error: Failed to compile winkit.exe"
    exit 1
fi

echo "Build successful! winkit.exe created."