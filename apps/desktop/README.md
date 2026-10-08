# WPF desktop

.NET 8 WPF, six dark views and real authenticated service bindings. Build with
`dotnet build apps/desktop/Neexitlag.Desktop.csproj` from the repository root.
See [root README](../../README.md) for Windows installation and loopback demo flags.
The UI contains no privileged networking code. Service absence, version errors,
invalid frames and timeouts produce an unavailable/disconnected state.
WPF cross-build and shared C# binding smoke pass on Linux; native WPF runtime
and named-pipe integration still require Windows acceptance.
