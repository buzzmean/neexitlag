# Referenced implementations and runtime dependency

Flowseal/zapret-discord-youtube and bol-van/zapret were studied at revisions recorded in docs/FLOWSEAL_ZAPRET_ANALYSIS.md. Independent Go implementation; no upstream fake files, full list snapshots, batch scripts, executable or packet-processing code is vendored. Their MIT notices are retained under licenses/.

Existing WinDivert 2.2.2 is dynamically loaded, unmodified, not relicensed by this project. Select LGPLv3 for distribution; include COPYING.LESSER, COPYING and the corresponding source link with binary artifacts. Users may replace the DLL with a compatible build. Sources and build instructions: https://github.com/basil00/WinDivert/tree/v2.2.2 . Do not distribute modified library/driver without fulfilling corresponding source obligations. Kernel-driver signing constraints still apply.

DIRECT+ v2 independently implements bounded fake/split/disorder concepts and manually curates small Discord/YouTube domain subsets after review of Flowseal eb16c1419057e36b8a36fbe7c959483450e6ac8b and zapret c4d0b1990129dce2b28095277173ee0d7702d304. Existing Flowseal/bol-van MIT copyright and license notices in licenses/ accompany distributions. No upstream fake templates, winws binary or C packet implementation is bundled. See docs/DIRECT_PLUS_V2.md.
