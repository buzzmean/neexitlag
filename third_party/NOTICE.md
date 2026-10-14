# Referenced implementations and runtime dependency

Flowseal/zapret-discord-youtube and bol-van/zapret were studied at revisions recorded in docs/FLOWSEAL_ZAPRET_ANALYSIS.md. Independent Go implementation; no upstream fake files, lists, batch scripts, executable or packet-processing code is vendored. Their MIT notices are retained under licenses/.

Existing WinDivert 2.2.2 is dynamically loaded, unmodified, not relicensed by this project. Select LGPLv3 for distribution; include COPYING.LESSER, COPYING and the corresponding source link with binary artifacts. Users may replace the DLL with a compatible build. Sources and build instructions: https://github.com/basil00/WinDivert/tree/v2.2.2 . Do not distribute modified library/driver without fulfilling corresponding source obligations. Kernel-driver signing constraints still apply.
