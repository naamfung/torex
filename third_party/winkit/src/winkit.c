// Tiny, standalone launcher for winkit
// * Sets $WINKIT to the release version (-DVERSION)
// * Sets $WINKIT_HOME to the install location
// * Maybe sets $HOME according to winkit.ini
// * Maybe sets $PATH according to winkit.ini
// * Starts a login shell with "sh -l"
//
// $ gcc -DVERSION="$VERSION" -nostartfiles -o winkit.exe
//       winkit.c -lmemory
//
// This is free and unencumbered software released into the public domain.

#define countof(a)    (iz)(sizeof(a) / sizeof(*(a)))
#define assert(c)     while (!(c)) __builtin_unreachable()
#define new(a, t, n)  (t *)alloc(a, sizeof(t), _Alignof(t), n)

typedef unsigned char    u8;
typedef unsigned short   u16;
typedef   signed short   i16;
typedef   signed int     i32;
typedef   signed int     b32;
typedef unsigned int     u32;
typedef __PTRDIFF_TYPE__ iz;
typedef __SIZE_TYPE__    uz;
typedef unsigned short   char16_t;  // for GDB
typedef char16_t         c16;

// 添加控制台字体相关的结构体定义
typedef struct {
    u32 cbSize;
    u32 nFont;
    struct {
        i16 X;
        i16 Y;
    } dwFontSize;
    u32 FontFamily;
    u32 FontWeight;
    char16_t FaceName[32];
} CONSOLE_FONT_INFOEX;

typedef struct {
    u32 cb;
    uz  a, b, c;
    i32 d, e, f, g, h, i, j, k;
    u16 l, m;
    uz  n, o, p, q;
} Si;

typedef struct {
    uz  process;
    uz  thread;
    u32 pid;
    u32 tid;
} Pi;

// Windows API 常量定义
#define MAX_PATH 260

// 控制台屏幕缓冲区信息结构体
typedef struct {
    u32 dwSize;             // 控制台屏幕缓冲区的大小（字符行列数）
    u32 dwCursorPosition;   // 光标位置
    u32 wAttributes;        // 文本属性
    struct {
        i16 Left;
        i16 Top;
        i16 Right;
        i16 Bottom;
    } srWindow;             // 控制台窗口显示的区域
    u32 dwMaximumWindowSize;// 最大窗口大小
} CONSOLE_SCREEN_BUFFER_INFO;

// 窗口矩形结构体
typedef struct {
    i32 left;
    i32 top;
    i32 right;
    i32 bottom;
} RECT;

// 全局变量，用于标记当前参数是通过--cd=指定的目录
static b32 g_is_cd_param = 0;

enum {
    MAX_ENVVAR     = 32767,
    CP_UTF8        = 65001,
    PAGE_READWRITE = 0x04,
    MEM_COMMIT     = 0x1000,
    MEM_RESERVE    = 0x2000,
    MEM_RELEASE    = 0x8000,
    GENERIC_READ   = (i32)0x80000000,
    OPEN_EXISTING  = 3,
    FILE_SHARE_ALL = 7,
    
    // 注册表相关常量
    REG_SZ            = 1,
    REG_EXPAND_SZ     = 2,
    KEY_READ          = 0x20019,
    KEY_WRITE         = 0x20006,
    KEY_CREATE_SUB_KEY = 0x4,
    KEY_SET_VALUE     = 0x2,
    KEY_CREATE_LINK   = 0x20,
    KEY_WOW64_64KEY   = 0x100,
    KEY_WOW64_32KEY   = 0x200,
    
    // ShellExecute参数
    SW_SHOWNORMAL = 1,
    SW_HIDE       = 0,
    SM_CXSCREEN   = 0,
    SM_CYSCREEN   = 1,
    SWP_NOSIZE    = 0x0001,
    SWP_NOZORDER  = 0x0004,
};

#define W32 __attribute__((dllimport, stdcall))
W32 b32    CloseHandle(uz);
W32 uz     CreateFileW(c16 *, i32, i32, void *, i32, i32, uz);
W32 b32    CreateProcessW(c16*,c16*,void*,void*,i32,i32,c16*,c16*,Si*,Pi*);
W32 void   ExitProcess(i32) __attribute__((noreturn));
W32 u32    ExpandEnvironmentStringsW(c16 *, c16 *, u32);
W32 u32    GetEnvironmentVariableW(c16 *, c16 *, u32);
W32 i32    GetExitCodeProcess(uz, i32 *);
W32 u32    GetFullPathNameW(c16 *, u32, c16 *, c16 *);
W32 u32    GetModuleFileNameW(uz, c16 *, u32);
W32 u32    GetWindowsDirectoryW(c16 *, u32);
W32 i32    GetVersion();
W32 uz     GetStdHandle(u32);
W32 i32    MessageBoxW(uz, c16 *, c16 *, i32);
W32 i32    MultiByteToWideChar(i32, i32, u8 *, i32, c16 *, i32);
W32 b32    ReadFile(uz, u8 *, i32, i32 *, void *);
W32 b32    SetConsoleTitleW(c16 *);
W32 b32    SetCurrentDirectoryW(c16 *);
W32 b32    SetEnvironmentVariableW(c16 *, c16 *);
W32 u8    *VirtualAlloc(u8 *, iz, i32, i32);
W32 b32    VirtualFree(u8 *, uz, i32);
W32 i32    WaitForSingleObject(uz, i32);
W32 u32    SetConsoleOutputCP(u32);
W32 u32    SetConsoleCP(u32);
W32 b32    SetCurrentConsoleFontEx(uz, b32, CONSOLE_FONT_INFOEX *);
W32 c16   *GetCommandLineW(void);
W32 u32    GetFileAttributesW(c16 *);
W32 b32    WriteFile(uz, u8 *, i32, i32 *, void *);  // 添加写文件函数声明
W32 uz     GetConsoleWindow(void);
W32 b32    SetWindowPos(uz, uz, i32, i32, i32, i32, u32);
W32 i32    GetSystemMetrics(i32);
W32 b32    GetConsoleScreenBufferInfo(uz, void *);
W32 b32    GetWindowRect(uz, void *);

// 注册表相关API函数声明
W32 i32    RegOpenKeyExW(uz, c16 *, u32, u32, uz *);
W32 i32    RegCreateKeyExW(uz, c16 *, u32, c16 *, u32, u32, void *, uz *, u32 *);
W32 i32    RegSetValueExW(uz, c16 *, u32, u32, u8 *, u32);
W32 i32    RegDeleteKeyW(uz, c16 *);
W32 i32    RegCloseKey(uz);
W32 b32    ShellExecuteW(uz, c16 *, c16 *, c16 *, c16 *, i32);

static void *memset(void *s, int c, uz n)
{
    u8 *p = (u8 *)s;
    for (uz i = 0; i < n; i++) {
        p[i] = (u8)c;
    }
    return s;
}

static void *memcpy(void *dest, const void *src, uz n)
{
    u8 *d = (u8 *)dest;
    const u8 *s = (const u8 *)src;
    for (uz i = 0; i < n; i++) {
        d[i] = s[i];
    }
    return dest;
}

#define S(s) (s8){(u8 *)s, countof(s)-1}
typedef struct {
    u8 *s;
    iz  len;
} s8;

#define U(s) (s16){s, countof(s)-1}
typedef struct {
    c16 *s;
    iz   len;
} s16;

static s8 s8span(u8 *beg, u8 *end)
{
    s8 s = {};
    s.s = beg;
    s.len = end - beg;
    return s;
}

static b32 s8equals(s8 a, s8 b)
{
    if (a.len != b.len) {
        return 0;
    }
    for (iz i = 0; i < a.len; i++) {
        if (a.s[i] != b.s[i]) {
            return 0;
        }
    }
    return 1;
}

static void fatal(c16 *msg)
{
    MessageBoxW(0, msg, u"WinKit launcher", 0x10);
    ExitProcess(2);
}

typedef struct {
    u8 *beg;
    u8 *end;
} Arena;

static void outofmemory()
{
    fatal(u"Out of Memory");
}

static u8 *alloc(Arena *a, iz size, iz align, iz count)
{
    iz padding = (iz)a->end & (align - 1);
    if (count > (a->end - a->beg - padding)/size) {
        outofmemory();
    }
    u8 *r = a->end -= count*size + padding;
    return memset(r, 0, (uz)(count*size));
}

typedef struct {
    u8   *base;
    Arena perm;
    Arena scratch;
} Memory;

static Memory newmemory(iz cap)
{
    Memory r = {};
    r.base = VirtualAlloc(0, cap, MEM_RESERVE|MEM_COMMIT, PAGE_READWRITE);
    if (!r.base) return r;
    r.perm.beg = r.base;
    r.perm.end = r.base + cap/2;
    r.scratch.beg = r.base + cap/2;
    r.scratch.end = r.base + cap;
    return r;
}

static void freememory(Memory m)
{
    VirtualFree(m.base, 0, MEM_RELEASE);
}

static i32 truncsize(iz len)
{
    i32 max = 0x7fffffff;
    return len>max ? max : (i32)len;
}

static u32 tou32(i32 len)
{
    assert(len >= 0);
    return (u32)len;
}

typedef enum {
    INI_eof,
    INI_section,
    INI_key,
    INI_value
} IniType;

typedef struct {
    s8      name;
    IniType type;
} IniToken;

typedef struct {
    u8 *beg;
    u8 *end;
    b32 invalue;
} IniParser;

static b32 inidone(IniParser *p)
{
    return p->beg == p->end;
}

static u8 ininext(IniParser *p)
{
    return *p->beg++;
}

static b32 iniblank(u8 c)
{
    return c==' ' || c=='\t' || c=='\r';
}

static void iniskip(IniParser *p)
{
    for (; !inidone(p) && iniblank(*p->beg); p->beg++) {}
}

static u8 *initok(IniParser *p, u8 term)
{
    u8 *end = p->beg;
    while (!inidone(p)) {
        u8 c = ininext(p);
        if (c == term) {
            return end;
        } else if (c == '\n') {
            break;
        } else if (!iniblank(c)) {
            end = p->beg;
        }
    }
    return term=='\n' ? end : 0;
}

static b32 iniquoted(s8 s)
{
    return s.len>1 && s.s[0]=='"' && s.s[s.len-1]=='"';
}

// Parses like GetPrivateProfileString except sections cannot contain
// newlines. Invalid input lines are ignored. Comment lines begin with
// ';' following any whitespace. No trailing comments. Trims leading and
// trailing whitespace from sections, keys, and values. To preserve
// whitespace, values may be double-quoted. No quote escapes. Content on
// a line following a closing section ']' is ignored. Token encoding
// matches input encoding. An INI_value always follows an INI_key key.
static IniToken iniparse(IniParser *p)
{
    IniToken r = {};

    if (p->invalue) {
        p->invalue = 0;
        iniskip(p);
        u8 *beg = p->beg;
        u8 *end = initok(p, '\n');
        r.name = s8span(beg, end);
        r.type = INI_value;
        if (iniquoted(r.name)) {
            r.name.s++;
            r.name.len -= 2;
        }
        return r;
    }

    for (;;) {
        iniskip(p);
        if (inidone(p)) {
            return r;
        }

        u8 *end;
        u8 *beg = p->beg;
        switch (ininext(p)) {
        case ';':
            while (!inidone(p) && ininext(p)!='\n') {}
            break;

        case '[':
            iniskip(p);
            beg = p->beg;
            end = initok(p, ']');
            if (end) {
                // skip over anything else on the line
                while (!inidone(p) && ininext(p)!='\n') {}
                r.name = s8span(beg, end);
                r.type = INI_section;
                return r;
            }
            break;

        case '\n':
            break;

        default:
            end = initok(p, '=');
            if (end) {
                p->invalue = 1;
                r.name = s8span(beg, end);
                r.type = INI_key;
                return r;
            }
        }
    }
}

typedef enum {
    sym_null = 0,
    sym_winkit,
    sym_home,
    sym_title,
    sym_path_type,
    sym_inherit,
    sym_minimal,
    sym_strict,
    sym_font,        // 添加字体配置符号
    sym_font_size,   // 添加字体大小配置符号
} Symbol;

static Symbol intern(s8 s)
{
    static struct {
        s8     name;
        Symbol symbol;
    } symbols[] = {
        {S("winkit"), sym_winkit},
        {S("home"),      sym_home},
        {S("title"),     sym_title},
        {S("path type"), sym_path_type},
        {S("inherit"),   sym_inherit},
        {S("minimal"),   sym_minimal},
        {S("strict"),    sym_strict},
        {S("font"),      sym_font},       // 添加字体配置
        {S("font size"), sym_font_size},  // 添加字体大小配置
    };
    for (iz i = 0; i < countof(symbols); i++) {
        if (s8equals(symbols[i].name, s)) {
            return symbols[i].symbol;
        }
    }
    return sym_null;
}

static c16 *expandvalue(s8 value, Arena *perm, Arena scratch)
{
    assert(value.len < 0x7fffffff);

    // First temporarily convert to a null-terminated wide string
    i32 len = MultiByteToWideChar(CP_UTF8, 0, value.s, (i32)value.len, 0, 0);
    if (len == 0x7fffffff) outofmemory();
    s16 w = {};
    w.len = len + 1;  // append null terminator
    w.s = new(&scratch, c16, w.len);
    MultiByteToWideChar(CP_UTF8, 0, value.s, (i32)value.len, w.s, len);

    len = (i32)ExpandEnvironmentStringsW(w.s, 0, 0);
    if (len < 0) outofmemory();
    c16 *r = new(perm, c16, len);
    ExpandEnvironmentStringsW(w.s, r, tou32(len));
    return r;
}

// Expand to a full path via GetFullPathNameW.
static c16 *tofullpath(c16 *path, Arena *perm)
{
    i32 len = (i32)GetFullPathNameW(path, 0, 0, 0);
    if (len < 0) outofmemory();
    c16 *r = new(perm, c16, len);
    GetFullPathNameW(path, tou32(len), r, 0);
    return r;
}

static s8 loadfile(c16 *path, Arena *perm)
{
    s8 r = {};

    uz h = CreateFileW(
        path,
        GENERIC_READ,
        FILE_SHARE_ALL,
        0,
        OPEN_EXISTING,
        0,
        0
    );
    if (h == (uz)-1) {
        return r;
    }

    r.s = (u8 *)perm->beg;
    r.len = truncsize(perm->end - perm->beg);
    i32 len;
    b32 ok = ReadFile(h, r.s, (i32)r.len, &len, 0);
    CloseHandle(h);
    r.len = len;
    r.s = ok ? r.s : 0;
    perm->beg += r.len;
    return r;
}

typedef struct {
    c16   *home;
    c16   *title;
    c16   *font;        // 添加字体配置
    i32    font_size;   // 添加字体大小配置
    Symbol path_type;
    b32    ok;
} Config;

static Config newconfig()
{
    Config r = {};
    r.title = u"winkit";
    r.path_type = sym_inherit;
    r.font = u"Consolas";  // 默认字体
    r.font_size = 24;      // 默认字体大小
    return r;
}

// Read entries from winkit.ini. Expands environment variables, and
// if "home" is relative, converts it to an absolute path. Before the
// call, the working directory must be location of winkit.exe.
static Config loadconfig(Arena *perm, Arena scratch)
{
    Config conf = newconfig();

    s8 ini = loadfile(u"winkit.ini", &scratch);
    if (!ini.s) return conf;

    IniParser *p = new(&scratch, IniParser, 1);
    p->beg = ini.s;
    p->end = ini.s + ini.len;

    for (Symbol section = 0, key = 0;;) {
        IniToken t = iniparse(p);
        switch (t.type) {
        case INI_eof:
            conf.ok = 1;
            return conf;
        case INI_section:
            section = intern(t.name);
            break;
        case INI_key:
            key = intern(t.name);
            break;
        case INI_value:
            if (section == sym_winkit) {
                if (!conf.home && key==sym_home) {
                    Arena temp = scratch;
                    conf.home = expandvalue(t.name, &temp, *perm);
                    conf.home = tofullpath(conf.home, perm);
                } else if (key == sym_title) {
                    conf.title = expandvalue(t.name, perm, scratch);
                } else if (key == sym_path_type) {
                    conf.path_type = intern(t.name);
                } else if (key == sym_font) {
                    conf.font = expandvalue(t.name, perm, scratch);
                } else if (key == sym_font_size) {
                    // 解析字体大小
                    s8 value = t.name;
                    i32 size = 0;
                    for (iz i = 0; i < value.len; i++) {
                        if (value.s[i] >= '0' && value.s[i] <= '9') {
                            size = size * 10 + (value.s[i] - '0');
                        }
                    }
                    if (size > 0) {
                        conf.font_size = size;
                    }
                }
            }
            break;
        }
    }
}

typedef struct {
    c16 *buf;
    iz   cap;
    iz   len;
    b32  err;
} Buf16;

static Buf16 newbuf16(Arena *a, iz cap)
{
    Buf16 buf = {};
    buf.buf = new(a, c16, cap);
    buf.cap = cap;
    return buf;
}

static void buf16cat(Buf16 *buf, s16 s)
{
    iz avail = buf->cap - buf->len;
    iz count = s.len<avail ? s.len : avail;
    c16 *dst = buf->buf + buf->len;
    for (iz i = 0; i < count; i++) {
        dst[i] = s.s[i];
    }
    buf->len += count;
    buf->err |= count < s.len;
}

static void buf16c16(Buf16 *buf, c16 c)
{
    s16 s = {&c, 1};
    buf16cat(buf, s);
}

static void buf16moduledir(Buf16 *buf, Arena scratch)
{
    // GetFullPathNameW does not allow querying the output size, instead
    // indicating whether or not the buffer was large enough. So simply
    // offer the entire scratch buffer, then crop out the actual result.
    scratch.beg += -(uz)scratch.beg & (sizeof(c16) - 1);  // align
    i32 len = truncsize((scratch.end - scratch.beg)/2);
    s16 path = {};
    path.s = (c16 *)scratch.beg;
    path.len = (i32)GetModuleFileNameW(0, path.s, tou32(len));
    if (len == path.len) outofmemory();
    for (; path.len; path.len--) {
        switch (path.s[path.len-1]) {
        case  '/':
        case '\\': path.len--;
                   buf16cat(buf, path);
                   return;
        }
    }
}

static void buf16getenv(Buf16 *buf, c16 *key, Arena scratch)
{
    i32 len = (i32)GetEnvironmentVariableW(key, 0, 0);
    if (len < 0) outofmemory();
    s16 var = {};
    var.s = new(&scratch, c16, len);
    var.len = (i32)GetEnvironmentVariableW(key, var.s, tou32(len));
    buf16cat(buf, var);
}

static s16 getwindir(Arena *perm)
{
    i32 len = (i32)GetWindowsDirectoryW(0, 0);
    if (len < 0) outofmemory();
    s16 r = {};
    r.s   = new(perm, c16, len);
    r.len = (i32)GetWindowsDirectoryW(r.s, tou32(len));
    return r;
}

static void buf16minpath(Buf16 *buf, Arena scratch)
{
    s16 windir = getwindir(&scratch);

    buf16cat(buf, windir);
    buf16cat(buf, U(u"\\System32;"));

    buf16cat(buf, windir);
    buf16cat(buf, U(u";"));

    buf16cat(buf, windir);
    buf16cat(buf, U(u"\\System32\\Wbem"));

    // PowerShell directory first appears in Windows 7
    u16 version = (u16)GetVersion();
    version = (u16)(version>>8 | version<<8);
    if (version > 0x0600) {
        buf16cat(buf, U(u";"));
        buf16cat(buf, windir);
        buf16cat(buf, U(u"\\System32\\WindowsPowerShell\\v1.0"));
    }
}

static void toslashes(c16 *path)
{
    for (iz i = 0; i < path[i]; i++) {
        path[i] = path[i]=='\\' ? '/' : path[i];
    }
}

// 添加处理文件路径的函数
static void handle_file_argument(c16 *filepath, Arena *perm, Arena *scratch) {
    // 对于文件参数，我们不需要特殊处理，直接传递给shell即可
    // 文件路径会在CreateProcessW调用中作为参数传递给shell
}

// 添加处理目录路径的函数
static void handle_directory_argument(c16 *dirpath, Arena *perm, Arena *scratch) {
    // 设置CD_DIR环境变量，用于在shell启动后切换目录
    // 处理路径中的单引号同双引号
    c16 *clean_path = new(perm, c16, 512);
    i32 i = 0;
    
    // 跳过所有引号字符（单引号同双引号）
    for (i32 j = 0; dirpath[j] && i < 510; j++) {
        if (dirpath[j] != '"' && dirpath[j] != '\'') {
            clean_path[i++] = dirpath[j];
        }
    }
    clean_path[i] = '\0';
    
    // 设置CD_DIR环境变量，使用清理后的路径
    SetEnvironmentVariableW(u"CD_DIR", clean_path);  // ignore errors
}

static i32 winkit(c16 *file_arg)
{
    Memory mem = newmemory(1<<22);
    if (!mem.base) {
        fatal(u"Out of Memory on startup");
    }
    Arena *perm = &mem.perm;
    Arena  scratch = mem.scratch;

    // 设置控制台代码页为UTF-8 (65001)
    SetConsoleOutputCP(CP_UTF8);
    SetConsoleCP(CP_UTF8);

    // First load the module directory into the fresh buffer, and use it
    // for a few different operations.
    Buf16 path = newbuf16(perm, MAX_ENVVAR);
    buf16moduledir(&path, scratch);
    Buf16 moduledir = path;  // to truncate back to the module dir

    buf16c16(&path, 0);  // null terminator
    
    // 处理参数
    b32 is_directory = 0;
    if (file_arg) {
        // 如果是通过--cd=指定的参数，直接视为目录
        if (g_is_cd_param) {
            is_directory = 1;
            handle_directory_argument(file_arg, perm, &scratch);
        } else {
            // 检查路径是文件还是目录
            u32 attrs = GetFileAttributesW(file_arg);
            if (attrs != 0xFFFFFFFF && (attrs & 0x10)) {  // DIRECTORY attribute
                // 是目录，设置CD_DIR环境变量
                is_directory = 1;
                handle_directory_argument(file_arg, perm, &scratch);
            } else {
                // 是文件，处理文件参数
                handle_file_argument(file_arg, perm, &scratch);
            }
        }
    }
    
    // 始终设置WINKIT_HOME为默认值或配置文件中的值
    SetEnvironmentVariableW(u"WINKIT_HOME", path.buf);  // ignore errors
    
    #ifdef VERSION
    #define LSTR(s) XSTR(s)
    #define XSTR(s) u ## # s
    SetEnvironmentVariableW(u"WINKIT", LSTR(VERSION));  // ignore errors
    #endif
    
    // Maybe set HOME from winkit.ini
    Config conf = newconfig();
    if (SetCurrentDirectoryW(path.buf)) {
        conf = loadconfig(perm, scratch);
        if (conf.home) {
            toslashes(conf.home);
            SetEnvironmentVariableW(u"HOME", conf.home);  // ignore errors
        }
    }
    
    // 设置APPDATA环境变量为$WINKIT_HOME/share
    path = moduledir;
    buf16cat(&path, U(u"\\share"));
    buf16c16(&path, 0);  // null terminator
    SetEnvironmentVariableW(u"APPDATA", path.buf);  // ignore errors

    // 设置控制台字体
    CONSOLE_FONT_INFOEX fontInfo = {0};
    fontInfo.cbSize = sizeof(CONSOLE_FONT_INFOEX);
    fontInfo.nFont = 0;
    fontInfo.dwFontSize.X = 0;
    fontInfo.dwFontSize.Y = conf.font_size;
    fontInfo.FontFamily = 54;
    fontInfo.FontWeight = 400;
    
    // 设置字体
    for (int i = 0; i < 32 && conf.font[i]; i++) {
        fontInfo.FaceName[i] = conf.font[i];
    }
    
    uz consoleHandle = GetStdHandle(-11); // STD_OUTPUT_HANDLE
    SetCurrentConsoleFontEx(consoleHandle, 0, &fontInfo);

    // Continue building PATH
    path = moduledir;
    buf16cat(&path, U(u"\\bin"));
    switch (conf.path_type) {
    case sym_inherit:
        buf16cat(&path, U(u";"));
        buf16getenv(&path, u"PATH", scratch);
        break;
    case sym_minimal:
        buf16cat(&path, U(u";"));
        buf16minpath(&path, scratch);
        break;
    case sym_strict:
        break;
    default:
        fatal(u"winkit.ini: 'path type' must be inherit|minimal|strict");
    }
    buf16c16(&path, 0);  // null terminator
    if (path.err || !SetEnvironmentVariableW(u"PATH", path.buf)) {
        fatal(u"Failed to configure $PATH");
    }

    // Set the console title as late as possible, but not after starting
    // the shell because .profile might change it.
    if (conf.title) {
        SetConsoleTitleW(conf.title);  // ignore errors
    }

    // 将控制台窗口居中显示
    uz hConsoleWindow = GetConsoleWindow();
    if (hConsoleWindow) {
        // 获取屏幕尺寸
        i32 screenWidth = GetSystemMetrics(SM_CXSCREEN);
        i32 screenHeight = GetSystemMetrics(SM_CYSCREEN);
        
        // 获取窗口当前的矩形区域，这样可以得到实际的窗口大小
        RECT windowRect;
        if (GetWindowRect(hConsoleWindow, &windowRect)) {
            // 计算窗口的实际像素大小
            i32 windowWidth = windowRect.right - windowRect.left;
            i32 windowHeight = windowRect.bottom - windowRect.top;
            
            // 计算居中位置
            i32 centerX = (screenWidth - windowWidth) / 2;
            i32 centerY = (screenHeight - windowHeight) / 2;
            
            // 确保位置不会为负数
            if (centerX < 0) centerX = 0;
            if (centerY < 0) centerY = 0;
            
            // 移动窗口到居中位置，保持当前大小及Z顺序
            SetWindowPos(hConsoleWindow, 0, centerX, centerY, 0, 0, SWP_NOSIZE | SWP_NOZORDER);
        }
    }

    path = moduledir;
    buf16cat(&path, U(u"\\bin\\busybox.exe"));
    buf16c16(&path, 0);  // null terminator

    // Start a BusyBox login shell
    Si si = {};
    si.cb = sizeof(si);
    Pi pi;
    
    // 构造命令行参数
    c16 *cmdline;
    c16 base_cmdline[] = u"sh -l";  // 基本命令行
    
    if (file_arg && !is_directory) {
        // 是文件参数，作为参数传递给shell
        // 分配足够的空间来容纳基本命令同文件路径
        cmdline = new(perm, c16, 512);  // 分配足够空间
        i32 i = 0;
        // 复制基本命令
        for (; base_cmdline[i]; i++) {
            cmdline[i] = base_cmdline[i];
        }
        // 添加空格
        cmdline[i++] = ' ';
        // 复制文件路径
        for (iz j = 0; file_arg[j] && i < 511; j++, i++) {
            cmdline[i] = file_arg[j];
        }
        cmdline[i] = 0;  // null terminator
    } else {
        // 没有参数或参数是目录，使用基本命令行
        cmdline = base_cmdline;
    }

    // 先启动登录shell
    if (!CreateProcessW(path.buf, cmdline, 0, 0, 1, 0, 0, 0, &si, &pi)) {
        fatal(u"Failed to launch a login shell");
    }

    // =============================================================
    // 由于视窗系统的限制，我们不能在CreateProcessW中设置工作目录
    // 登录后，shell会自动加载profile脚本并检查是否有CD_DIR环境变量
    // 注意：CD_DIR环境变量经已在handle_directory_argument函数中设置
    // 切换目录的逻辑会在shell启动后由profile脚本处理
    // =============================================================

    // Wait for shell to exit
    freememory(mem);
    i32 ret;
    WaitForSingleObject(pi.process, -1);
    GetExitCodeProcess(pi.process, &ret);
    return ret;
}

// 命令行参数解析函数
static c16 *find_next_arg(c16 *cmdline, c16 **end) {
    // 跳过前导空格
    while (*cmdline == ' ' || *cmdline == '\t') {
        cmdline++;
    }
    
    if (*cmdline == 0) {
        *end = cmdline;
        return 0;
    }
    
    c16 *start = cmdline;
    c16 *arg_end = cmdline;
    
    // 处理引号
    if (*cmdline == '"') {
        cmdline++;
        start = cmdline;
        // 找到匹配的引号
        while (*cmdline && *cmdline != '"') {
            cmdline++;
        }
        arg_end = cmdline;
        if (*cmdline == '"') {
            cmdline++;
        }
    } else {
        // 找到参数的结束位置（空格或字符串结束）
        while (*cmdline && *cmdline != ' ' && *cmdline != '\t') {
            cmdline++;
        }
        arg_end = cmdline;
    }
    
    *end = cmdline;
    return start;
}

[[gnu::stdcall]]
// 添加到注册表的函数
static b32 add_to_registry() {
    // 获取当前可执行文件的路径
    c16 exe_path[MAX_PATH];
    u32 len = GetModuleFileNameW(0, exe_path, MAX_PATH);
    if (len == 0 || len >= MAX_PATH) {
        fatal(u"Failed to get executable path");
    }
    
    // 为目录添加右键菜单
    { 
        c16 shell_path[] = u"Directory\\Background\\shell\\WinKit";
        c16 cmd_path[] = u"Directory\\Background\\shell\\WinKit\\command";
        c16 menu_text[] = u"在此目录启动WinKit";
        uz key;
        u32 disp;
        i32 result = RegCreateKeyExW((uz)(-2147483648), shell_path, 0, 0, 0, 
                                     KEY_SET_VALUE | KEY_WOW64_64KEY, 0, &key, &disp);
        if (result == 0) {
            // 设置右键菜单显示文本
            RegSetValueExW(key, 0, 0, REG_SZ, (u8 *)menu_text, sizeof(menu_text));
            RegCloseKey(key);
        } else {
            return 0;
        }
        
        // 创建command子键
        result = RegCreateKeyExW((uz)(-2147483648), cmd_path, 0, 0, 0, 
                                 KEY_SET_VALUE | KEY_WOW64_64KEY, 0, &key, &disp);
        if (result == 0) {
            // 构造命令："exe_path" "--cd='%V'"
            c16 command[MAX_PATH + 30];
            i32 i = 0;
            command[i++] = '"';
            for (u32 j = 0; j < len; j++) {
                command[i++] = exe_path[j];
            }
            command[i++] = '"';
            command[i++] = ' ';
            command[i++] = '"';
            command[i++] = '-';
            command[i++] = '-';
            command[i++] = 'c';
            command[i++] = 'd';
            command[i++] = '=';
            command[i++] = '\'';
            command[i++] = '%';
            command[i++] = 'V';
            command[i++] = '\'';
            command[i++] = '"';
            command[i] = 0;
            
            // 设置默认值，使用REG_EXPAND_SZ类型以支持环境变量展开
            RegSetValueExW(key, 0, 0, REG_EXPAND_SZ, (u8 *)command, i * sizeof(c16));
            RegCloseKey(key);
        } else {
            return 0;
        }
    }
    

    
    return 1;
}

// 从注册表删除的函数
static b32 remove_from_registry() {
    // 删除目录右键菜单的子键command，然后再删除主键
    c16 reg_cmd_path[] = u"Directory\\Background\\shell\\WinKit\\command";
    RegDeleteKeyW((uz)(-2147483648), reg_cmd_path); // 忽略错误，因为可能不存在
    
    c16 reg_path[] = u"Directory\\Background\\shell\\WinKit";
    i32 result = RegDeleteKeyW((uz)(-2147483648), reg_path);
    // 只在不是"键不存在"的错误时返回失败
    // ERROR_FILE_NOT_FOUND = 2
    if (result != 0 && result != 2) {
        return 0;
    }
    
    return 1;
}

void mainCRTStartup()
{
    // 获取命令行
    c16 *cmdline = GetCommandLineW();
    
    // 跳过程序名（第一个参数）
    c16 *end;
    c16 *program_name = find_next_arg(cmdline, &end);
    
    // 查找第二个参数（如果存在）
    c16 *file_arg = find_next_arg(end, &end);
    
    // 特殊处理：检查是否有--cd=参数
    // 这个修改解决了引号包含目录参数导致程序将其视为文件执行的问题
    // 通过识别--cd=格式的参数，我们可以明确知道这是一个目录路径
    g_is_cd_param = 0;
    if (file_arg) {
        // 检查是否以"--cd="开头
        c16 cd_prefix[] = u"--cd=";
        b32 is_cd_param = 1;
        for (i32 i = 0; cd_prefix[i]; i++) {
            if (file_arg[i] != cd_prefix[i]) {
                is_cd_param = 0;
                break;
            }
        }
        
        if (is_cd_param) {
            // 将--cd=后面的内容提取出来作为目录路径
            // 跳过"--cd="这6个字符
            file_arg += 6;
            g_is_cd_param = 1;
        }
    }
    
    // 检查是否有特殊参数
    if (file_arg && file_arg[0] == '-') {
        c16 *arg_start = file_arg + 1;
        
        // 检查是否是 -install 参数
        b32 is_install = 1;
        c16 install_str[] = u"install";
        for (i32 i = 0; install_str[i]; i++) {
            if (arg_start[i] != install_str[i]) {
                is_install = 0;
                break;
            }
        }
        
        if (is_install && !arg_start[7]) {
            // 安装右键菜单
            if (add_to_registry()) {
                MessageBoxW(0, u"WinKit右键菜单已成功安装！", u"WinKit", 0);
            } else {
                MessageBoxW(0, u"WinKit右键菜单安装失败！请以管理员身份运行。", u"WinKit", 0x10);
            }
            ExitProcess(0);
        }
        
        // 检查是否是 -uninstall 参数
        b32 is_uninstall = 1;
        c16 uninstall_str[] = u"uninstall";
        for (i32 i = 0; uninstall_str[i]; i++) {
            if (arg_start[i] != uninstall_str[i]) {
                is_uninstall = 0;
                break;
            }
        }
        
        if (is_uninstall && !arg_start[9]) {
            // 卸载右键菜单
            if (remove_from_registry()) {
                MessageBoxW(0, u"WinKit右键菜单已成功卸载！", u"WinKit", 0);
            } else {
                MessageBoxW(0, u"WinKit右键菜单卸载失败！请以管理员身份运行。", u"WinKit", 0x10);
            }
            ExitProcess(0);
        }
        
        // 如果不是有效的特殊参数，将其视为普通参数
        // （保持兼容性）
    }
    
    // 将参数传递给winkit函数
    // 这可能是文件路径或目录路径
    i32 r = winkit(file_arg);
    ExitProcess(r);
}
