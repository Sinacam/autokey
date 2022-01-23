#ifndef SYS_WINDOWS_HPP_INCLUDED
#define SYS_WINDOWS_HPP_INCLUDED

#include <windows.h>
#include <string.h>
#include <stdatomic.h>
#include <stdint.h>
#include <stdlib.h>

char *GetClipboardText()
{
    if (!OpenClipboard(NULL))
        return "";

    HANDLE clip = GetClipboardData(CF_TEXT);
    if (clip == NULL)
        return "";

    const char *s = GlobalLock(clip);
    if (s == NULL)
        return "";

    char *ret = malloc(strlen(s) + 1);
    strcpy(ret, s);

    GlobalUnlock(clip);
    CloseClipboard();

    return ret;
}

static uint64_t keystates[4];

LRESULT globalHook(int n, WPARAM w, LPARAM l)
{
    if (w != WM_KEYDOWN && w != WM_KEYUP)
        return CallNextHookEx(NULL, n, w, l);

    PKBDLLHOOKSTRUCT phs = (PKBDLLHOOKSTRUCT)l;
    DWORD code = phs->vkCode;
    if (w == WM_KEYDOWN)
        atomic_fetch_or(&keystates[code / 64], (uint64_t)1 << (code % 64));
    else
        atomic_fetch_and(&keystates[code / 64], ~((uint64_t)1 << (code % 64)));

    return CallNextHookEx(NULL, n, w, l);
}

HHOOK SetGlobalHook()
{
    return SetWindowsHookEx(WH_KEYBOARD_LL, globalHook, NULL, 0);
}

#endif