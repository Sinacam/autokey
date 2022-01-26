#ifndef SYS_WINDOWS_HPP_INCLUDED
#define SYS_WINDOWS_HPP_INCLUDED

#include <stdint.h>
#include <windows.h>

#ifdef __cplusplus
extern "C"
{
#endif

    const char* getClipboardText();
    LRESULT globalHook(int n, WPARAM w, LPARAM l);
    void setGlobalHook();
    void unhook();

    struct input_t
    {
        uint16_t key;
        uint64_t flag;
    };

    struct input_t getInput();

#ifdef __cplusplus
}
#endif

#endif
