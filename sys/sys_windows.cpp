#include "sys_windows.h"
#include <algorithm>
#include <atomic>
#include <condition_variable>
#include <cstdlib>
#include <mutex>
#include <thread>

const char* getClipboardText()
{
    if(!OpenClipboard(nullptr))
        return "";

    HANDLE clip = GetClipboardData(CF_TEXT);
    if(clip == nullptr)
        return "";

    auto s = (char*)GlobalLock(clip);
    if(s == nullptr)
        return "";

    auto len = strlen(s);
    auto ret = (char*)std::malloc(len + 1);
    std::copy(s, s + len, ret);
    ret[len] = 0;

    GlobalUnlock(clip);
    CloseClipboard();

    return ret;
}

namespace input
{
    HHOOK hook;
    std::mutex mtx;
    input_t value{};
    std::condition_variable cv;
    bool ready = false;
    std::atomic<int> refCount{};
} // namespace input

LRESULT globalHook(int n, WPARAM w, LPARAM l)
{
    if(w != WM_KEYDOWN && w != WM_KEYUP)
        return CallNextHookEx(nullptr, n, w, l);

    auto& hs = *(PKBDLLHOOKSTRUCT)l;
    DWORD code = hs.vkCode;

    {
        std::lock_guard lk{input::mtx};
        input::value = {.key = uint16_t(code), .flag = uint64_t(w)};
        input::ready = true;
    }
    input::cv.notify_one();

    return CallNextHookEx(nullptr, n, w, l);
}

void setGlobalHook()
{
    input::hook = SetWindowsHookEx(WH_KEYBOARD_LL, globalHook, nullptr, 0);
    input::refCount++;
    MSG msg;
    while(input::refCount > 0 && !GetMessage(&msg, nullptr, 0, 0))
    {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
}

void unhook()
{
    UnhookWindowsHookEx(input::hook);
    input::refCount--;
}

input_t getInput()
{
    std::unique_lock lk{input::mtx};
    input::cv.wait(lk, [] { return input::ready || input::refCount == 0; });
    input::ready = false;
    return input::value;
}